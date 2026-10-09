package api

import (
	"net/http"

	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ReportAPI struct {
	reportClient usecase.ReportGateway
	authClient   usecase.AuthGateway
}

func NewReportAPI(reportClient usecase.ReportGateway, authClient usecase.AuthGateway) *ReportAPI {
	return &ReportAPI{
		reportClient: reportClient,
		authClient:   authClient,
	}
}

// fullReportAccess tells whether weak points and answer reviews may be shown:
// they are a Pro feature, so Basic users get only the count.
func (a *ReportAPI) fullReportAccess(c *gin.Context, user middleware.User) bool {
	return a.reportPlan(c, user) == planPaid
}

// reportPlan is the plan reports are built and shown for; admins count as Pro.
func (a *ReportAPI) reportPlan(c *gin.Context, user middleware.User) string {
	if middleware.IsAdmin(user) {
		return planPaid
	}
	return subscriptionPlan(c.Request.Context(), a.authClient, user.ID)
}

func (a *ReportAPI) CreateReport(c *gin.Context) {
	var req struct {
		InterviewID         string `json:"interview_id" binding:"required"`
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
		apperr.Bind(c, err)
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.reportClient.CreateReport(
		c.Request.Context(),
		req.InterviewID, user.ID, req.OverallRating,
		req.TechnicalSkills, req.CommunicationSkills, req.ProblemSolving,
		req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
	)
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

func (a *ReportAPI) GenerateReport(c *gin.Context) {
	var req struct {
		InterviewID string                `json:"interview_id" binding:"required"`
		Answers     sessionAnswersRequest `json:"answers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	plan := a.reportPlan(c, user)
	resp, err := a.reportClient.GenerateReport(
		c.Request.Context(),
		req.InterviewID,
		user.ID,
		toSessionAnswerInputs(req.Answers),
		plan,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	full := plan == planPaid
	reportMap := protoReportToMap(resp.GetReport())
	c.JSON(http.StatusCreated, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(reportMap, full),
		"scores":   visibleScores(resp.GetScores(), full),
	})
}

func (a *ReportAPI) AnalyzeReport(c *gin.Context) {
	reportResp, user, ok := a.loadOwnedReport(c)
	if !ok {
		return
	}

	report := reportResp.GetReport()
	if report == nil {
		apperr.Public(c, http.StatusNotFound, "report not found")
		return
	}

	plan := a.reportPlan(c, user)
	full := plan == planPaid
	mapped := mapReportResponse(protoReportToMap(report), full)
	if analyzed, ok := mapped["ai_analyzed"].(bool); ok && analyzed {
		c.JSON(http.StatusOK, gin.H{
			"response": gin.H{"success": true, "message": "Report already analyzed"},
			"report":   mapped,
		})
		return
	}

	var req struct {
		Answers sessionAnswersRequest `json:"answers"`
	}
	_ = c.ShouldBindJSON(&req)

	resp, err := a.reportClient.GenerateReport(
		c.Request.Context(),
		report.GetInterviewId(),
		user.ID,
		toSessionAnswerInputs(req.Answers),
		plan,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	reportMap := protoReportToMap(resp.GetReport())
	c.JSON(http.StatusOK, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(reportMap, full),
		"scores":   visibleScores(resp.GetScores(), full),
	})
}

type sessionAnswerPayload struct {
	StepID              string `json:"step_id" binding:"required"`
	QuestionID          string `json:"question_id"`
	ItemType            string `json:"item_type" binding:"required"`
	SelectedOptionIndex *int32 `json:"selected_option_index"`
	TaskAnswer          string `json:"task_answer"`
}

type sessionAnswersRequest []sessionAnswerPayload

func toSessionAnswerInputs(answers sessionAnswersRequest) []*pb.SessionAnswerInput {
	result := make([]*pb.SessionAnswerInput, 0, len(answers))
	for _, answer := range answers {
		input := &pb.SessionAnswerInput{
			StepId:     answer.StepID,
			QuestionId: answer.QuestionID,
			ItemType:   answer.ItemType,
			TaskAnswer: answer.TaskAnswer,
		}
		if answer.SelectedOptionIndex != nil {
			input.SelectedOptionIndex = *answer.SelectedOptionIndex
		}
		result = append(result, input)
	}
	return result
}

func (a *ReportAPI) GetReports(c *gin.Context) {
	page, limit := parsePagination(c)

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.reportClient.GetReports(c.Request.Context(), user.ID, page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}

	full := a.fullReportAccess(c, user)
	reports := make([]map[string]any, 0, len(resp.GetReports()))
	for _, report := range resp.GetReports() {
		reports = append(reports, mapReportResponse(protoReportToMap(report), full))
	}

	c.JSON(http.StatusOK, gin.H{
		"response":   resp.GetResponse(),
		"reports":    reports,
		"pagination": resp.GetPagination(),
	})
}

// loadOwnedReport fetches the report and makes sure it belongs to the current
// user (admins may access any report). Foreign reports are reported as not found.
func (a *ReportAPI) loadOwnedReport(c *gin.Context) (*pb.GetReportResponse, middleware.User, bool) {
	reportID := c.Param("id")
	if reportID == "" {
		apperr.Public(c, http.StatusBadRequest, "report ID is required")
		return nil, middleware.User{}, false
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return nil, middleware.User{}, false
	}

	resp, err := a.reportClient.GetReport(c.Request.Context(), reportID)
	if err != nil {
		apperr.Internal(c, err)
		return nil, middleware.User{}, false
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return nil, middleware.User{}, false
	}
	if resp.GetReport() == nil || (resp.GetReport().GetUserId() != user.ID && !middleware.IsAdmin(user)) {
		apperr.Public(c, http.StatusNotFound, "report not found")
		return nil, middleware.User{}, false
	}

	return resp, user, true
}

func (a *ReportAPI) GetReport(c *gin.Context) {
	resp, user, ok := a.loadOwnedReport(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(protoReportToMap(resp.GetReport()), a.fullReportAccess(c, user)),
	})
}

func (a *ReportAPI) UpdateReport(c *gin.Context) {
	_, user, ok := a.loadOwnedReport(c)
	if !ok {
		return
	}
	reportID := c.Param("id")

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
		apperr.Bind(c, err)
		return
	}

	resp, err := a.reportClient.UpdateReport(
		c.Request.Context(),
		reportID, req.OverallRating, req.TechnicalSkills, req.CommunicationSkills,
		req.ProblemSolving, req.Strengths, req.Weaknesses, req.Recommendations, req.Notes,
	)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	// Never echo raw notes: they carry Pro-only weak points.
	c.JSON(http.StatusOK, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(protoReportToMap(resp.GetReport()), a.fullReportAccess(c, user)),
	})
}

func (a *ReportAPI) DeleteReport(c *gin.Context) {
	if _, _, ok := a.loadOwnedReport(c); !ok {
		return
	}
	reportID := c.Param("id")

	resp, err := a.reportClient.DeleteReport(c.Request.Context(), reportID)
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

// visibleScores drops the AI comments for Basic users: they name weak topics.
func visibleScores(scores *pb.AnalysisScores, full bool) *pb.AnalysisScores {
	if scores == nil || full {
		return scores
	}
	return &pb.AnalysisScores{
		OverallScore:      scores.GetOverallScore(),
		AlgorithmScore:    scores.GetAlgorithmScore(),
		ArchitectureScore: scores.GetArchitectureScore(),
		CodingScore:       scores.GetCodingScore(),
		SoftSkillsScore:   scores.GetSoftSkillsScore(),
	}
}
