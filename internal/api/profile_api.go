package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
)

type ProfileAPI struct {
	profileClient *clients.ProfileClient
}

func NewProfileAPI(profileClient *clients.ProfileClient) *ProfileAPI {
	return &ProfileAPI{profileClient: profileClient}
}

func (a *ProfileAPI) GetProfile(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}
	if userID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	resp, err := a.profileClient.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   resp.Response.Error,
			"message": resp.Response.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": resp.Profile})
}

func (a *ProfileAPI) UpdateProfile(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}
	if userID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		WorkExperience  string `json:"work_experience"`
		AvatarURL       string `json:"avatar_url"`
		AboutMe         string `json:"about_me"`
		HigherEducation string `json:"higher_education"`
		EnglishLevel    string `json:"english_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.profileClient.UpdateProfile(
		c.Request.Context(),
		userID,
		req.WorkExperience,
		req.AvatarURL,
		req.AboutMe,
		req.HigherEducation,
		req.EnglishLevel,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   resp.Response.Error,
			"message": resp.Response.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": resp.Profile})
}
