package api

import (
	"context"
	"encoding/json"
	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

type practiceGateway interface {
	PracticeState(context.Context, *pb.PracticeStateRequest) (*pb.PracticeStateResponse, error)
}
type PracticeAPI struct {
	client  practiceGateway
	auth    usecase.AuthGateway
	reports usecase.ReportGateway
}

func NewPracticeAPI(c practiceGateway, a usecase.AuthGateway, r usecase.ReportGateway) *PracticeAPI {
	return &PracticeAPI{c, a, r}
}
func (a *PracticeAPI) Restore(c *gin.Context)   { a.run(c, "restore") }
func (a *PracticeAPI) RetryMail(c *gin.Context) { a.run(c, "retry_email") }
func (a *PracticeAPI) Check(c *gin.Context)     { a.run(c, "check") }
func (a *PracticeAPI) List(c *gin.Context)      { a.run(c, "list") }
func (a *PracticeAPI) Save(c *gin.Context)      { a.run(c, "save") }
func (a *PracticeAPI) Dismiss(c *gin.Context)   { a.run(c, "dismiss") }
func (a *PracticeAPI) Start(c *gin.Context)     { a.run(c, "start") }
func (a *PracticeAPI) News(c *gin.Context)      { a.run(c, "news") }
func (a *PracticeAPI) ReadNews(c *gin.Context)  { a.run(c, "read_news") }
func (a *PracticeAPI) run(c *gin.Context, action string) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	body := map[string]any{}
	if action != "list" && action != "news" {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16000)
		if c.ShouldBindJSON(&body) != nil {
			c.JSON(400, gin.H{"error": "Некорректный запрос"})
			return
		}
	}
	if action == "save" && body["kind"] == "review" {
		id, _ := body["target_id"].(string)
		resp, e := a.reports.GetReport(c.Request.Context(), id)
		if e != nil || resp.GetReport() == nil || resp.GetReport().GetUserId() != user.ID {
			c.JSON(404, gin.H{"error": "Отчёт не найден"})
			return
		}
		mapped := mapReportResponse(protoReportToMap(resp.GetReport()), true)
		points, _ := mapped["weak_points"].([]reportWeakPoint)
		ids := []string{}
		seen := map[string]bool{}
		for _, p := range points {
			if p.StepID != "" && !seen[p.StepID] {
				ids = append(ids, p.StepID)
				seen[p.StepID] = true
			}
		}
		if len(ids) == 0 {
			c.JSON(409, gin.H{"error": "В отчёте нет вопросов с ошибками для повторения"})
			return
		}
		body["payload"] = map[string]any{"source_interview_id": resp.GetReport().GetInterviewId(), "step_ids": ids}
	}
	raw, _ := json.Marshal(body)
	plan := ""
	if action == "start" {
		plan = subscriptionPlan(c.Request.Context(), a.auth, user.ID)
	}
	resp, e := a.client.PracticeState(c.Request.Context(), &pb.PracticeStateRequest{UserId: user.ID, Action: action, PayloadJson: string(raw), SubscriptionPlan: plan})
	if e != nil {
		code := 500
		msg := "Не удалось сохранить изменения. Попробуйте ещё раз"
		switch status.Code(e) {
		case codes.InvalidArgument:
			code = 400
			msg = status.Convert(e).Message()
		case codes.NotFound:
			code = 404
			msg = status.Convert(e).Message()
		case codes.Unavailable:
			code = 503
			msg = status.Convert(e).Message()
		case codes.Unauthenticated:
			code = 401
		case codes.PermissionDenied:
			code = 403
			msg = status.Convert(e).Message()
		case codes.Aborted, codes.FailedPrecondition:
			code = 409
			msg = status.Convert(e).Message()
		case codes.ResourceExhausted:
			code = 429
			msg = status.Convert(e).Message()
			if strings.Contains(msg, "training_limit_exceeded") {
				c.JSON(code, gin.H{"error": "Лимит тренировок исчерпан", "code": "training_limit_exceeded"})
				return
			}
		}
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.Data(200, "application/json; charset=utf-8", []byte(resp.PayloadJson))
}
