package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ContributionAPI struct {
	questionClient usecase.QuestionGateway
}

func NewContributionAPI(questionClient usecase.QuestionGateway) *ContributionAPI {
	return &ContributionAPI{questionClient: questionClient}
}

func (a *ContributionAPI) SubmitQuestion(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	var req struct {
		Text       string `json:"text" binding:"required"`
		Answer     string `json:"answer"`
		Technology string `json:"technology"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 10 || utf8.RuneCountInString(text) > 10000 {
		apperr.Public(c, http.StatusBadRequest, "Текст вопроса должен содержать от 10 до 10 000 символов")
		return
	}

	answer := strings.TrimSpace(req.Answer)
	if answer == "" {
		answer = "Ожидает модерации"
	}

	technology := strings.TrimSpace(req.Technology)
	if technology == "" {
		apperr.Public(c, http.StatusBadRequest, "Выбери технологию")
		return
	}

	resp, err := a.questionClient.CreateQuestion(
		c.Request.Context(),
		text,
		"contribution",
		"pending_moderation",
		technology,
		[]string{"user_contribution", "user:" + user.ID},
		answer,
		nil,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Вопрос отправлен на модерацию. После одобрения начислим 30 дней Pro.",
		"question": resp.Question,
	})
}
