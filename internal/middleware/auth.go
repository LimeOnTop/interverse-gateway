package middleware

import (
	"net/http"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/authjwt"
	"github.com/gin-gonic/gin"
)

const userContextKey = "auth_user"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthRequired validates JWT locally in the gateway (HMAC + optional Redis blacklist).
func AuthRequired(validator *authjwt.Validator, blacklist *authjwt.AccessBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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
