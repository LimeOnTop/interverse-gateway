package api

import (
	"context"
	authpb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	reportpb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

type reviewReportStub struct {
	usecase.ReportGateway
	owner string
}

func (s *reviewReportStub) GetReport(context.Context, string) (*reportpb.GetReportResponse, error) {
	return &reportpb.GetReportResponse{Response: &reportpb.Response{Success: true}, Report: &reportpb.Report{Id: "source-report", UserId: s.owner, InterviewId: "source-interview", Notes: `{"weak_points":[{"step_id":"wrong-step","item_type":"question"}],"answer_reviews":[{"step_id":"correct-step","is_correct":true}]}`}}, nil
}

type reviewPlanStub struct {
	usecase.AuthGateway
	paid bool
}

func (s *reviewPlanStub) GetUser(context.Context, string) (*authpb.GetUserResponse, error) {
	return &authpb.GetUserResponse{Response: &authpb.Response{Success: true}, User: &authpb.User{Id: "1", SubscriptionActive: s.paid, SubscriptionPlan: map[bool]string{true: "paid", false: "free"}[s.paid]}}, nil
}

type reviewRPCStub struct {
	calls int
	r     *pb.ReviewTrainingRequest
}

func (s *reviewRPCStub) ReviewTraining(_ context.Context, r *pb.ReviewTrainingRequest) (*pb.StartSessionResponse, error) {
	s.calls++
	s.r = r
	return &pb.StartSessionResponse{Response: &pb.Response{Success: true}}, nil
}
func TestReviewTrainingOwnershipPlanAndSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, owner string
		paid        bool
		status      int
	}{{"Pro owner", "1", true, 201}, {"foreign report", "2", true, 404}, {"Basic", "1", false, 403}} {
		t.Run(tc.name, func(t *testing.T) {
			rpc := &reviewRPCStub{}
			handler := NewReportAPI(&reviewReportStub{owner: tc.owner}, &reviewPlanStub{paid: tc.paid})
			handler.SetReviewClient(rpc)
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("auth_user", middleware.User{ID: "1", Role: "user"}); c.Next() })
			router.POST("/reports/:id/practice", handler.StartReviewTraining)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("POST", "/reports/source-report/practice", strings.NewReader(`{"request_id":"request-key","user_id":"999","step_ids":["forged-step"]}`)))
			if w.Code != tc.status {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}
			if tc.status == 201 {
				if rpc.calls != 1 || rpc.r.UserId != "1" || len(rpc.r.StepIds) != 1 || rpc.r.StepIds[0] != "wrong-step" {
					t.Fatalf("wrong payload %+v", rpc.r)
				}
			} else if rpc.calls != 0 {
				t.Fatal("unauthorized creation")
			}
		})
	}
}
