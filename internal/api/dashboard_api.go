package api

import (
	"log"
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type DashboardAPI struct {
	interviewClient *clients.InterviewClient
	reportClient    *clients.ReportClient
	authClient      *clients.AuthClient
}

func NewDashboardAPI(interviewClient *clients.InterviewClient, reportClient *clients.ReportClient, authClient *clients.AuthClient) *DashboardAPI {
	return &DashboardAPI{interviewClient: interviewClient, reportClient: reportClient, authClient: authClient}
}

// Summary returns everything the dashboard and progress pages show: plan
// quota as the server enforces it, training counts and the latest report.
func (a *DashboardAPI) Summary(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	plan := planPaid
	if !middleware.IsAdmin(user) {
		plan = subscriptionPlan(ctx, a.authClient, user.ID)
	}
	full := plan == planPaid

	stats, err := a.interviewClient.GetTrainingStats(ctx, user.ID, plan)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if stats.GetResponse() != nil && !stats.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadRequest, stats.GetResponse().GetError())
		return
	}

	var lastReport map[string]any
	reports, err := a.reportClient.GetReports(ctx, user.ID, 1, 1)
	if err != nil {
		// The dashboard still works without the report block.
		log.Printf("dashboard: load last report: %v", err)
	} else if list := reports.GetReports(); len(list) > 0 {
		lastReport = mapReportResponse(protoReportToMap(list[0]), full)
	}

	remaining := stats.GetQuotaLimit() - stats.GetQuotaUsed()
	if remaining < 0 {
		remaining = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
		"quota": gin.H{
			"used":      stats.GetQuotaUsed(),
			"limit":     stats.GetQuotaLimit(),
			"remaining": remaining,
			"period":    stats.GetQuotaPeriod(),
			"resets_at": stats.GetQuotaResetsAt(),
		},
		"trainings": gin.H{
			"total":       stats.GetTotal(),
			"completed":   stats.GetCompleted(),
			"in_progress": stats.GetInProgress(),
			"scheduled":   stats.GetScheduled(),
		},
		"last_report": lastReport,
	})
}
