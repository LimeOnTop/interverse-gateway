package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"strings"
)

// Counters are shared by all gateway replicas and expire after a fixed window.
func AuthRateLimit(cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request too large"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var req struct {
			Email    string `json:"email"`
			Username string `json:"username"`
		}
		_ = json.Unmarshal(body, &req)
		account := strings.ToLower(strings.TrimSpace(req.Email))
		if account == "" {
			account = strings.ToLower(strings.TrimSpace(req.Username))
		}
		keys := []string{fmt.Sprintf("auth:limit:ip:%x", sha256.Sum256([]byte(c.ClientIP())))}
		limits := []int{60}
		if account != "" {
			keys = append(keys, fmt.Sprintf("auth:limit:account:%x", sha256.Sum256([]byte(account))))
			limits = append(limits, 10)
		}
		if cache == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication unavailable"})
			return
		}
		for i, key := range keys {
			allowed, err := cache.Eval(c.Request.Context(), `local n=redis.call('INCR',KEYS[1]);if n==1 then redis.call('EXPIRE',KEYS[1],ARGV[2]) end;if n>tonumber(ARGV[1]) then return 0 end;return 1`, []string{key}, limits[i], 900).Int()
			if err != nil {
				log.Printf("auth rate limiter Redis failure: %v", err)
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication unavailable"})
				return
			}
			if allowed != 1 {
				c.Header("Retry-After", "900")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много попыток. Попробуйте через 15 минут."})
				return
			}
		}
		c.Next()
	}
}
