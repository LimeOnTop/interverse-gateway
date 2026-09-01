package middleware

import (
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

const userContextKey = "auth_user"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func AuthRequired(authClient *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		resp, err := authClient.ValidateToken(c.Request.Context(), authHeader[7:])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if resp.Response != nil && !resp.Response.Success {
			c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Response.Error})
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set(userContextKey, User{
			ID:    resp.Id,
			Email: resp.Email,
			Role:  resp.Role,
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
