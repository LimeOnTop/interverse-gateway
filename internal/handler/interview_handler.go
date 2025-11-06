package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/api-gateway/internal/service"
	pb "github.com/inter-verse/interview-service/gen"
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

	fmt.Printf("DEBUG: CreateInterview received candidate_id=%s, title=%s\n", req.CandidateID, req.Title)

	// Validate candidate_id
	if req.CandidateID == "" {
		fmt.Printf("DEBUG: CandidateID is empty, returning error\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate_id is required"})
		return
	}

	// Call interview service via gRPC
	response, err := h.interviewService.CreateInterview(req.Title, req.Description, req.CandidateID, req.Specialization, req.Level, req.Duration, req.ScheduledAt, req.TechStack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.CreateInterviewResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	if pbResponse.Interview != nil {
		// Convert technologies array to JSON string
		technologies := pbResponse.Interview.Technologies
		if technologies == nil {
			technologies = []string{}
		}
		techStackJSON, _ := json.Marshal(technologies)

		interviewMap := map[string]interface{}{
			"id":             pbResponse.Interview.Id,
			"title":          pbResponse.Interview.Title,
			"description":    pbResponse.Interview.Description,
			"status":         pbResponse.Interview.Status,
			"scheduled_at":   pbResponse.Interview.ScheduledAt,
			"created_at":     pbResponse.Interview.CreatedAt,
			"updated_at":     pbResponse.Interview.UpdatedAt,
			"level":          pbResponse.Interview.Level,
			"specialization": pbResponse.Interview.Specialization,
			"tech_stack":     string(techStackJSON),
			"duration":       req.Duration,
		}

		// Add candidate if available
		if pbResponse.Interview.Candidate != nil {
			interviewMap["candidate"] = map[string]interface{}{
				"name":  pbResponse.Interview.Candidate.Name,
				"email": pbResponse.Interview.Candidate.Email,
			}
		}

		c.JSON(http.StatusCreated, map[string]interface{}{
			"interview": interviewMap,
		})
	} else {
		c.JSON(http.StatusCreated, response)
	}
}

