package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/services/api-gateway/internal/service"
)

type InterviewHandler struct {
	interviewService *service.InterviewService
}

func NewInterviewHandler(interviewService *service.InterviewService) *InterviewHandler {
	return &InterviewHandler{
		interviewService: interviewService,
	}
}

func (h *InterviewHandler) CreateInterview(c *gin.Context) {
	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		CandidateID    string `json:"candidate_id"`
		Specialization string `json:"specialization"`
		Level          string `json:"level"`
		Duration       int    `json:"duration"`
		ScheduledAt    string `json:"scheduled_at,omitempty"`
		TechStack      string `json:"tech_stack,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call interview service via gRPC
	response, err := h.interviewService.CreateInterview(req.Title, req.Description, req.CandidateID, req.Specialization, req.Level, req.Duration, req.ScheduledAt, req.TechStack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *InterviewHandler) GetInterviews(c *gin.Context) {
	// Get query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	status := c.Query("status")

	// Call interview service via gRPC
	response, err := h.interviewService.GetInterviews(page, limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InterviewHandler) GetInterview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetInterview endpoint - to be implemented"})
}

func (h *InterviewHandler) UpdateInterview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateInterview endpoint - to be implemented"})
}

func (h *InterviewHandler) DeleteInterview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteInterview endpoint - to be implemented"})
}

func (h *InterviewHandler) GetScheduledInterviews(c *gin.Context) {
	// Get query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Call interview service via gRPC
	response, err := h.interviewService.GetScheduledInterviews(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InterviewHandler) GenerateQuestions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GenerateQuestions endpoint - to be implemented"})
}
