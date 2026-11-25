package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type operationsStub struct {
	usecase.PaymentGateway
	calls int
}

func (s *operationsStub) ListOperations(context.Context, int64) (*pb.ListOperationsResponse, error) {
	s.calls++
	return &pb.ListOperationsResponse{}, nil
}

func TestOperationsAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role  string
		want  int
		calls int
	}{
		{"", http.StatusUnauthorized, 0}, {"user", http.StatusForbidden, 0}, {"admin", http.StatusOK, 1},
	} {
		t.Run(tc.role, func(t *testing.T) {
			stub := &operationsStub{}
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tc.role != "" {
					c.Set("auth_user", middleware.User{ID: "1", Role: tc.role})
				}
				c.Next()
			})
			router.GET("/admin/operations", middleware.AdminRequired(), NewAdminAPI(nil, nil, stub).ListOperations)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin/operations", nil))
			if response.Code != tc.want || stub.calls != tc.calls {
				t.Fatalf("status=%d calls=%d", response.Code, stub.calls)
			}
			if tc.role == "admin" && response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("financial journal must not be cached")
			}
		})
	}
}
