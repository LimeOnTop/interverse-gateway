package api

import (
	pb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func (a *AdminAPI) ListUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 1000000 {
		apperr.Public(c, http.StatusBadRequest, "Некорректная страница")
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if err != nil || limit < 1 || limit > 100 {
		apperr.Public(c, http.StatusBadRequest, "Некорректный размер страницы")
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if utf8.RuneCountInString(query) > 100 {
		apperr.Public(c, http.StatusBadRequest, "Поисковый запрос должен быть не длиннее 100 символов")
		return
	}
	day, week, month := metricPeriods(time.Now())
	var since int64
	switch c.DefaultQuery("period", "total") {
	case "day":
		since = day.Unix()
	case "week":
		since = week.Unix()
	case "month":
		since = month.Unix()
	case "total":
	default:
		apperr.Public(c, http.StatusBadRequest, "Некорректный период")
		return
	}
	resp, err := a.authClient.ListUsers(c.Request.Context(), &pb.ListUsersRequest{Query: query, RegisteredSinceUnix: since, Page: int32(page), Limit: int32(limit)})
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.GetResponse() == nil || !resp.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadGateway, "Не удалось загрузить пользователей")
		return
	}
	// Hand-build public DTOs so protobuf additions can never expose credentials.
	users := make([]gin.H, 0, len(resp.GetUsers()))
	for _, u := range resp.GetUsers() {
		users = append(users, gin.H{"id": u.GetId(), "name": u.GetName(), "email": u.GetEmail(), "created_at": u.GetCreatedAt()})
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"users": users, "pagination": gin.H{"total": resp.GetTotal(), "page": resp.GetPage(), "limit": resp.GetLimit()}, "period": c.DefaultQuery("period", "total")})
}
