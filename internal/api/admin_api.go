package api

import (
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if resp.Response != nil && !resp.Response.Success {
			c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	resp, err := a.questionClient.GetQuestionsByTechnology(c.Request.Context(), technology, difficulty, category, page, limit)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *AdminAPI) ApproveModerationQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	difficulty := strings.TrimSpace(req.Difficulty)
	switch difficulty {
	case "junior", "middle", "senior":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be junior, middle or senior"})
		return
	}

	existing, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.Response != nil && !existing.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": existing.Response.Error})
		return
	}
	if existing.Question == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	q := existing.Question
	if q.Difficulty != "pending_moderation" && q.Category != "contribution" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is not pending moderation"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
		return
	}

	existing, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.Question == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	resp, err := a.questionClient.DeleteQuestion(c.Request.Context(), questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp != nil && !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вопрос отклонён и удалён"})
}
