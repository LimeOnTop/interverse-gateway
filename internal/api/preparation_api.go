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
	"io"
	"net/http"
	"strconv"
)

type PreparationGateway interface {
	VacancyPreparation(context.Context, *pb.VacancyPreparationRequest) (*pb.VacancyPreparationResponse, error)
}
type PreparationAPI struct {
	client PreparationGateway
	auth   usecase.AuthGateway
}

func NewPreparationAPI(c PreparationGateway, a usecase.AuthGateway) *PreparationAPI {
	return &PreparationAPI{c, a}
}
func (a *PreparationAPI) Update(c *gin.Context)  { a.execute(c, "update") }
func (a *PreparationAPI) Archive(c *gin.Context) { a.execute(c, "archive") }
func (a *PreparationAPI) List(c *gin.Context)    { a.execute(c, "list") }
func (a *PreparationAPI) Create(c *gin.Context)  { a.execute(c, "create") }
func (a *PreparationAPI) Prepare(c *gin.Context) { a.execute(c, "prepare_step") }
func (a *PreparationAPI) execute(c *gin.Context, action string) {
	u, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	r := &pb.VacancyPreparationRequest{UserId: u.ID, Action: action, PreparationId: c.Param("id")}
	if action != "list" {
		r.SubscriptionPlan = subscriptionPlan(c.Request.Context(), a.auth, u.ID)
	}
	if action == "create" || action == "update" || action == "archive" {
		raw, e := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 16000))
		if e != nil || !json.Valid(raw) {
			c.JSON(400, gin.H{"error": "Некорректные данные подготовки"})
			return
		}
		r.PayloadJson = string(raw)
	}
	if action == "prepare_step" {
		i, e := strconv.Atoi(c.Param("step"))
		if e != nil || i < 0 || i > 29 {
			c.JSON(400, gin.H{"error": "Этап не найден"})
			return
		}
		r.StepIndex = int32(i)
		var body struct {
			RequestID string `json:"request_id"`
		}
		if c.ShouldBindJSON(&body) != nil {
			c.JSON(400, gin.H{"error": "Некорректный запрос"})
			return
		}
		r.RequestId = body.RequestID
	}
	resp, e := a.client.VacancyPreparation(c.Request.Context(), r)
	if e != nil {
		code := 500
		message := "Не удалось загрузить подготовку. Попробуйте ещё раз"
		switch status.Code(e) {
		case codes.InvalidArgument:
			code = 400
			message = status.Convert(e).Message()
		case codes.NotFound:
			code = 404
			message = "Подготовка не найдена"
		case codes.PermissionDenied:
			code = 403
			message = status.Convert(e).Message()
		case codes.FailedPrecondition, codes.Aborted:
			code = 409
			message = status.Convert(e).Message()
		case codes.ResourceExhausted:
			c.JSON(429, gin.H{"error": status.Convert(e).Message(), "code": "training_limit_exceeded"})
			return
		}
		c.JSON(code, gin.H{"error": message})
		return
	}
	c.Data(200, "application/json; charset=utf-8", []byte(resp.PayloadJson))
}
