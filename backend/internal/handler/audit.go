package handler

import (
	"github.com/gin-gonic/gin"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type AuditHandler struct{ service *service.AuditService }

func NewAuditHandler(service *service.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}
func (h *AuditHandler) List(c *gin.Context) {
	page, size := pageQuery(c)
	items, total, err := h.service.List(repository.AuditFilter{RequestID: cleanQuery(c, "request_id"), EntityType: cleanQuery(c, "entity_type"), Actor: cleanQuery(c, "actor"), Page: page, PageSize: size})
	if err != nil {
		writeServiceError(c, err, "审计记录")
		return
	}
	api.Page(c, items, page, size, total)
}
