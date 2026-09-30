package middleware

import (
	"net/http"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/authjwt"
	"github.com/gin-gonic/gin"
)

const userContextKey = "auth_user"

// AccessTokenCookie is read by AuthRequired for browser navigations (e.g. Grafana proxy).
const AccessTokenCookie = "iv_access_token"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func extractAccessToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token != "" {
			return token
		}
	}
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil {
		return strings.TrimSpace(cookie)
	}
	return ""
}

// AuthRequired validates JWT locally in the gateway (HMAC + optional Redis blacklist).
// Accepts Authorization: Bearer <token> or cookie iv_access_token.
func AuthRequired(validator *authjwt.Validator, blacklist *authjwt.AccessBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAccessToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		claims, err := validator.ParseAccess(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		revoked, err := blacklist.IsRevoked(c.Request.Context(), claims.JTI)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		if revoked {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		c.Set(userContextKey, User{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		})
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (User, bool) {
	val, ok := c.Get(userContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return User{}, false
	}

	user, ok := val.(User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return User{}, false
	}

	return user, true
}
