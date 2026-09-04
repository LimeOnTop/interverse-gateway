package api

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type AuthAPI struct {
	authClient    *clients.AuthClient
	adminUsername string
}

func NewAuthAPI(authClient *clients.AuthClient, adminUsername string) *AuthAPI {
	return &AuthAPI{
		authClient:    authClient,
		adminUsername: strings.TrimSpace(adminUsername),
	}
}

func (a *AuthAPI) isAdminLogin(login string) bool {
	return a.adminUsername != "" && login == a.adminUsername
}

func (a *AuthAPI) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if a.isAdminLogin(strings.TrimSpace(req.Email)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this login is reserved"})
		return
	}

	resp, err := a.authClient.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (a *AuthAPI) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	login := strings.TrimSpace(req.Email)
	if login == "" {
		login = strings.TrimSpace(req.Username)
	}
	password := req.Password
	if login == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and password are required"})
		return
	}

	// Regular users must provide a valid email; admin may use ADMIN_USERNAME as-is.
	if !a.isAdminLogin(login) {
		if _, err := mail.ParseAddress(login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
			return
		}
	}

	resp, err := a.authClient.Login(c.Request.Context(), login, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	resp, err := a.authClient.Logout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp != nil && !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.authClient.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) GetMe(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.authClient.GetUser(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	if resp.User == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                       resp.User.Id,
			"email":                    resp.User.Email,
			"name":                     resp.User.Name,
			"role":                     resp.User.Role,
			"subscription_plan":        resp.User.SubscriptionPlan,
			"subscription_active":      resp.User.SubscriptionActive,
			"subscription_expires_at":  resp.User.SubscriptionExpiresAt,
		},
	})
}

func (a *AuthAPI) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	resp, err := a.authClient.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.authClient.UpdateUser(c.Request.Context(), userID, req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	resp, err := a.authClient.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp != nil && !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}
