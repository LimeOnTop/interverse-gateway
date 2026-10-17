package middleware

import (
	"context"
	"fmt"
	"github.com/LimeOnTop/interverse-gateway/internal/authjwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAdminSessionRevokedAfterVersionChange(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("requires isolated Redis")
	}
	cache := redis.NewClient(&redis.Options{Addr: addr, DB: 1})
	defer cache.Close()
	uid := time.Now().UnixNano()
	key := fmt.Sprintf("auth:version:%d", uid)
	defer cache.Del(context.Background(), key)
	claims := authjwt.Claims{UserID: uid, Role: "admin", TokenType: "access", RegisteredClaims: jwt.RegisteredClaims{ID: fmt.Sprintf("fixture-%d", uid), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("fixture"))
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin", AuthRequired(authjwt.NewValidator("fixture"), authjwt.NewAccessBlacklist(cache)), AdminRequired(), func(c *gin.Context) { c.Status(204) })
	request := func() int {
		r := httptest.NewRequest("GET", "/admin", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w.Code
	}
	if request() != 204 {
		t.Fatal("initial admin denied")
	}
	cache.Set(context.Background(), key, 1, time.Minute)
	if request() != 401 {
		t.Fatal("stale admin retained access")
	}
	closed := gin.New()
	closed.GET("/admin", AuthRequired(authjwt.NewValidator("fixture"), authjwt.NewAccessBlacklist(nil)), func(c *gin.Context) { c.Status(204) })
	r := httptest.NewRequest("GET", "/admin", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	closed.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("missing Redis bypassed")
	}
}
