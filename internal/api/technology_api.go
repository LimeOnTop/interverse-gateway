package api

import (
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

type TechnologyAPI struct {
	technologyClient *clients.TechnologyClient
}

func NewTechnologyAPI(technologyClient *clients.TechnologyClient) *TechnologyAPI {
	return &TechnologyAPI{
		technologyClient: technologyClient,
	}
}

func (a *TechnologyAPI) GetTechnologies(c *gin.Context) {
	page, limit := parsePagination(c)

	resp, err := a.technologyClient.GetTechnologies(c.Request.Context(), page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *TechnologyAPI) GetTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		apperr.Public(c, http.StatusBadRequest, "technology ID is required")
		return
	}

	resp, err := a.technologyClient.GetTechnology(c.Request.Context(), technologyID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *TechnologyAPI) CreateTechnology(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Category    string   `json:"category" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.technologyClient.CreateTechnology(c.Request.Context(), req.Name, req.Category, req.Description, req.Tags)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (a *TechnologyAPI) UpdateTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		apperr.Public(c, http.StatusBadRequest, "technology ID is required")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.technologyClient.UpdateTechnology(c.Request.Context(), technologyID, req.Name, req.Category, req.Description, req.Tags)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *TechnologyAPI) DeleteTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		apperr.Public(c, http.StatusBadRequest, "technology ID is required")
		return
	}

	resp, err := a.technologyClient.DeleteTechnology(c.Request.Context(), technologyID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp != nil && !resp.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *TechnologyAPI) SearchTechnologies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		apperr.Public(c, http.StatusBadRequest, "search query is required")
		return
	}

	page, limit := parsePagination(c)

	resp, err := a.technologyClient.SearchTechnologies(c.Request.Context(), query, page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	c.JSON(http.StatusOK, resp)
}
