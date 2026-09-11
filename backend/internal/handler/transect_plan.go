package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type TransectPlanHandler struct{ service *service.TransectPlanService }

func NewTransectPlanHandler(service *service.TransectPlanService) *TransectPlanHandler {
	return &TransectPlanHandler{service: service}
}

func (h *TransectPlanHandler) List(c *gin.Context) {
	page, size := pageQuery(c)
	items, total, err := h.service.List(dto.TransectPlanQuery{SurveyAreaID: uintQuery(c, "survey_area_id"), State: cleanQuery(c, "state"), Page: page, PageSize: size})
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Page(c, items, page, size, total)
}
func (h *TransectPlanHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusOK, item)
}
func (h *TransectPlanHandler) Create(c *gin.Context) {
	var request dto.CreateTransectPlanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Create(request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusCreated, item)
}
func (h *TransectPlanHandler) Generate(c *gin.Context) {
	var request dto.GenerateLinesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Generate(request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusCreated, item)
}
func (h *TransectPlanHandler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var request dto.UpdateTransectPlanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Update(id, request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusOK, item)
}
func (h *TransectPlanHandler) Lock(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var request dto.PlanTransitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Lock(id, request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusOK, item)
}
func (h *TransectPlanHandler) Copy(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := h.service.Copy(id, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测线规划")
		return
	}
	api.Success(c, http.StatusCreated, item)
}
