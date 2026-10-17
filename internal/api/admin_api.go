package api

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	authpb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	paymentpb "github.com/LimeOnTop/interverse-contracts/payment/gen"
	questionpb "github.com/LimeOnTop/interverse-contracts/question/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AdminAPI struct {
	questionClient usecase.QuestionGateway
	authClient     usecase.AuthGateway
	paymentClient  usecase.PaymentGateway
}

func NewAdminAPI(
	questionClient usecase.QuestionGateway,
	authClient usecase.AuthGateway,
	paymentClient usecase.PaymentGateway,
) *AdminAPI {
	return &AdminAPI{
		questionClient: questionClient,
		authClient:     authClient,
		paymentClient:  paymentClient,
	}
}

func (a *AdminAPI) Stats(c *gin.Context) {
	ctx := c.Request.Context()

	totalResp, err := a.questionClient.GetQuestions(ctx, 1, 1)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	type levelStat struct {
		Difficulty string `json:"difficulty"`
		Total      int32  `json:"total"`
	}

	levels := []string{"junior", "middle", "senior", "pending_moderation"}
	byDifficulty := make([]levelStat, 0, len(levels))
	for _, level := range levels {
		category := ""
		if level == "pending_moderation" {
			category = "contribution"
		}
		resp, err := a.questionClient.GetQuestionsByTechnology(ctx, "", level, category, 1, 1)
		if err != nil {
			apperr.Internal(c, err)
			return
		}
		total := int32(0)
		if resp.Pagination != nil {
			total = resp.Pagination.Total
		}
		byDifficulty = append(byDifficulty, levelStat{Difficulty: level, Total: total})
	}

	total := int32(0)
	if totalResp.Pagination != nil {
		total = totalResp.Pagination.Total
	}

	c.JSON(http.StatusOK, gin.H{
		"total_questions": total,
		"by_difficulty":   byDifficulty,
	})
}

// metricsLocation is Europe/Moscow. Moscow has had a fixed UTC+3 offset since 2014, so a fixed zone
// avoids depending on tzdata in the container image.
var metricsLocation = time.FixedZone("Europe/Moscow", 3*60*60)

type periodCounts struct {
	Today int64 `json:"today"`
	Week  int64 `json:"week"`
	Month int64 `json:"month"`
	Total int64 `json:"total"`
}

// metricPeriods returns calendar boundaries in Moscow time: today from 00:00,
// the last 7 days including today, and the current month from the 1st.
func metricPeriods(now time.Time) (dayStart, weekStart, monthStart time.Time) {
	local := now.In(metricsLocation)
	dayStart = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, metricsLocation)
	weekStart = dayStart.AddDate(0, 0, -6)
	monthStart = time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, metricsLocation)
	return dayStart, weekStart, monthStart
}

// Metrics returns registrations and confirmed payments (any plan, status paid) per period.
func (a *AdminAPI) Metrics(c *gin.Context) {
	ctx := c.Request.Context()
	dayStart, weekStart, monthStart := metricPeriods(time.Now())

	var (
		wg             sync.WaitGroup
		regResp        *authpb.GetRegistrationStatsResponse
		payResp        *paymentpb.GetPaymentStatsResponse
		regErr, payErr error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		regResp, regErr = a.authClient.GetRegistrationStats(ctx, &authpb.PeriodBoundaries{
			DayStartUnix:   dayStart.Unix(),
			WeekStartUnix:  weekStart.Unix(),
			MonthStartUnix: monthStart.Unix(),
		})
	}()
	go func() {
		defer wg.Done()
		payResp, payErr = a.paymentClient.GetPaymentStats(ctx, &paymentpb.PeriodBoundaries{
			DayStartUnix:   dayStart.Unix(),
			WeekStartUnix:  weekStart.Unix(),
			MonthStartUnix: monthStart.Unix(),
		})
	}()
	wg.Wait()

	if regErr != nil {
		apperr.Internal(c, regErr)
		return
	}
	if regResp.Response != nil && !regResp.Response.Success {
		apperr.Upstream(c, http.StatusBadGateway, regResp.Response.Error)
		return
	}
	if payErr != nil {
		apperr.Internal(c, payErr)
		return
	}
	if payResp.Response != nil && !payResp.Response.Success {
		apperr.Upstream(c, http.StatusBadGateway, payResp.Response.Error)
		return
	}

	registrations := periodCounts{}
	if r := regResp.GetRegistrations(); r != nil {
		registrations = periodCounts{Today: r.Today, Week: r.Week, Month: r.Month, Total: r.Total}
	}
	payments := periodCounts{}
	if p := payResp.GetPaid(); p != nil {
		payments = periodCounts{Today: p.Today, Week: p.Week, Month: p.Month, Total: p.Total}
	}

	c.JSON(http.StatusOK, gin.H{
		"timezone":      "Europe/Moscow",
		"day_start":     dayStart.Format(time.RFC3339),
		"week_start":    weekStart.Format(time.RFC3339),
		"month_start":   monthStart.Format(time.RFC3339),
		"registrations": registrations,
		"payments":      payments,
	})
}

