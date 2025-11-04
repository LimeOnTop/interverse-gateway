package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/api-gateway/internal/service"
)

type QuestionHandler struct {
	questionService *service.QuestionService
}

func NewQuestionHandler(questionService *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
	}
}

func (h *QuestionHandler) GetQuestions(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	response, err := h.questionService.GetQuestions(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuestionHandler) GetQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
		return
	}

	response, err := h.questionService.GetQuestion(questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuestionHandler) CreateQuestion(c *gin.Context) {
	var req struct {
		Text       string   `json:"text" binding:"required"`
		Category   string   `json:"category" binding:"required"`
		Difficulty string   `json:"difficulty" binding:"required"`
		Technology string   `json:"technology" binding:"required"`
		Tags       []string `json:"tags"`
		Answer     string   `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.questionService.CreateQuestion(req.Text, req.Category, req.Difficulty, req.Technology, req.Tags, req.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *QuestionHandler) UpdateQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
		return
	}

	var req struct {
		Text       string   `json:"text"`
		Category   string   `json:"category"`
		Difficulty string   `json:"difficulty"`
		Technology string   `json:"technology"`
		Tags       []string `json:"tags"`
		Answer     string   `json:"answer"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.questionService.UpdateQuestion(questionID, req.Text, req.Category, req.Difficulty, req.Technology, req.Tags, req.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuestionHandler) DeleteQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
		return
	}

	response, err := h.questionService.DeleteQuestion(questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuestionHandler) SearchQuestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	response, err := h.questionService.SearchQuestions(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuestionHandler) GetQuestionsByTechnology(c *gin.Context) {
	technology := c.Query("technology")
	if technology == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "technology parameter is required"})
		return
	}

	difficulty := c.Query("difficulty")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	response, err := h.questionService.GetQuestionsByTechnology(technology, difficulty, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
