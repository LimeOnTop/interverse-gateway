package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parsePagination(c *gin.Context) (page, limit int32) {
	page = 1
	limit = 10
	if p, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32); err == nil && p > 0 {
		page = int32(p)
	}
	if l, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32); err == nil && l > 0 {
		limit = int32(l)
	}
	return page, limit
}