func (a *AdminAPI) ListQuestions(c *gin.Context) {
	page, limit := parsePagination(c)
	if limit > 100 {
		limit = 100
	}

	technology := strings.TrimSpace(c.Query("technology"))
	difficulty := strings.TrimSpace(c.Query("difficulty"))
	category := strings.TrimSpace(c.Query("category"))

	// No filters → full bank (all directions and levels).
	if technology == "" && difficulty == "" && category == "" {
		resp, err := a.questionClient.GetQuestions(c.Request.Context(), page, limit)
		if err != nil {
			apperr.Internal(c, err)
			return
		}
		if resp.Response != nil && !resp.Response.Success {
			apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	resp, err := a.questionClient.GetQuestionsByTechnology(c.Request.Context(), technology, difficulty, category, page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AdminAPI) ListModerationQuestions(c *gin.Context) {
	page, limit := parsePagination(c)
	if limit > 100 {
		limit = 100
	}

	resp, err := a.questionClient.GetQuestionsByTechnology(
		c.Request.Context(),
		c.Query("technology"),
		"pending_moderation",
		"contribution",
		page,
		limit,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AdminAPI) ApproveModerationQuestion(c *gin.Context) {
	var req questionpb.ApproveContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}
	// Author and bonus duration never come from the HTTP request.
	req.QuestionId = c.Param("id")
	client, ok := a.questionClient.(interface {
		ApproveContribution(context.Context, *questionpb.ApproveContributionRequest) (*questionpb.ApproveContributionResponse, error)
	})
	if !ok {
		apperr.Public(c, http.StatusServiceUnavailable, "Одобрение временно недоступно")
		return
	}
	resp, err := client.ApproveContribution(c.Request.Context(), &req)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if !resp.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadRequest, resp.GetResponse().GetError())
		return
	}
	message := "Вопрос добавлен в банк. Автору начислено 30 дней Pro."
	if !resp.RewardGranted {
		message = "Вопрос добавлен в банк. Начисление 30 дней Pro поставлено в очередь и повторится автоматически."
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "question": resp.Question, "reward_granted": resp.RewardGranted, "reward_days": 30})
}

func (a *AdminAPI) RejectModerationQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	existing, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if existing.Question == nil {
		apperr.Public(c, http.StatusNotFound, "question not found")
		return
	}

	resp, err := a.questionClient.DeleteQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp != nil && !resp.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вопрос отклонён и удалён"})
}

// GrafanaAuth is used by nginx auth_request before proxying /grafana/.
// Returns 200 for admin JWTs and exposes X-Grafana-User for Grafana auth proxy.
func (a *AdminAPI) GrafanaAuth(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	name := strings.TrimSpace(user.Email)
	if name == "" {
		name = strings.TrimSpace(user.ID)
	}
	if name == "" {
		name = "admin"
	}
	c.Header("X-Grafana-User", name)
	c.Status(http.StatusOK)
}
