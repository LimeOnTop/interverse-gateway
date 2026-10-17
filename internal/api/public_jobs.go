package api

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	pb "github.com/LimeOnTop/interverse-contracts/vacancy/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var publicJobQueries = map[string]string{"react": "React frontend", "frontend": "Frontend", "python": "Python developer", "java": "Java developer"}

type publicFeed struct {
	Track     string        `json:"track"`
	Items     []*pb.Vacancy `json:"items"`
	UpdatedAt string        `json:"updated_at"`
	Partial   bool          `json:"partial"`
	Stale     bool          `json:"stale"`
}
type PublicJobsAPI struct {
	cache       redis.UniversalClient
	vacancies   usecase.VacancyGateway
	templateDir string
	locks       sync.Map
	background  sync.Map
}

func NewPublicJobsAPI(cache redis.UniversalClient, vacancies usecase.VacancyGateway, templateDir string) *PublicJobsAPI {
	return &PublicJobsAPI{cache: cache, vacancies: vacancies, templateDir: templateDir}
}
func (a *PublicJobsAPI) allowed(c *gin.Context) (string, bool) {
	track := c.Param("track")
	if _, ok := publicJobQueries[track]; !ok {
		c.JSON(404, gin.H{"error": "Направление не найдено"})
		return "", false
	}
	ctx := c.Request.Context()
	key := "public:jobs-rate:" + c.ClientIP()
	count, err := a.cache.Incr(ctx, key).Result()
	if err != nil {
		c.JSON(503, gin.H{"error": "Подборка временно недоступна"})
		return "", false
	}
	if count == 1 {
		a.cache.Expire(ctx, key, time.Minute)
	}
	if count > 120 {
		c.Header("Retry-After", "60")
		c.JSON(429, gin.H{"error": "Слишком много запросов"})
		return "", false
	}
	return track, true
}
func (a *PublicJobsAPI) load(ctx context.Context, track string) (publicFeed, error) {
	read := func(key string) (publicFeed, bool) {
		raw, err := a.cache.Get(ctx, key).Result()
		var feed publicFeed
		ok := err == nil && json.Unmarshal([]byte(raw), &feed) == nil
		return feed, ok
	}
	freshKey := "public:jobs:v2:" + track
	if f, ok := read(freshKey); ok {
		return f, nil
	}
	value, _ := a.locks.LoadOrStore(track, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	if f, ok := read(freshKey); ok {
		return f, nil
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	resp, err := a.vacancies.SearchVacancies(fetchCtx, publicJobQueries[track], "", 0, 50, nil)
	if err != nil || resp == nil || resp.Response != nil && !resp.Response.Success || len(resp.Items) == 0 && len(resp.Errors) > 0 {
		if f, ok := read(freshKey + ":stale"); ok {
			f.Stale = true
			return f, nil
		}
		if err != nil {
			return publicFeed{}, err
		}
		return publicFeed{}, redis.Nil
	}
	feed := publicFeed{Track: track, Items: make([]*pb.Vacancy, 0, len(resp.Items)), UpdatedAt: time.Now().UTC().Format(time.RFC3339), Partial: len(resp.Errors) > 0}
	for _, v := range resp.Items {
		if v == nil || strings.TrimSpace(v.Title) == "" || !publicJobURL(v.Url) || !publicJobRelevant(track, v) {
			continue
		}
		feed.Items = append(feed.Items, v)
	}
	raw, _ := json.Marshal(feed)
	if err := a.cache.Set(ctx, freshKey, raw, 10*time.Minute).Err(); err != nil {
		return publicFeed{}, err
	}
	_ = a.cache.Set(ctx, freshKey+":stale", raw, 24*time.Hour).Err()
	return feed, nil
}

var publicJobTerms = map[string]*regexp.Regexp{
	"react":    regexp.MustCompile(`(?i)\breact\b`),
	"frontend": regexp.MustCompile(`(?i)front[ -]?end|фронт[ -]?енд|фронт[ -]?энд|\breact\b|\bvue\b|\bangular\b`),
	"python":   regexp.MustCompile(`(?i)\bpython\b|питон`),
	"java":     regexp.MustCompile(`(?i)\bjava\b`),
}
var publicNonDeveloper = regexp.MustCompile(`(?i)\bqa\b|тестировщик|рекрутер|\brecruiter\b`)

func publicJobRelevant(track string, v *pb.Vacancy) bool {
	if publicNonDeveloper.MatchString(v.Title) {
		return false
	}
	text := v.Title + " " + v.Snippet + " " + strings.Join(v.Skills, " ")
	return publicJobTerms[track].MatchString(text)
}
func publicJobURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil
}
func (a *PublicJobsAPI) Get(c *gin.Context) {
	track, ok := a.allowed(c)
	if !ok {
		return
	}
	feed, err := a.load(c.Request.Context(), track)
	if err != nil {
		log.Printf("public jobs %s: %v", track, err)
		c.JSON(503, gin.H{"error": "Не удалось загрузить вакансии. Попробуйте позже"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(200, feed)
}

var jobCards = template.Must(template.New("cards").Parse(`{{range .Items}}<article class="public-job"><div class="public-job-meta">{{.SourceLabel}} · {{.Area}}</div><h2><a href="{{.Url}}" target="_blank" rel="noopener noreferrer">{{.Title}}</a></h2><p class="public-job-company">{{.CompanyName}}</p><p>{{.Snippet}}</p><div class="public-job-meta">{{.Experience}} · {{.Schedule}}</div><a class="public-job-source" href="{{.Url}}" target="_blank" rel="noopener noreferrer">Открыть вакансию у источника →</a></article>{{else}}<p>В этой подборке пока нет вакансий. Проверьте соседние направления или вернитесь позже.</p>{{end}}`))

func (a *PublicJobsAPI) Page(c *gin.Context) {
	track, ok := a.allowed(c)
	if !ok {
		return
	}
	shell, err := os.ReadFile(filepath.Join(a.templateDir, track+".html"))
	if os.IsNotExist(err) {
		shell, err = os.ReadFile(filepath.Join(a.templateDir, "jobs", track, "index.html"))
	}
	if err != nil {
		c.Status(503)
		return
	}
	// Render from cache immediately. Provider latency belongs to the separate
	// JSON request, never to delivery of the page's HTML or its navigation.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Millisecond)
	feed := publicFeed{Track: track, Items: []*pb.Vacancy{}}
	fresh := false
	for i, key := range []string{"public:jobs:v2:" + track, "public:jobs:v2:" + track + ":stale"} {
		raw, cacheErr := a.cache.Get(ctx, key).Result()
		if cacheErr == nil && json.Unmarshal([]byte(raw), &feed) == nil {
			fresh = i == 0
			feed.Stale = !fresh
			break
		}
	}
	cancel()
	if !fresh {
		if _, running := a.background.LoadOrStore(track, true); !running {
			go func() {
				defer a.background.Delete(track)
				ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
				defer cancel()
				if _, err := a.load(ctx, track); err != nil {
					log.Printf("public jobs background %s: %v", track, err)
				}
			}()
		}
	}
	var cards bytes.Buffer
	view := feed
	if len(view.Items) > 20 {
		view.Items = view.Items[:20]
	}
	if err = jobCards.Execute(&cards, view); err != nil {
		c.Status(503)
		return
	}
	if feed.UpdatedAt == "" {
		cards.Reset()
		cards.WriteString(`<p role="status">Загружаем актуальные вакансии…</p>`)
	}
	raw, _ := json.Marshal(feed)
	html := strings.Replace(string(shell), "<!--PUBLIC_JOB_CARDS-->", cards.String(), 1)
	html = strings.Replace(html, "<!--PUBLIC_JOB_DATA-->", `<script id="public-jobs-data" type="application/json">`+string(raw)+`</script>`, 1)
	c.Header("Cache-Control", "public, max-age=60")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
