package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type SonarRunHandler struct{ service *service.SonarRunService }

func NewSonarRunHandler(service *service.SonarRunService) *SonarRunHandler {
	return &SonarRunHandler{service: service}
}
func (h *SonarRunHandler) List(c *gin.Context) {
	page, size := pageQuery(c)
	items, total, err := h.service.List(dto.SonarRunQuery{TransectPlanID: uintQuery(c, "transect_plan_id"), State: cleanQuery(c, "state"), Quality: cleanQuery(c, "quality"), Page: page, PageSize: size})
	if err != nil {
		writeServiceError(c, err, "声呐运行")
		return
	}
	api.Page(c, items, page, size, total)
}
func (h *SonarRunHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		writeServiceError(c, err, "声呐运行")
		return
	}
	api.Success(c, http.StatusOK, item)
}
func (h *SonarRunHandler) Import(c *gin.Context) {
	var request dto.ImportSonarRunRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, idempotent, err := h.service.Import(request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "声呐运行")
		return
	}
	status := http.StatusCreated
	if idempotent {
		status = http.StatusOK
	}
	api.Success(c, status, gin.H{"run": item, "idempotent": idempotent})
}
func (h *SonarRunHandler) Quality(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	evidence, err := h.service.Quality(id)
	if err != nil {
		writeServiceError(c, err, "声呐运行")
		return
	}
	api.Success(c, http.StatusOK, evidence)
}
func (h *SonarRunHandler) Transition(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var request dto.RunTransitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Transition(id, request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "声呐运行")
		return
	}
	api.Success(c, http.StatusOK, item)
}
