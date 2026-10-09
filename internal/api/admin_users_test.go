package api

import (
	"context"
	"encoding/json"
	pb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type directoryStub struct {
	usecase.AuthGateway
	calls   int
	request *pb.ListUsersRequest
}

func (s *directoryStub) ListUsers(_ context.Context, r *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.calls++
	s.request = r
	return &pb.ListUsersResponse{Response: &pb.Response{Success: true}, Users: []*pb.AdminUser{{Id: "1", Name: "Name", Email: "name@fixture.invalid"}}, Total: 1, Page: r.GetPage(), Limit: r.GetLimit()}, nil
}
func TestAdminUserDirectoryAccessAndValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, role, params string
		status             int
		calls              int
	}{
		{"anonymous", "", "", 401, 0}, {"user", "user", "", 403, 0}, {"admin", "admin", "", 200, 1},
		{"bad period", "admin", "?period=bad", 400, 0}, {"bad page", "admin", "?page=-1", 400, 0}, {"bad limit", "admin", "?limit=999", 400, 0},
		{"long query", "admin", "?query=" + strings.Repeat("a", 101), 400, 0}, {"day", "admin", "?period=day&query=alice", 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &directoryStub{}
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.role != "" {
					c.Set("auth_user", middleware.User{ID: "1", Role: tc.role})
				}
				c.Next()
			})
			r.GET("/admin/users", middleware.AdminRequired(), NewAdminAPI(nil, stub, nil).ListUsers)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/users"+tc.params, nil))
			if w.Code != tc.status || stub.calls != tc.calls {
				t.Fatalf("status=%d calls=%d body=%s", w.Code, stub.calls, w.Body)
			}
			if tc.status == 200 {
				var result map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if w.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("directory can be cached")
				}
				if tc.name == "day" && (stub.request.GetRegisteredSinceUnix() == 0 || stub.request.GetQuery() != "alice") {
					t.Fatal("filters not forwarded")
				}
			}
		})
	}
}
