package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
	"strings"
)

type CoverageGapHandler struct{ service *service.CoverageGapService }

func NewCoverageGapHandler(service *service.CoverageGapService) *CoverageGapHandler {
	return &CoverageGapHandler{service: service}
}
func (h *CoverageGapHandler) List(c *gin.Context) {
	page, size := pageQuery(c)
	items, total, err := h.service.List(dto.CoverageGapQuery{SurveyAreaID: uintQuery(c, "survey_area_id"), State: cleanQuery(c, "state"), Severity: cleanQuery(c, "severity"), Page: page, PageSize: size})
	if err != nil {
		writeServiceError(c, err, "覆盖缺口")
		return
	}
	api.Page(c, items, page, size, total)
}
func (h *CoverageGapHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		writeServiceError(c, err, "覆盖缺口")
		return
	}
	api.Success(c, http.StatusOK, item)
}
func (h *CoverageGapHandler) Detect(c *gin.Context) {
	var request dto.DetectCoverageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Detect(request, strings.TrimSpace(c.GetHeader("Idempotency-Key")), actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "覆盖缺口")
		return
	}
	status := http.StatusCreated
	if item.Idempotent {
		status = http.StatusOK
	}
	api.Success(c, status, item)
}
func (h *CoverageGapHandler) Transition(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var request dto.GapTransitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Transition(id, request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "覆盖缺口")
		return
	}
	api.Success(c, http.StatusOK, item)
}
