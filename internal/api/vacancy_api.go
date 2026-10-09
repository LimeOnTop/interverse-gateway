package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type VacancyAPI struct {
	vacancyClient usecase.VacancyGateway
}

func NewVacancyAPI(vacancyClient usecase.VacancyGateway) *VacancyAPI {
	return &VacancyAPI{vacancyClient: vacancyClient}
}

func (a *VacancyAPI) GetVacancies(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	page := queryInt32(c, "page", 0)
	perPage := queryInt32(c, "per_page", 20)
	area := strings.TrimSpace(c.Query("area"))
	sources := splitCSV(c.Query("sources"))

	resp, err := a.vacancyClient.GetVacanciesForUser(
		c.Request.Context(),
		user.ID,
		page,
		perPage,
		area,
		sources,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile_ready": resp.ProfileReady,
		"query_text":    resp.QueryText,
		"items":         resp.Items,
		"page":          resp.Page,
		"per_page":      resp.PerPage,
		"total":         resp.Total,
		"sources":       resp.Sources,
		"errors":        resp.Errors,
	})
}

func queryInt32(c *gin.Context, key string, fallback int32) int32 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return int32(n)
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
