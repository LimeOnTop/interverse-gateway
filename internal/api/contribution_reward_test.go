package api

import (
	"context"
	pb "github.com/LimeOnTop/interverse-contracts/question/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type approvalStub struct {
	usecase.QuestionGateway
	calls   int
	request *pb.ApproveContributionRequest
}

func (s *approvalStub) ApproveContribution(_ context.Context, r *pb.ApproveContributionRequest) (*pb.ApproveContributionResponse, error) {
	s.calls++
	s.request = r
	return &pb.ApproveContributionResponse{Response: &pb.Response{Success: true}, Question: &pb.Question{Id: r.QuestionId}, RewardGranted: true}, nil
}
func TestContributionApprovalAdminOnlyAndOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role   string
		status int
	}{{"", 401}, {"user", 403}, {"admin", 200}} {
		t.Run(tc.role, func(t *testing.T) {
			stub := &approvalStub{}
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.role != "" {
					c.Set("auth_user", middleware.User{ID: "44", Role: tc.role})
				}
				c.Next()
			})
			r.POST("/admin/moderation/questions/:id/approve", middleware.AdminRequired(), NewAdminAPI(stub, nil, nil).ApproveModerationQuestion)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/moderation/questions/path-id/approve", strings.NewReader(`{"question_id":"forged-id","user_id":"999","reward_days":3000,"text":"question","options":[{"text":"correct","is_correct":true},{"text":"wrong"}]}`)))
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.role == "admin" {
				if stub.calls != 1 || stub.request.QuestionId != "path-id" || len(stub.request.Options) != 2 || !stub.request.Options[0].IsCorrect {
					t.Fatalf("payload %+v", stub.request)
				}
			} else if stub.calls != 0 {
				t.Fatal("unauthorized RPC call")
			}
		})
	}
}
