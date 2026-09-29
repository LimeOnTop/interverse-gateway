package api

import (
	"net/http"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
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

// ImportProfile parses a publicly viewable hh.ru resume by URL (no OAuth).
func (a *HHImportAPI) ImportProfile(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	userID := c.Param("id")
	if userID == "" || userID != user.ID {
		apperr.Public(c, http.StatusForbidden, "access denied")
		return
	}

	if a.hhClient == nil {
		apperr.Public(c, http.StatusServiceUnavailable, "Импорт из hh.ru временно недоступен")
		return
	}

	var req struct {
		ResumeURL string `json:"resume_url"`
		ResumeID  string `json:"resume_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resumeRef := strings.TrimSpace(req.ResumeURL)
	if resumeRef == "" {
		resumeRef = strings.TrimSpace(req.ResumeID)
	}
	if resumeRef == "" {
		apperr.Public(c, http.StatusBadRequest, "укажите ссылку на резюме hh.ru")
		return
	}

	resume, err := a.hhClient.FetchPublicResume(c.Request.Context(), resumeRef)
	if err != nil {
		msg := err.Error()
		if isPublicHHClientError(msg) {
			apperr.Public(c, http.StatusBadRequest, msg)
			return
		}
		apperr.Internal(c, err)
		return
	}

	avatarURL := ""
	if photoURL := resume.PhotoURL(); photoURL != "" {
		avatarURL, err = a.hhClient.DownloadImageAsDataURL(c.Request.Context(), "", photoURL)
		if err != nil {
			avatarURL = photoURL
		}
	}

	imported := hh.MapResume(resume, avatarURL)
	c.JSON(http.StatusOK, gin.H{
		"profile": imported,
		"message": "Данные резюме hh.ru загружены. Проверьте поля и сохраните профиль.",
	})
}

func isPublicHHClientError(msg string) bool {
	lower := strings.ToLower(msg)
	markers := []string{
		"укажите ссылку",
		"некорректная ссылка",
		"ожидается ссылка",
		"не найдено",
		"доступ к резюме ограничен",
		"открыто для просмотра",
		"капчу",
		"не удалось прочитать",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}
