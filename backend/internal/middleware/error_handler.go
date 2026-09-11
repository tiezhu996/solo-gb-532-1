package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

func ErrorHandler(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		for _, ginErr := range c.Errors {
			attributes := []any{"request_id", c.GetString("request_id"), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration_ms", time.Since(started).Milliseconds()}
			if appErr, ok := ginErr.Err.(*api.AppError); ok && appErr.Status < 500 {
				log.Warn("business request rejected", append(attributes, "code", appErr.Code)...)
				continue
			}
			log.Error("request failed", append(attributes, "error", ginErr.Err.Error())...)
		}
	}
}
