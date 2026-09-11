package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			api.WriteError(c, api.Unauthorized("请先登录后访问"))
			return
		}
		claims, err := authService.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			api.WriteError(c, err)
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("display_name", claims.DisplayName)
		c.Set("role", claims.Role)
		c.Next()
	}
}
