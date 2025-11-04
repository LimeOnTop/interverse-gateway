package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/api-gateway/internal/service"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) CreateReport(c *gin.Context) {
	var req struct {
		InterviewID         string `json:"interview_id" binding:"required"`
		CandidateID         string `json:"candidate_id" binding:"required"`
		OverallRating       string `json:"overall_rating" binding:"required"`
		TechnicalSkills     string `json:"technical_skills" binding:"required"`
		CommunicationSkills string `json:"communication_skills" binding:"required"`
		ProblemSolving      string `json:"problem_solving" binding:"required"`
		Strengths           string `json:"strengths"`
		Weaknesses          string `json:"weaknesses"`
		Recommendations     string `json:"recommendations"`
		Notes               string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.reportService.CreateReport(
		req.InterviewID, req.CandidateID, interviewerID, req.OverallRating,
		req.TechnicalSkills, req.CommunicationSkills, req.ProblemSolving,
		req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ReportHandler) GetReports(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// TODO: Get interviewer ID from auth context
	interviewerID := "550e8400-e29b-41d4-a716-446655440000"

	response, err := h.reportService.GetReports(interviewerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReportHandler) GetReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	response, err := h.reportService.GetReport(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReportHandler) UpdateReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	var req struct {
		OverallRating       string `json:"overall_rating"`
		TechnicalSkills     string `json:"technical_skills"`
		CommunicationSkills string `json:"communication_skills"`
		ProblemSolving      string `json:"problem_solving"`
		Strengths           string `json:"strengths"`
		Weaknesses          string `json:"weaknesses"`
		Recommendations     string `json:"recommendations"`
		Notes               string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.reportService.UpdateReport(
		reportID, req.OverallRating, req.TechnicalSkills, req.CommunicationSkills,
		req.ProblemSolving, req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReportHandler) DeleteReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	response, err := h.reportService.DeleteReport(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
