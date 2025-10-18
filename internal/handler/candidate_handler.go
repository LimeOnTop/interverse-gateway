package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/services/api-gateway/internal/service"
)

type CandidateHandler struct {
	candidateService *service.CandidateService
}

func NewCandidateHandler(candidateService *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{
		candidateService: candidateService,
	}
}

func (h *CandidateHandler) CreateCandidate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Phone       string `json:"phone"`
		Position    string `json:"position"`
		Experience  string `json:"experience"`
		Skills      string `json:"skills"`
		ResumeURL   string `json:"resume_url"`
		LinkedinURL string `json:"linkedin_url"`
		GithubURL   string `json:"github_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.candidateService.CreateCandidate(
		req.Name, req.Email, req.Phone, req.Position, req.Experience,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, interviewerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *CandidateHandler) GetCandidates(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.candidateService.GetCandidates(interviewerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CandidateHandler) GetCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	response, err := h.candidateService.GetCandidate(candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CandidateHandler) UpdateCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Position    string `json:"position"`
		Experience  string `json:"experience"`
		Skills      string `json:"skills"`
		ResumeURL   string `json:"resume_url"`
		LinkedinURL string `json:"linkedin_url"`
		GithubURL   string `json:"github_url"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.candidateService.UpdateCandidate(
		candidateID, req.Name, req.Email, req.Phone, req.Position, req.Experience,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, req.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CandidateHandler) DeleteCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	response, err := h.candidateService.DeleteCandidate(candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CandidateHandler) SearchCandidates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.candidateService.SearchCandidates(query, interviewerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
