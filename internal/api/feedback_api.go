package api

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type FeedbackAPI struct {
	client pb.FeedbackServiceClient
}

func NewFeedbackAPI(client pb.FeedbackServiceClient) *FeedbackAPI {
	return &FeedbackAPI{client: client}
}

func (a *FeedbackAPI) Submit(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || utf8.RuneCountInString(message) > 3000 || strings.ContainsRune(message, 0) {
		apperr.Public(c, http.StatusBadRequest, "Сообщение должно содержать от 1 до 3000 символов.")
		return
	}
	_, err := a.client.SubmitFeedback(c.Request.Context(), &pb.SubmitFeedbackRequest{
		UserId: user.ID, UserEmail: user.Email, Message: message,
	})
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (a *FeedbackAPI) List(c *gin.Context) {
	beforeID, err := strconv.ParseInt(c.DefaultQuery("before_id", "0"), 10, 64)
	if err != nil || beforeID < 0 {
		apperr.Public(c, http.StatusBadRequest, "Некорректная страница.")
		return
	}
	result, err := a.client.ListFeedback(c.Request.Context(), &pb.ListFeedbackRequest{BeforeId: beforeID})
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	entries := result.GetEntries()
	if entries == nil {
		entries = []*pb.FeedbackEntry{}
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"entries": entries, "next_before_id": result.GetNextBeforeId()})
}
