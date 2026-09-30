package api

import (
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"net/http"

	pb "github.com/LimeOnTop/interverse-contracts/question/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type questionOptionRequest struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int32  `json:"sort_order"`
}

type QuestionAPI struct {
	questionClient *clients.QuestionClient
}

func NewQuestionAPI(questionClient *clients.QuestionClient) *QuestionAPI {
	return &QuestionAPI{
		questionClient: questionClient,
	}
}

func (a *QuestionAPI) GetQuestions(c *gin.Context) {
	page, limit := parsePagination(c)

	resp, err := a.questionClient.GetQuestions(c.Request.Context(), page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	redactQuestions(c, resp.Questions)
	c.JSON(http.StatusOK, resp)
}

func (a *QuestionAPI) GetQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	resp, err := a.questionClient.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	redactQuestions(c, []*pb.Question{resp.Question})
	c.JSON(http.StatusOK, resp)
}

func (a *QuestionAPI) CreateQuestion(c *gin.Context) {
	var req struct {
		Text       string                  `json:"text" binding:"required"`
		Category   string                  `json:"category" binding:"required"`
		Difficulty string                  `json:"difficulty" binding:"required"`
		Technology string                  `json:"technology" binding:"required"`
		Tags       []string                `json:"tags"`
		Answer     string                  `json:"answer" binding:"required"`
		Options    []questionOptionRequest `json:"options"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.questionClient.CreateQuestion(c.Request.Context(), req.Text, req.Category, req.Difficulty, req.Technology, req.Tags, req.Answer, toProtoOptions(req.Options))
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

func (a *QuestionAPI) UpdateQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	var req struct {
		Text       string                  `json:"text"`
		Category   string                  `json:"category"`
		Difficulty string                  `json:"difficulty"`
		Technology string                  `json:"technology"`
		Tags       []string                `json:"tags"`
		Answer     string                  `json:"answer"`
		Options    []questionOptionRequest `json:"options"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.questionClient.UpdateQuestion(c.Request.Context(), questionID, req.Text, req.Category, req.Difficulty, req.Technology, req.Tags, req.Answer, toProtoOptions(req.Options))
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

func (a *QuestionAPI) DeleteQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		apperr.Public(c, http.StatusBadRequest, "question ID is required")
		return
	}

	resp, err := a.questionClient.DeleteQuestion(c.Request.Context(), questionID)
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

func (a *QuestionAPI) SearchQuestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		apperr.Public(c, http.StatusBadRequest, "search query is required")
		return
	}

	page, limit := parsePagination(c)

	resp, err := a.questionClient.SearchQuestions(c.Request.Context(), query, page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	redactQuestions(c, resp.Questions)
	c.JSON(http.StatusOK, resp)
}

func (a *QuestionAPI) GetQuestionsByTechnology(c *gin.Context) {
	technology := c.Param("technology")
	if technology == "" {
		technology = c.Query("technology")
	}

	difficulty := c.Query("difficulty")
	page, limit := parsePagination(c)

	resp, err := a.questionClient.GetQuestionsByTechnology(c.Request.Context(), technology, difficulty, c.Query("category"), page, limit)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	if resp.Response != nil && !resp.Response.Success {
		apperr.Upstream(c, http.StatusBadRequest, resp.Response.Error)
		return
	}
	redactQuestions(c, resp.Questions)
	c.JSON(http.StatusOK, resp)
}

func toProtoOptions(options []questionOptionRequest) []*pb.QuestionOption {
	result := make([]*pb.QuestionOption, 0, len(options))
	for _, option := range options {
		result = append(result, &pb.QuestionOption{
			Text:      option.Text,
			IsCorrect: option.IsCorrect,
			SortOrder: option.SortOrder,
		})
	}
	return result
}

// redactQuestions hides correct-answer markers and reference answers from
// non-admin users so the question bank cannot be used to look up answers.
func redactQuestions(c *gin.Context, questions []*pb.Question) {
	// The questions group is behind AuthRequired, so the user is always present.
	if user, ok := middleware.CurrentUser(c); ok && middleware.IsAdmin(user) {
		return
	}
	for _, question := range questions {
		if question == nil {
			continue
		}
		question.Answer = ""
		for _, option := range question.Options {
			option.IsCorrect = false
		}
	}
}
