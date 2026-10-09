package api

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthAPI struct {
	authClient    usecase.AuthGateway
	adminUsername string
}

func NewAuthAPI(authClient usecase.AuthGateway, adminUsername string) *AuthAPI {
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
		apperr.Bind(c, err)
		return
	}

	if a.isAdminLogin(strings.TrimSpace(req.Email)) {
		apperr.Public(c, http.StatusBadRequest, "this login is reserved")
		return
	}

	resp, err := a.authClient.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	if resp.VerificationRequired {
		c.JSON(http.StatusAccepted, gin.H{
			"verification_required": true,
			"email":                 req.Email,
		})
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
		apperr.Bind(c, err)
		return
	}

	login := strings.TrimSpace(req.Email)
	if login == "" {
		login = strings.TrimSpace(req.Username)
	}
	password := req.Password
	if login == "" || password == "" {
		apperr.Public(c, http.StatusBadRequest, "login and password are required")
		return
	}

	// Regular users must provide a valid email; admin may use ADMIN_USERNAME as-is.
	if !a.isAdminLogin(login) {
		if _, err := mail.ParseAddress(login); err != nil {
			apperr.Public(c, http.StatusBadRequest, "invalid email")
			return
		}
	}

	resp, err := a.authClient.Login(c.Request.Context(), login, password)
	if err != nil {
		apperr.Public(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		if resp.VerificationRequired {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                 "email not verified",
				"verification_required": true,
				"email":                 login,
			})
			return
		}
		apperr.Upstream(c, http.StatusUnauthorized, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) VerifyEmail(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required,len=6,numeric"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Public(c, http.StatusBadRequest, "invalid email or code")
		return
	}

	resp, err := a.authClient.VerifyEmail(c.Request.Context(), strings.TrimSpace(req.Email), req.Code, req.Password)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) ResendEmailVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Public(c, http.StatusBadRequest, "invalid email")
		return
	}

	resp, err := a.authClient.SendEmailVerification(c.Request.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		if resp.RetryAfterSeconds > 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "code was sent recently, try again later",
				"retry_after_seconds": resp.RetryAfterSeconds,
			})
			return
		}
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{"retry_after_seconds": resp.RetryAfterSeconds})
}

func (a *AuthAPI) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		apperr.Public(c, http.StatusBadRequest, "Authorization header required")
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	resp, err := a.authClient.Logout(c.Request.Context(), token)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp != nil && !resp.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.authClient.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		apperr.Public(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusUnauthorized, resp.Response.Error)
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
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	if resp.User == nil {
		apperr.Public(c, http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                      resp.User.Id,
			"email":                   resp.User.Email,
			"name":                    resp.User.Name,
			"role":                    resp.User.Role,
			"subscription_plan":       resp.User.SubscriptionPlan,
			"subscription_active":     resp.User.SubscriptionActive,
			"subscription_expires_at": resp.User.SubscriptionExpiresAt,
		},
	})
}

// authorizeUserAccess allows access to /users/:id only for the user themself or an admin.
func authorizeUserAccess(c *gin.Context) (middleware.User, string, bool) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return middleware.User{}, "", false
	}

	userID := c.Param("id")
	if userID == "" {
		apperr.Public(c, http.StatusBadRequest, "user ID is required")
		return middleware.User{}, "", false
	}
	if userID != user.ID && !middleware.IsAdmin(user) {
		apperr.Public(c, http.StatusForbidden, "access denied")
		return middleware.User{}, "", false
	}

	return user, userID, true
}

func (a *AuthAPI) GetUser(c *gin.Context) {
	_, userID, ok := authorizeUserAccess(c)
	if !ok {
		return
	}

	resp, err := a.authClient.GetUser(c.Request.Context(), userID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) UpdateUser(c *gin.Context) {
	user, userID, ok := authorizeUserAccess(c)
	if !ok {
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}
	if req.Role != "" && !middleware.IsAdmin(user) {
		apperr.Public(c, http.StatusForbidden, "only admin can change user role")
		return
	}

	resp, err := a.authClient.UpdateUser(c.Request.Context(), userID, req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AuthAPI) DeleteUser(c *gin.Context) {
	_, userID, ok := authorizeUserAccess(c)
	if !ok {
		return
	}

	resp, err := a.authClient.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp != nil && !resp.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}
