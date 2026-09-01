package api

import (
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type CandidateAPI struct {
	candidateClient *clients.CandidateClient
}

func NewCandidateAPI(candidateClient *clients.CandidateClient) *CandidateAPI {
	return &CandidateAPI{
		candidateClient: candidateClient,
	}
}

func (a *CandidateAPI) CreateCandidate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Email       string `json:"email" binding:"omitempty,email"`
		Phone       string `json:"phone"`
		Position    string `json:"position"`
		Experience  int64  `json:"experience"`
		Skills      string `json:"skills"`
		ResumeURL   string `json:"resume_url"`
		LinkedinURL string `json:"linkedin_url"`
		GithubURL   string `json:"github_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.candidateClient.CreateCandidate(
		c.Request.Context(),
		req.Name, req.Email, req.Phone, req.Position, req.Experience,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, user.ID,
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

func (a *CandidateAPI) GetCandidates(c *gin.Context) {
	page, limit := parsePagination(c)

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.candidateClient.GetCandidates(c.Request.Context(), user.ID, page, limit)
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

func (a *CandidateAPI) GetCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	resp, err := a.candidateClient.GetCandidate(c.Request.Context(), candidateID)
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

func (a *CandidateAPI) UpdateCandidate(c *gin.Context) {
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
		Experience  int64  `json:"experience"`
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

	resp, err := a.candidateClient.UpdateCandidate(
		c.Request.Context(),
		candidateID, req.Name, req.Email, req.Phone, req.Position, req.Experience,
		req.Skills, req.ResumeURL, req.LinkedinURL, req.GithubURL, req.Status,
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

func (a *CandidateAPI) DeleteCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	resp, err := a.candidateClient.DeleteCandidate(c.Request.Context(), candidateID)
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

func (a *CandidateAPI) SearchCandidates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	page, limit := parsePagination(c)

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.candidateClient.SearchCandidates(c.Request.Context(), query, user.ID, page, limit)
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
