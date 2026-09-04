package api

import (
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
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

func (a *ReportAPI) GenerateReport(c *gin.Context) {
	var req struct {
		InterviewID string `json:"interview_id" binding:"required"`
		Answers     sessionAnswersRequest `json:"answers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	resp, err := a.reportClient.GenerateReport(
		c.Request.Context(),
		req.InterviewID,
		user.ID,
		toSessionAnswerInputs(req.Answers),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}

	reportMap := protoReportToMap(resp.GetReport())
	c.JSON(http.StatusCreated, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(reportMap),
		"scores":   resp.GetScores(),
	})
}

func (a *ReportAPI) AnalyzeReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report ID is required"})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}

	reportResp, err := a.reportClient.GetReport(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if reportResp.Response != nil && !reportResp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": reportResp.Response.Error})
		return
	}

	report := reportResp.GetReport()
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	mapped := mapReportResponse(protoReportToMap(report))
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
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}

	reportMap := protoReportToMap(resp.GetReport())
	c.JSON(http.StatusOK, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(reportMap),
		"scores":   resp.GetScores(),
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Response.Error})
		return
	}

	reports := make([]map[string]any, 0, len(resp.GetReports()))
	for _, report := range resp.GetReports() {
		reports = append(reports, mapReportResponse(protoReportToMap(report)))
	}

	c.JSON(http.StatusOK, gin.H{
		"response":   resp.GetResponse(),
		"reports":    reports,
		"pagination": resp.GetPagination(),
	})
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

	c.JSON(http.StatusOK, gin.H{
		"response": resp.GetResponse(),
		"report":   mapReportResponse(protoReportToMap(resp.GetReport())),
	})
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
