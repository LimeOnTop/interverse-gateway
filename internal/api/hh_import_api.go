package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/hh"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type HHImportAPI struct {
	hhClient *hh.Client
}

func NewHHImportAPI(hhClient *hh.Client) *HHImportAPI {
	return &HHImportAPI{hhClient: hhClient}
}

func (a *HHImportAPI) GetAuthURL(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	userID := c.Param("id")
	if userID == "" || userID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if a.hhClient == nil || !a.hhClient.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "hh.ru import is not configured",
			"message": "Задайте HH_CLIENT_ID, HH_CLIENT_SECRET и HH_REDIRECT_URI. Приложение регистрируется на https://dev.hh.ru/",
		})
		return
	}

	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create oauth state"})
		return
	}

	authURL, err := a.hhClient.AuthURL(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

func (a *HHImportAPI) ImportProfile(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	userID := c.Param("id")
	if userID == "" || userID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if a.hhClient == nil || !a.hhClient.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "hh.ru import is not configured",
			"message": "Задайте HH_CLIENT_ID, HH_CLIENT_SECRET и HH_REDIRECT_URI. Приложение регистрируется на https://dev.hh.ru/",
		})
		return
	}

	var req struct {
		Code      string `json:"code"`
		ResumeURL string `json:"resume_url"`
		ResumeID  string `json:"resume_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	token, err := a.hhClient.ExchangeCode(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preferredID := req.ResumeID
	if preferredID == "" {
		preferredID = hh.ExtractResumeID(req.ResumeURL)
	}

	items, err := a.hhClient.ListMineResumes(c.Request.Context(), token.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resumeID := hh.PickResumeID(items, preferredID)
	if resumeID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "у аккаунта hh.ru нет резюме"})
		return
	}

	resume, err := a.hhClient.GetResume(c.Request.Context(), token.AccessToken, resumeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	avatarURL := ""
	if photoURL := resume.PhotoURL(); photoURL != "" {
		avatarURL, err = a.hhClient.DownloadImageAsDataURL(c.Request.Context(), token.AccessToken, photoURL)
		if err != nil {
			// Keep import usable even if photo download fails.
			avatarURL = photoURL
		}
	}

	imported := hh.MapResume(resume, avatarURL)
	c.JSON(http.StatusOK, gin.H{
		"profile": imported,
		"message": "Данные резюме hh.ru загружены. Проверьте поля и сохраните профиль.",
	})
}

func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
