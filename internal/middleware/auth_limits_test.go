package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthRateLimitAccountAndFailClosed(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("requires isolated Redis")
	}
	cache := redis.NewClient(&redis.Options{Addr: addr, DB: 1})
	defer cache.Close()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", AuthRateLimit(cache), func(c *gin.Context) { c.Status(204) })
	account := fmt.Sprintf("limit-%d@example.invalid", time.Now().UnixNano())
	for i := 0; i < 11; i++ {
		r := httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"`+account+`"}`))
		r.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", i+1)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		want := 204
		if i == 10 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("attempt %d got %d want %d", i, w.Code, want)
		}
	}
	closed := gin.New()
	closed.POST("/login", AuthRateLimit(nil), func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	closed.ServeHTTP(w, httptest.NewRequest("POST", "/login", strings.NewReader(`{}`)))
	if w.Code != 503 {
		t.Fatal("missing Redis bypassed")
	}
}
