package api

import (
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ProfileAPI struct {
	profileClient usecase.ProfileGateway
}

func NewProfileAPI(profileClient usecase.ProfileGateway) *ProfileAPI {
	return &ProfileAPI{profileClient: profileClient}
}

func (a *ProfileAPI) GetProfile(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		apperr.Public(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := c.Param("id")
	if userID == "" {
		apperr.Public(c, http.StatusBadRequest, "user ID is required")
		return
	}
	if userID != user.ID {
		apperr.Public(c, http.StatusForbidden, "access denied")
		return
	}

	resp, err := a.profileClient.GetProfile(c.Request.Context(), userID)
	if err != nil {
		apperr.Internal(c, err)
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
		apperr.Public(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := c.Param("id")
	if userID == "" {
		apperr.Public(c, http.StatusBadRequest, "user ID is required")
		return
	}
	if userID != user.ID {
		apperr.Public(c, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		WorkExperience  string `json:"work_experience"`
		AvatarURL       string `json:"avatar_url"`
		AboutMe         string `json:"about_me"`
		HigherEducation string `json:"higher_education"`
		EnglishLevel    string `json:"english_level"`
		Skills          string `json:"skills"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
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
		req.Skills,
	)
	if err != nil {
		apperr.Internal(c, err)
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
