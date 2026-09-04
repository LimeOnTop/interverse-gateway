package api

import (
	"net/http"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type ContributionAPI struct {
	questionClient *clients.QuestionClient
}

func NewContributionAPI(questionClient *clients.QuestionClient) *ContributionAPI {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text := strings.TrimSpace(req.Text)
	if len(text) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question text is too short"})
		return
	}

	answer := strings.TrimSpace(req.Answer)
	if answer == "" {
		answer = "Ожидает модерации"
	}

	technology := strings.TrimSpace(req.Technology)
	if technology == "" {
		technology = "Общее"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Спасибо! Вопрос отправлен на модерацию",
		"question": resp.Question,
	})
}
