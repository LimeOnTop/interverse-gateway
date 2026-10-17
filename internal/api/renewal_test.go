package api

import (
	"context"
	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type renewalStub struct {
	usecase.PaymentGateway
	id      string
	enabled bool
	calls   int
}

func (s *renewalStub) SetRenewal(_ context.Context, id string, enabled bool) (*pb.RenewalResponse, error) {
	s.id = id
	s.enabled = enabled
	s.calls++
	return &pb.RenewalResponse{Response: &pb.Response{Success: true}, Enabled: enabled}, nil
}
func TestRenewalUsesAuthenticatedOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role, body  string
		want, calls int
	}{{"", `{"enabled":false}`, 401, 0}, {"admin", `{"enabled":false}`, 400, 0}, {"user", `{}`, 400, 0}, {"user", `{"enabled":false,"user_id":"999"}`, 200, 1}, {"user", `{"enabled":true}`, 200, 1}} {
		stub := &renewalStub{}
		r := gin.New()
		r.Use(func(c *gin.Context) {
			if tc.role != "" {
				c.Set("auth_user", middleware.User{ID: "123", Role: tc.role})
			}
			c.Next()
		})
		r.PUT("/renewal", NewPaymentAPI(stub).SetRenewal)
		req := httptest.NewRequest(http.MethodPut, "/renewal", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want || stub.calls != tc.calls {
			t.Fatalf("status=%d calls=%d body=%s", w.Code, stub.calls, w.Body)
		}
		if stub.calls > 0 && stub.id != "123" {
			t.Fatal("client can change another user's renewal")
		}
	}
}
func TestCheckoutRequiresRenewalConsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &renewalStub{}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("auth_user", middleware.User{ID: "123", Role: "user"}) })
	r.POST("/payments", NewPaymentAPI(stub).CreatePayment)
	for _, body := range []string{`{"plan":"paid_1m"}`, `{"plan":"paid_1m","auto_renew_consent":false}`, `{`} {
		req := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("checkout accepted without consent: %d", w.Code)
		}
	}
}
