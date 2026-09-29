package apperr

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DevMode controls whether clients receive detailed internal error strings.
// Default false (production-safe). Set via Configure from config.DEV_MODE.
var DevMode bool

const (
	MsgInvalidRequest     = "invalid request"
	MsgInternal           = "internal server error"
	MsgUnauthorized       = "unauthorized"
	MsgForbidden          = "forbidden"
	MsgNotFound           = "not found"
	MsgServiceUnavailable = "service temporarily unavailable"
)

func Configure(devMode bool) {
	DevMode = devMode
}

// Public responds with a known safe client message (no sanitization).
func Public(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// Bind responds to request validation/binding failures.
func Bind(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if DevMode {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": MsgInvalidRequest})
}

// Internal logs the real error and hides details unless DevMode is on.
func Internal(c *gin.Context, err error) {
	if err == nil {
		return
	}
	log.Printf("api error: %v", err)
	msg := MsgInternal
	if DevMode {
		msg = err.Error()
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}

// Upstream maps a microservice Response.Error (or similar) for the client.
// In DevMode the original string is returned; otherwise internal-looking
// messages are replaced with a generic user-facing error.
func Upstream(c *gin.Context, status int, message string) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = MsgInternal
	}
	if !DevMode && looksInternal(msg) {
		log.Printf("api upstream error (sanitized): %s", msg)
		switch {
		case status == http.StatusUnauthorized:
			msg = MsgUnauthorized
		case status == http.StatusForbidden:
			msg = MsgForbidden
		case status == http.StatusNotFound:
			msg = MsgNotFound
		case status >= 500:
			msg = MsgInternal
		default:
			msg = MsgInvalidRequest
		}
	} else if !DevMode {
		// Keep business messages; still log for ops visibility.
		log.Printf("api upstream error: %s", msg)
	}
	c.JSON(status, gin.H{"error": msg})
}

func looksInternal(msg string) bool {
	lower := strings.ToLower(msg)
	markers := []string{
		"pq:",
		"sql:",
		"redis",
		"rpc error",
		"connection refused",
		"dial tcp",
		"i/o timeout",
		"context deadline",
		"context canceled",
		"driver:",
		"kafka",
		"elasticsearch",
		"begin tx",
		"commit tx",
		"wrap:",
		"panic:",
		"stack",
		"goroutine",
		"/users/",
		"/home/",
		"password=",
		"postgres://",
		"get interview:",
		"get session content:",
		"get session items:",
		"load session:",
		"start session:",
		"complete interview:",
		"generate report:",
		"replace fallback",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}
