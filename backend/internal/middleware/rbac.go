package middleware

import (
	"github.com/gin-gonic/gin"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

func RBAC(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		role := c.GetString("role")
		if _, ok := allowed[role]; !ok {
			api.WriteError(c, api.Forbidden("当前角色无权执行此操作"))
			return
		}
		c.Next()
	}
}
