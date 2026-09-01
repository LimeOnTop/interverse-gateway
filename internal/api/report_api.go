package api

import (
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type ReportAPI struct {
	reportClient *clients.ReportClient
}

func NewReportAPI(reportClient *clients.ReportClient) *ReportAPI {
	return &ReportAPI{
		reportClient: reportClient,
	}
}

func (a *ReportAPI) CreateReport(c *gin.Context) {
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

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.reportClient.CreateReport(
		c.Request.Context(),
		req.InterviewID, req.CandidateID, user.ID, req.OverallRating,
		req.TechnicalSkills, req.CommunicationSkills, req.ProblemSolving,
		req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (a *ReportAPI) GetReports(c *gin.Context) {
	page, limit := parsePagination(c)

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.reportClient.GetReports(c.Request.Context(), user.ID, page, limit)
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

func (a *ReportAPI) GetReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	resp, err := a.reportClient.GetReport(c.Request.Context(), reportID)
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

func (a *ReportAPI) UpdateReport(c *gin.Context) {
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

	resp, err := a.reportClient.UpdateReport(
		c.Request.Context(),
		reportID, req.OverallRating, req.TechnicalSkills, req.CommunicationSkills,
		req.ProblemSolving, req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
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

func (a *ReportAPI) DeleteReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	resp, err := a.reportClient.DeleteReport(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp != nil && !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	c.JSON(http.StatusOK, resp)
}
