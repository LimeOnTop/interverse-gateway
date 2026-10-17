package api

import (
	"context"
	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type reviewTrainingGateway interface {
	ReviewTraining(context.Context, *pb.ReviewTrainingRequest) (*pb.StartSessionResponse, error)
}

func (a *ReportAPI) SetReviewClient(client reviewTrainingGateway) { a.reviewClient = client }
func (a *ReportAPI) StartReviewTraining(c *gin.Context) {
	resp, user, ok := a.loadOwnedReport(c)
	if !ok {
		return
	}
	if !a.fullReportAccess(c, user) {
		apperr.Public(c, 403, "Повторение по ошибкам доступно в Pro")
		return
	}
	var body struct {
		RequestID string `json:"request_id"`
	}
	if c.ShouldBindJSON(&body) != nil {
		apperr.Public(c, 400, "Некорректные параметры повторения")
		return
	}
	mapped := mapReportResponse(protoReportToMap(resp.GetReport()), true)
	points, _ := mapped["weak_points"].([]reportWeakPoint)
	ids := []string{}
	seen := map[string]bool{}
	for _, point := range points {
		if point.StepID != "" && !seen[point.StepID] {
			ids = append(ids, point.StepID)
			seen[point.StepID] = true
		}
	}
	if len(ids) == 0 {
		apperr.Public(c, 409, "В отчёте нет вопросов с ошибками для повторения")
		return
	}
	if a.reviewClient == nil {
		apperr.Public(c, 503, "Повторение временно недоступно")
		return
	}
	result, err := a.reviewClient.ReviewTraining(c.Request.Context(), &pb.ReviewTrainingRequest{UserId: user.ID, SourceInterviewId: resp.GetReport().GetInterviewId(), SourceReportId: resp.GetReport().GetId(), StepIds: ids, RequestId: body.RequestID, SubscriptionPlan: planPaid})
	if err != nil {
		switch status.Code(err) {
		case codes.ResourceExhausted:
			c.JSON(429, gin.H{"error": "Лимит тренировок исчерпан", "code": "training_limit_exceeded"})
		case codes.InvalidArgument:
			apperr.Public(c, 400, status.Convert(err).Message())
		case codes.NotFound:
			apperr.Public(c, 404, "Исходная тренировка не найдена")
		case codes.FailedPrecondition:
			apperr.Public(c, 409, status.Convert(err).Message())
		default:
			apperr.Internal(c, err)
		}
		return
	}
	c.JSON(http.StatusCreated, result)
}
