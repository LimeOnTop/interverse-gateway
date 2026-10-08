package api

import (
	"net/http"
	"strconv"
	"strings"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

const (
	supportRoleUser  = "user"
	supportRoleAdmin = "admin"

	supportMessagesPageSize = 50
	supportTicketsPageSize  = 20
)

// SupportAPI serves both the user's own tickets and the admin inbox.
// The admin flag decides whose tickets are visible and who is the message author.
type SupportAPI struct {
	supportClient *clients.SupportClient
	admin         bool
}

func NewSupportAPI(supportClient *clients.SupportClient) *SupportAPI {
	return &SupportAPI{supportClient: supportClient}
}

func NewAdminSupportAPI(supportClient *clients.SupportClient) *SupportAPI {
	return &SupportAPI{supportClient: supportClient, admin: true}
}

// scope returns the user id tickets are filtered by: empty for the admin inbox.
func (a *SupportAPI) scope(c *gin.Context) (middleware.User, string, bool) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return middleware.User{}, "", false
	}
	if a.admin {
		return user, "", true
	}
	if middleware.IsAdmin(user) {
		apperr.Public(c, http.StatusForbidden, "администратор отвечает на обращения в админ-панели")
		return middleware.User{}, "", false
	}
	return user, user.ID, true
}

func (a *SupportAPI) role() string {
	if a.admin {
		return supportRoleAdmin
	}
	return supportRoleUser
}

func ticketIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		apperr.Public(c, http.StatusBadRequest, "некорректный номер обращения")
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, fallback int32) int32 {
	v, err := strconv.ParseInt(c.Query(key), 10, 32)
	if err != nil || v < 0 {
		return fallback
	}
	return int32(v)
}

// supportFailed writes the upstream error and reports whether the response was a failure.
func supportFailed(c *gin.Context, resp *pb.Response) bool {
	if resp == nil || resp.GetSuccess() {
		return false
	}
	msg := resp.GetError()
	status := http.StatusBadRequest
	switch {
	case strings.Contains(msg, "не найдено"):
		status = http.StatusNotFound
	case strings.Contains(msg, "закрыто"):
		status = http.StatusConflict
	}
	apperr.Upstream(c, status, msg)
	return true
}

func (a *SupportAPI) CreateTicket(c *gin.Context) {
	user, userID, ok := a.scope(c)
	if !ok {
		return
	}
	var req struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.supportClient.CreateTicket(c.Request.Context(), userID, user.Email, req.Subject, req.Message)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ticket": resp.GetTicket()})
}

func (a *SupportAPI) ListTickets(c *gin.Context) {
	_, userID, ok := a.scope(c)
	if !ok {
		return
	}
	limit := queryInt(c, "limit", supportTicketsPageSize)
	offset := queryInt(c, "offset", 0)

	resp, err := a.supportClient.ListTickets(c.Request.Context(), userID, strings.TrimSpace(c.Query("status")), limit, offset)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	tickets := resp.GetTickets()
	if tickets == nil {
		tickets = []*pb.SupportTicket{}
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets, "total": resp.GetTotal()})
}

func (a *SupportAPI) GetTicket(c *gin.Context) {
	_, userID, ok := a.scope(c)
	if !ok {
		return
	}
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	resp, err := a.supportClient.GetTicket(c.Request.Context(), ticketID, userID)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": resp.GetTicket()})
}

// ListMessages pages from the newest message: offset=0 is the latest page,
// offset=50 the 50 messages before it, and so on.
func (a *SupportAPI) ListMessages(c *gin.Context) {
	_, userID, ok := a.scope(c)
	if !ok {
		return
	}
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	limit := queryInt(c, "limit", supportMessagesPageSize)
	offset := queryInt(c, "offset", 0)

	resp, err := a.supportClient.ListMessages(c.Request.Context(), ticketID, userID, limit, offset)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	messages := resp.GetMessages()
	if messages == nil {
		messages = []*pb.SupportMessage{}
	}
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    resp.GetTotal(),
		"has_more": resp.GetHasMore(),
	})
}

func (a *SupportAPI) AddMessage(c *gin.Context) {
	user, userID, ok := a.scope(c)
	if !ok {
		return
	}
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}

	resp, err := a.supportClient.AddMessage(c.Request.Context(), ticketID, userID, a.role(), user.ID, req.Body)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": resp.GetMessage(), "ticket": resp.GetTicket()})
}

func (a *SupportAPI) CloseTicket(c *gin.Context) {
	_, userID, ok := a.scope(c)
	if !ok {
		return
	}
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	resp, err := a.supportClient.CloseTicket(c.Request.Context(), ticketID, userID, a.role())
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if supportFailed(c, resp.GetResponse()) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": resp.GetTicket()})
}