func (h *InterviewHandler) GetInterviews(c *gin.Context) {
	// Get query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	status := c.Query("status")
	level := c.Query("level")

	fmt.Printf("DEBUG: GetInterviews called with page=%s, limit=%s, status=%s, level=%s\n", page, limit, status, level)

	// Call interview service via gRPC
	response, err := h.interviewService.GetInterviews(page, limit, status, level)
	if err != nil {
		fmt.Printf("DEBUG: GetInterviews error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.GetInterviewsResponse)
	if !ok {
		fmt.Printf("DEBUG: GetInterviews invalid response type\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	fmt.Printf("DEBUG: GetInterviews received %d interviews\n", len(pbResponse.Interviews))

	interviews := make([]map[string]interface{}, 0, len(pbResponse.Interviews))
	for _, interview := range pbResponse.Interviews {
		// Convert technologies array to JSON string
		technologies := interview.Technologies
		if technologies == nil {
			technologies = []string{}
		}
		techStackJSON, _ := json.Marshal(technologies)

		interviewMap := map[string]interface{}{
			"id":             interview.Id,
			"title":          interview.Title,
			"description":    interview.Description,
			"status":         interview.Status,
			"scheduled_at":   interview.ScheduledAt,
			"created_at":     interview.CreatedAt,
			"updated_at":     interview.UpdatedAt,
			"level":          interview.Level,
			"specialization": interview.Specialization,
			"tech_stack":     string(techStackJSON),
			"duration":       60, // Default duration, should be added to proto
		}

		// Add candidate if available
		if interview.Candidate != nil {
			interviewMap["candidate"] = map[string]interface{}{
				"name":  interview.Candidate.Name,
				"email": interview.Candidate.Email,
			}
		}

		interviews = append(interviews, interviewMap)
	}

	result := map[string]interface{}{
		"interviews": interviews,
	}
	if pbResponse.Pagination != nil {
		result["pagination"] = map[string]interface{}{
			"page":  pbResponse.Pagination.Page,
			"limit": pbResponse.Pagination.Limit,
			"total": pbResponse.Pagination.Total,
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *InterviewHandler) GetInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	// Call interview service via gRPC
	response, err := h.interviewService.GetInterview(interviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.GetInterviewResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	if pbResponse.Interview != nil {
		// Convert technologies array to JSON string
		technologies := pbResponse.Interview.Technologies
		if technologies == nil {
			technologies = []string{}
		}
		techStackJSON, _ := json.Marshal(technologies)

		interviewMap := map[string]interface{}{
			"id":             pbResponse.Interview.Id,
			"title":          pbResponse.Interview.Title,
			"description":    pbResponse.Interview.Description,
			"status":         pbResponse.Interview.Status,
			"scheduled_at":   pbResponse.Interview.ScheduledAt,
			"created_at":     pbResponse.Interview.CreatedAt,
			"updated_at":     pbResponse.Interview.UpdatedAt,
			"level":          pbResponse.Interview.Level,
			"specialization": pbResponse.Interview.Specialization,
			"tech_stack":     string(techStackJSON),
			"candidate_id":   pbResponse.Interview.CandidateId,
		}

		// Add candidate if available
		if pbResponse.Interview.Candidate != nil {
			interviewMap["candidate"] = map[string]interface{}{
				"name":  pbResponse.Interview.Candidate.Name,
				"email": pbResponse.Interview.Candidate.Email,
			}
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"interview": interviewMap,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
	}
}

func (h *InterviewHandler) UpdateInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		Status         string `json:"status"`
		ScheduledAt    string `json:"scheduled_at,omitempty"`
		Specialization string `json:"specialization"`
		Level          string `json:"level"`
		TechStack      string `json:"tech_stack,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call interview service via gRPC
	response, err := h.interviewService.UpdateInterview(interviewID, req.Title, req.Description, req.Status, req.ScheduledAt, req.Specialization, req.Level, req.TechStack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.UpdateInterviewResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	if pbResponse.Interview != nil {
		// Convert technologies array to JSON string
		technologies := pbResponse.Interview.Technologies
		if technologies == nil {
			technologies = []string{}
		}
		techStackJSON, _ := json.Marshal(technologies)

		interviewMap := map[string]interface{}{
			"id":             pbResponse.Interview.Id,
			"title":          pbResponse.Interview.Title,
			"description":    pbResponse.Interview.Description,
			"status":         pbResponse.Interview.Status,
			"scheduled_at":   pbResponse.Interview.ScheduledAt,
			"created_at":     pbResponse.Interview.CreatedAt,
			"updated_at":     pbResponse.Interview.UpdatedAt,
			"level":          pbResponse.Interview.Level,
			"specialization": pbResponse.Interview.Specialization,
			"tech_stack":     string(techStackJSON),
		}

		// Add candidate if available
		if pbResponse.Interview.Candidate != nil {
			interviewMap["candidate"] = map[string]interface{}{
				"name":  pbResponse.Interview.Candidate.Name,
				"email": pbResponse.Interview.Candidate.Email,
			}
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"interview": interviewMap,
		})
	} else {
		c.JSON(http.StatusOK, response)
	}
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

	// Convert gRPC response to frontend format
	pbResponse, ok := response.(*pb.GetScheduledInterviewsResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	interviews := make([]map[string]interface{}, 0, len(pbResponse.Interviews))
	for _, interview := range pbResponse.Interviews {
		// Convert technologies array to JSON string
		techStackJSON, _ := json.Marshal(interview.Technologies)

		interviewMap := map[string]interface{}{
			"id":             interview.Id,
			"title":          interview.Title,
			"description":    interview.Description,
			"status":         interview.Status,
			"scheduled_at":   interview.ScheduledAt,
			"created_at":     interview.CreatedAt,
			"updated_at":     interview.UpdatedAt,
			"level":          interview.Level,
			"specialization": interview.Specialization,
			"tech_stack":     string(techStackJSON),
			"duration":       60, // Default duration, should be added to proto
		}

		// Add candidate if available
		if interview.Candidate != nil {
			interviewMap["candidate"] = map[string]interface{}{
				"name":  interview.Candidate.Name,
				"email": interview.Candidate.Email,
			}
		}

		interviews = append(interviews, interviewMap)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"interviews": interviews,
	})
}

func (h *InterviewHandler) GenerateQuestions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GenerateQuestions endpoint - to be implemented"})
}
