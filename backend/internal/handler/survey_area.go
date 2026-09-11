package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type SurveyAreaHandler struct{ service *service.SurveyAreaService }

func NewSurveyAreaHandler(service *service.SurveyAreaService) *SurveyAreaHandler {
	return &SurveyAreaHandler{service: service}
}

func (h *SurveyAreaHandler) List(c *gin.Context) {
	page, size := pageQuery(c)
	items, total, err := h.service.List(dto.SurveyAreaQuery{Status: cleanQuery(c, "status"), Search: cleanQuery(c, "search"), Page: page, PageSize: size})
	if err != nil {
		writeServiceError(c, err, "测区")
		return
	}
	api.Page(c, items, page, size, total)
}

func (h *SurveyAreaHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		writeServiceError(c, err, "测区")
		return
	}
	api.Success(c, http.StatusOK, item)
}

func (h *SurveyAreaHandler) Create(c *gin.Context) {
	var request dto.CreateSurveyAreaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Create(request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测区")
		return
	}
	api.Success(c, http.StatusCreated, item)
}

func (h *SurveyAreaHandler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var request dto.UpdateSurveyAreaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	item, err := h.service.Update(id, request, actorFrom(c))
	if err != nil {
		writeServiceError(c, err, "测区")
		return
	}
	api.Success(c, http.StatusOK, item)
}
