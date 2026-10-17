package api

import (
	"context"
	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

type limitsClient struct{ usecase.InterviewGateway }

func (limitsClient) GetTrainingStats(_ context.Context, user, plan string) (*pb.GetTrainingStatsResponse, error) {
	if user != "0" || plan != "free" {
		panic("unexpected identity")
	}
	return &pb.GetTrainingStatsResponse{QuotaLimit: 5}, nil
}
func TestPublicTrainingLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	a := NewDashboardAPI(limitsClient{}, nil, nil)
	r.GET("/limits", a.PublicTrainingLimits)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/limits", nil))
	if w.Code != 200 || w.Body.String() != `{"free_trainings_total":5}` {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
