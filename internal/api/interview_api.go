package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type InterviewAPI struct {
	interviewClient *clients.InterviewClient
}

func NewInterviewAPI(interviewClient *clients.InterviewClient) *InterviewAPI {
	return &InterviewAPI{
		interviewClient: interviewClient,
	}
}

func parseTechStack(techStack string) []string {
	if techStack == "" {
		return nil
	}
	var technologies []string
	if err := json.Unmarshal([]byte(techStack), &technologies); err != nil {
		return []string{techStack}
	}
	return technologies
}

func (a *InterviewAPI) CreateInterview(c *gin.Context) {
	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		Specialization string `json:"specialization"`
		Level          string `json:"level"`
		ScheduledAt    string `json:"scheduled_at,omitempty"`
		TechStack      string `json:"tech_stack,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.CreateInterview(
		c.Request.Context(),
		user.ID, req.Title, req.Description, req.ScheduledAt,
		req.Level, req.Specialization, parseTechStack(req.TechStack),
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

func (a *InterviewAPI) GetInterviews(c *gin.Context) {
	page, limit := parsePagination(c)
	status := c.Query("status")

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.GetInterviews(c.Request.Context(), user.ID, status, page, limit)
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

func (a *InterviewAPI) GetInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	resp, err := a.interviewClient.GetInterview(c.Request.Context(), interviewID)
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

func (a *InterviewAPI) UpdateInterview(c *gin.Context) {
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

	resp, err := a.interviewClient.UpdateInterview(
		c.Request.Context(),
		interviewID, req.Title, req.Description, req.Status, req.ScheduledAt,
		req.Level, req.Specialization, parseTechStack(req.TechStack),
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

func (a *InterviewAPI) DeleteInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	resp, err := a.interviewClient.DeleteInterview(c.Request.Context(), interviewID)
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

func (a *InterviewAPI) GetScheduledInterviews(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.GetScheduledInterviews(c.Request.Context(), user.ID, date)
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

func (a *InterviewAPI) StartSession(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.StartSession(c.Request.Context(), interviewID, user.ID)
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

func (a *InterviewAPI) GetSessionContent(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interview ID is required"})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.GetSessionContent(c.Request.Context(), interviewID, user.ID)
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

// GenerateQuestions keeps backward compatibility with the frontend alias.
func (a *InterviewAPI) GenerateQuestions(c *gin.Context) {
	a.StartSession(c)
}

func (a *InterviewAPI) StartSessionFromBody(c *gin.Context) {
	var req struct {
		InterviewID string `json:"interview_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.interviewClient.StartSession(c.Request.Context(), req.InterviewID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": resp.Interview.GetId(),
		"interview":  resp.Interview,
		"questions":  resp.Questions,
		"tasks":      resp.Tasks,
	})
}
