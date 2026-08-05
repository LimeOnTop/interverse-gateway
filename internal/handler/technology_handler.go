package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LimeOnTop/interverse-gateway/internal/service"
)

type TechnologyHandler struct {
	technologyService *service.TechnologyService
}

func NewTechnologyHandler(technologyService *service.TechnologyService) *TechnologyHandler {
	return &TechnologyHandler{
		technologyService: technologyService,
	}
}

func (h *TechnologyHandler) GetTechnologies(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	response, err := h.technologyService.GetTechnologies(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TechnologyHandler) GetTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "technology ID is required"})
		return
	}

	response, err := h.technologyService.GetTechnology(technologyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TechnologyHandler) CreateTechnology(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Category    string   `json:"category" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.technologyService.CreateTechnology(req.Name, req.Category, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *TechnologyHandler) UpdateTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "technology ID is required"})
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.technologyService.UpdateTechnology(technologyID, req.Name, req.Category, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TechnologyHandler) DeleteTechnology(c *gin.Context) {
	technologyID := c.Param("id")
	if technologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "technology ID is required"})
		return
	}

	response, err := h.technologyService.DeleteTechnology(technologyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TechnologyHandler) SearchTechnologies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	response, err := h.technologyService.SearchTechnologies(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
