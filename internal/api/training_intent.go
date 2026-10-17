package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const intentTTL = 24 * time.Hour

var intentIDPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

var landingTechnologies = map[string][]string{
	"frontend":     {"React", "Vue.js", "Angular", "TypeScript", "JavaScript", "HTML5", "CSS3", "Sass", "Less", "Webpack", "Vite", "Next.js", "Nuxt.js", "Svelte", "Tailwind CSS"},
	"backend":      {"Node.js", "Python", "Go", "Java", "C#", "PHP", "Ruby", "Rust", "Express.js", "Django", "Flask", "Gin", "Spring Boot", "Laravel", "Rails", "Actix"},
	"devops":       {"Docker", "Kubernetes", "AWS", "Azure", "GCP", "Terraform", "Ansible", "Jenkins", "GitLab CI", "GitHub Actions", "Prometheus", "Grafana", "ELK Stack"},
	"qa":           {"Selenium", "Cypress", "Playwright", "Jest", "Mocha", "Chai", "TestNG", "JUnit", "Postman", "Newman", "Appium", "Robot Framework", "Cucumber"},
	"data_science": {"Python", "R", "SQL", "Pandas", "NumPy", "Scikit-learn", "TensorFlow", "PyTorch", "Jupyter", "Tableau", "Power BI", "Apache Spark", "Hadoop"},
}

type trainingIntent struct {
	Specialization string   `json:"specialization"`
	Level          string   `json:"level"`
	Technologies   []string `json:"technologies"`
	UserID         string   `json:"user_id,omitempty"`
	InterviewID    string   `json:"interview_id,omitempty"`
}

type TrainingIntentAPI struct {
	cache      redis.UniversalClient
	interviews usecase.InterviewGateway
	auth       usecase.AuthGateway
}

func NewTrainingIntentAPI(cache redis.UniversalClient, interviews usecase.InterviewGateway, auth usecase.AuthGateway) *TrainingIntentAPI {
	return &TrainingIntentAPI{cache: cache, interviews: interviews, auth: auth}
}
func (a *TrainingIntentAPI) Save(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2048)
	var draft trainingIntent
	if c.ShouldBindJSON(&draft) != nil || !strings.Contains("|backend|frontend|devops|qa|data_science|", "|"+draft.Specialization+"|") || !strings.Contains("|intern|junior|middle|senior|lead|", "|"+draft.Level+"|") || len(draft.Technologies) != 1 || len(strings.TrimSpace(draft.Technologies[0])) == 0 || len(draft.Technologies[0]) > 80 {
		c.JSON(400, gin.H{"error": "Выберите направление, технологию и уровень"})
		return
	}
	validTechnology := false
	for _, technology := range landingTechnologies[draft.Specialization] {
		if technology == draft.Technologies[0] {
			validTechnology = true
			break
		}
	}
	if !validTechnology {
		c.JSON(400, gin.H{"error": "Выберите технологию выбранного направления"})
		return
	}
	// Never accept ownership or an interview identifier from a public request.
	draft.UserID, draft.InterviewID = "", ""
	ctx := c.Request.Context()
	rateKey := "landing:intent-rate:" + c.ClientIP()
	count, err := a.cache.Incr(ctx, rateKey).Result()
	if err != nil {
		c.JSON(503, gin.H{"error": "Не удалось сохранить параметры. Попробуйте ещё раз"})
		return
	}
	if count == 1 {
		a.cache.Expire(ctx, rateKey, time.Hour)
	}
	if count > 30 {
		c.JSON(429, gin.H{"error": "Слишком много попыток. Попробуйте позже"})
		return
	}
	token := make([]byte, 32)
	if _, err = rand.Read(token); err != nil {
		c.Status(503)
		return
	}
	id := hex.EncodeToString(token)
	data, _ := json.Marshal(draft)
	if err = a.cache.Set(ctx, "landing:intent:"+id, data, intentTTL).Err(); err != nil {
		c.JSON(503, gin.H{"error": "Не удалось сохранить параметры. Попробуйте ещё раз"})
		return
	}
	c.JSON(201, gin.H{"id": id, "expires_in": 86400})
}
func (a *TrainingIntentAPI) Resume(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if !intentIDPattern.MatchString(id) {
		c.JSON(404, gin.H{"error": "Параметры тренировки не найдены"})
		return
	}
	ctx := c.Request.Context()
	key := "landing:intent:" + id
	lockValue := make([]byte, 16)
	if _, err := rand.Read(lockValue); err != nil {
		c.Status(503)
		return
	}
	lockToken := hex.EncodeToString(lockValue)
	locked, err := a.cache.SetNX(ctx, key+":lock", lockToken, 2*time.Minute).Result()
	if err != nil {
		c.Status(503)
		return
	}
	if !locked {
		c.JSON(409, gin.H{"error": "Тренировка уже запускается. Попробуйте через несколько секунд"})
		return
	}
	defer a.cache.Eval(ctx, `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) end return 0`, []string{key + ":lock"}, lockToken)
	data, err := a.cache.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.JSON(410, gin.H{"error": "Время хранения параметров истекло. Выберите тренировку ещё раз"})
		return
	}
	var draft trainingIntent
	if err != nil || json.Unmarshal(data, &draft) != nil {
		c.Status(503)
		return
	}
	if draft.UserID != "" && draft.UserID != user.ID {
		c.JSON(404, gin.H{"error": "Параметры тренировки не найдены"})
		return
	}
	// Bind before creating. Retrying a started flow reuses its interview.
	draft.UserID = user.ID
	persist := func() bool {
		data, _ := json.Marshal(draft)
		if a.cache.Set(ctx, key, data, intentTTL).Err() != nil {
			c.Status(503)
			return false
		}
		return true
	}
	if !persist() {
		return
	}
	if draft.InterviewID == "" {
		result, err := a.interviews.CreateInterview(ctx, user.ID, "Тренировка: "+draft.Technologies[0]+" · "+draft.Level, "", "", draft.Level, draft.Specialization, subscriptionPlan(ctx, a.auth, user.ID), draft.Technologies)
		if err != nil {
			c.JSON(502, gin.H{"error": "Не удалось создать тренировку. Попробуйте ещё раз"})
			return
		}
		if result.GetResponse() != nil && !result.GetResponse().GetSuccess() {
			status := 400
			if strings.Contains(result.GetResponse().GetError(), "training_limit_exceeded") {
				status = 429
			}
			c.JSON(status, gin.H{"error": result.GetResponse().GetError()})
			return
		}
		if result.GetInterview().GetId() == "" {
			c.Status(502)
			return
		}
		draft.InterviewID = result.GetInterview().GetId()
		if !persist() {
			return
		}
	}
	result, err := a.interviews.StartSession(ctx, draft.InterviewID, user.ID)
	if err != nil {
		c.JSON(502, gin.H{"error": "Не удалось запустить тренировку. Попробуйте ещё раз"})
		return
	}
	if result.GetResponse() != nil && !result.GetResponse().GetSuccess() {
		c.JSON(400, gin.H{"error": result.GetResponse().GetError(), "interview_id": draft.InterviewID})
		return
	}
	c.JSON(200, result)
}
