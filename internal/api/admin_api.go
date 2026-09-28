package api

import (
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"net/http"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

type AdminAPI struct {
	questionClient *clients.QuestionClient
}

func NewAdminAPI(questionClient *clients.QuestionClient) *AdminAPI {
	return &AdminAPI{questionClient: questionClient}
}

func (a *AdminAPI) Stats(c *gin.Context) {
	ctx := c.Request.Context()

	totalResp, err := a.questionClient.GetQuestions(ctx, 1, 1)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	type levelStat struct {
		Difficulty string `json:"difficulty"`
		Total      int32  `json:"total"`
	}

	levels := []string{"junior", "middle", "senior", "pending_moderation"}
	byDifficulty := make([]levelStat, 0, len(levels))
	for _, level := range levels {
		category := ""
		if level == "pending_moderation" {
			category = "contribution"
		}
		resp, err := a.questionClient.GetQuestionsByTechnology(ctx, "", level, category, 1, 1)
		if err != nil {
			apperr.Internal(c, err)
			return
		}
		total := int32(0)
		if resp.Pagination != nil {
			total = resp.Pagination.Total
		}
		byDifficulty = append(byDifficulty, levelStat{Difficulty: level, Total: total})
	}

	total := int32(0)
	if totalResp.Pagination != nil {
		total = totalResp.Pagination.Total
	}

	c.JSON(http.StatusOK, gin.H{
		"total_questions": total,
		"by_difficulty":   byDifficulty,
	})
}

func (a *AdminAPI) ListQuestions(c *gin.Context) {
	page, limit := parsePagination(c)
	if limit > 100 {
		limit = 100
	}

	technology := strings.TrimSpace(c.Query("technology"))
	difficulty := strings.TrimSpace(c.Query("difficulty"))
	category := strings.TrimSpace(c.Query("category"))

	// No filters → full bank (all directions and levels).
	if technology == "" && difficulty == "" && category == "" {
		resp, err := a.questionClient.GetQuestions(c.Request.Context(), page, limit)
		if err != nil {
			apperr.Internal(c, err)
			return
		}
		if resp.Response != nil && !resp.Response.Success {
			apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	resp, err := a.questionClient.GetQuestionsByTechnology(c.Request.Context(), technology, difficulty, category, page, limit)
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

func (a *AdminAPI) ListModerationQuestions(c *gin.Context) {
	page, limit := parsePagination(c)
	if limit > 100 {
		limit = 100
	}

	resp, err := a.questionClient.GetQuestionsByTechnology(
		c.Request.Context(),
		c.Query("technology"),
		"pending_moderation",
		"contribution",
		page,
		limit,
	)
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

func (a *AdminAPI) ApproveModerationQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	var req struct {
		Difficulty string `json:"difficulty" binding:"required"`
		Technology string `json:"technology"`
		Category   string `json:"category"`
		Text       string `json:"text"`
		Answer     string `json:"answer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	difficulty := strings.TrimSpace(req.Difficulty)
	switch difficulty {
	case "junior", "middle", "senior":
	default:
		apperr.Public(c, http.StatusBadRequest, "difficulty must be junior, middle or senior")
		return
	}

	existing, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if existing.Response != nil && !existing.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, existing.Response.Error)
		return
	}
	if existing.Question == nil {
		apperr.Public(c, http.StatusNotFound, "question not found")
		return
	}
	q := existing.Question
	if q.Difficulty != "pending_moderation" && q.Category != "contribution" {
		apperr.Public(c, http.StatusBadRequest, "question is not pending moderation")
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		text = q.Text
	}
	answer := strings.TrimSpace(req.Answer)
	if answer == "" {
		answer = q.Answer
	}
	technology := strings.TrimSpace(req.Technology)
	if technology == "" {
		technology = q.Technology
	}
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "question"
	}

	tags := append([]string{}, q.Tags...)
	hasModerated := false
	for _, tag := range tags {
		if tag == "moderated" {
			hasModerated = true
			break
		}
	}
	if !hasModerated {
		tags = append(tags, "moderated")
	}

	resp, err := a.questionClient.UpdateQuestion(
		c.Request.Context(),
		questionID,
		text,
		category,
		difficulty,
		technology,
		tags,
		answer,
		q.Options,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Вопрос одобрен и добавлен в банк",
		"question": resp.Question,
	})
}

func (a *AdminAPI) RejectModerationQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	existing, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if existing.Question == nil {
		apperr.Public(c, http.StatusNotFound, "question not found")
		return
	}

	resp, err := a.questionClient.DeleteQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp != nil && !resp.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вопрос отклонён и удалён"})
}
