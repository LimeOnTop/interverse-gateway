package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/api-gateway/internal/service"
	pb "github.com/inter-verse/candidate-service/gen"
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
		Name        string      `json:"name" binding:"required"`
		Email       string      `json:"email" binding:"omitempty,email"`
		Phone       string      `json:"phone"`
		Position    string      `json:"position"`
		Experience  interface{} `json:"experience"`
		Skills      string      `json:"skills"`
		ResumeURL   string      `json:"resume_url"`
		LinkedinURL string      `json:"linkedin_url"`
		GithubURL   string      `json:"github_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert experience to string
	experienceStr := ""
	if req.Experience != nil {
		switch v := req.Experience.(type) {
		case string:
			experienceStr = v
		case float64:
			experienceStr = fmt.Sprintf("%.0f", v)
		case int:
			experienceStr = fmt.Sprintf("%d", v)
		case int64:
			experienceStr = fmt.Sprintf("%d", v)
		}
	}

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.candidateService.CreateCandidate(
		req.Name, req.Email, req.Phone, req.Position, experienceStr,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, interviewerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.CreateCandidateResponse)
	if !ok {
		fmt.Printf("DEBUG: CreateCandidate invalid response type\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	fmt.Printf("DEBUG: CreateCandidate response: Candidate=%v, Response.Success=%v\n", pbResponse.Candidate != nil, pbResponse.Response != nil && pbResponse.Response.Success)

	if pbResponse.Candidate != nil {
		fmt.Printf("DEBUG: CreateCandidate returning candidate with id=%s\n", pbResponse.Candidate.Id)
		candidateMap := map[string]interface{}{
			"id":             pbResponse.Candidate.Id,
			"name":           pbResponse.Candidate.Name,
			"email":          pbResponse.Candidate.Email,
			"phone":          pbResponse.Candidate.Phone,
			"position":       pbResponse.Candidate.Position,
			"experience":     pbResponse.Candidate.Experience,
			"skills":         pbResponse.Candidate.Skills,
			"resume_url":     pbResponse.Candidate.ResumeUrl,
			"linkedin_url":   pbResponse.Candidate.LinkedinUrl,
			"github_url":     pbResponse.Candidate.GithubUrl,
			"status":         pbResponse.Candidate.Status,
			"created_at":     pbResponse.Candidate.CreatedAt,
			"updated_at":     pbResponse.Candidate.UpdatedAt,
			"interviewer_id": pbResponse.Candidate.InterviewerId,
		}
		c.JSON(http.StatusCreated, candidateMap)
	} else {
		fmt.Printf("DEBUG: CreateCandidate Candidate is nil, returning raw response\n")
		c.JSON(http.StatusCreated, response)
	}
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

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.GetCandidateResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	if pbResponse.Candidate != nil {
		candidateMap := map[string]interface{}{
			"id":             pbResponse.Candidate.Id,
			"name":           pbResponse.Candidate.Name,
			"email":          pbResponse.Candidate.Email,
			"phone":          pbResponse.Candidate.Phone,
			"position":       pbResponse.Candidate.Position,
			"experience":     pbResponse.Candidate.Experience,
			"skills":         pbResponse.Candidate.Skills,
			"resume_url":     pbResponse.Candidate.ResumeUrl,
			"linkedin_url":   pbResponse.Candidate.LinkedinUrl,
			"github_url":     pbResponse.Candidate.GithubUrl,
			"status":         pbResponse.Candidate.Status,
			"created_at":     pbResponse.Candidate.CreatedAt,
			"updated_at":     pbResponse.Candidate.UpdatedAt,
			"interviewer_id": pbResponse.Candidate.InterviewerId,
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"candidate": candidateMap,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidate not found"})
	}
}

func (h *CandidateHandler) UpdateCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	var req struct {
		Name        string      `json:"name"`
		Email       string      `json:"email"`
		Phone       string      `json:"phone"`
		Position    string      `json:"position"`
		Experience  interface{} `json:"experience"`
		Skills      string      `json:"skills"`
		ResumeURL   string      `json:"resume_url"`
		LinkedinURL string      `json:"linkedin_url"`
		GithubURL   string      `json:"github_url"`
		Status      string      `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert experience to string
	experienceStr := ""
	if req.Experience != nil {
		switch v := req.Experience.(type) {
		case string:
			experienceStr = v
		case float64:
			experienceStr = fmt.Sprintf("%.0f", v)
		case int:
			experienceStr = fmt.Sprintf("%d", v)
		case int64:
			experienceStr = fmt.Sprintf("%d", v)
		}
	}

	response, err := h.candidateService.UpdateCandidate(
		candidateID, req.Name, req.Email, req.Phone, req.Position, experienceStr,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, req.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.UpdateCandidateResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	if pbResponse.Candidate != nil {
		candidateMap := map[string]interface{}{
			"id":             pbResponse.Candidate.Id,
			"name":           pbResponse.Candidate.Name,
			"email":          pbResponse.Candidate.Email,
			"phone":          pbResponse.Candidate.Phone,
			"position":       pbResponse.Candidate.Position,
			"experience":     pbResponse.Candidate.Experience,
			"skills":         pbResponse.Candidate.Skills,
			"resume_url":     pbResponse.Candidate.ResumeUrl,
			"linkedin_url":   pbResponse.Candidate.LinkedinUrl,
			"github_url":     pbResponse.Candidate.GithubUrl,
			"status":         pbResponse.Candidate.Status,
			"created_at":     pbResponse.Candidate.CreatedAt,
			"updated_at":     pbResponse.Candidate.UpdatedAt,
			"interviewer_id": pbResponse.Candidate.InterviewerId,
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"candidate": candidateMap,
		})
	} else {
		c.JSON(http.StatusOK, response)
	}
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
