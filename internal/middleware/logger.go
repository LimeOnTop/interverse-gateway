package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			strings.SplitN(param.Path, "?", 2)[0],
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}
