package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/geometry"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type SurveyAreaView struct {
	model.SurveyArea
	Summary dto.AreaSummary `json:"summary"`
}

type SurveyAreaService struct {
	repository *repository.SurveyAreaRepository
	audit      *AuditService
}

func NewSurveyAreaService(repository *repository.SurveyAreaRepository, audit *AuditService) *SurveyAreaService {
	return &SurveyAreaService{repository: repository, audit: audit}
}

func (s *SurveyAreaService) List(query dto.SurveyAreaQuery) ([]SurveyAreaView, int64, error) {
	items, total, err := s.repository.List(query)
	if err != nil {
		return nil, 0, err
	}
	views := make([]SurveyAreaView, 0, len(items))
	for _, item := range items {
		summary, summaryErr := s.repository.Summary(item.ID)
		if summaryErr != nil {
			return nil, 0, fmt.Errorf("summarize area %d: %w", item.ID, summaryErr)
		}
		views = append(views, SurveyAreaView{SurveyArea: item, Summary: summary})
	}
	return views, total, nil
}

func (s *SurveyAreaService) Get(id uint) (SurveyAreaView, error) {
	item, err := s.repository.Get(id)
	if err != nil {
		return SurveyAreaView{}, mapDatabaseError(err, "测区")
	}
	summary, err := s.repository.Summary(id)
	if err != nil {
		return SurveyAreaView{}, err
	}
	return SurveyAreaView{SurveyArea: item, Summary: summary}, nil
}

func (s *SurveyAreaService) Create(request dto.CreateSurveyAreaRequest, actor Actor) (SurveyAreaView, error) {
	if err := validateAreaGeometry(request.BoundaryGeoJSON, request.CoordinateSystem); err != nil {
		return SurveyAreaView{}, err
	}
	item := model.SurveyArea{AreaCode: strings.ToUpper(strings.TrimSpace(request.AreaCode)), Name: strings.TrimSpace(request.Name), BoundaryGeoJSON: datatypes.JSON(request.BoundaryGeoJSON), TargetResolutionM: request.TargetResolutionM, CoordinateSystem: strings.ToUpper(strings.TrimSpace(request.CoordinateSystem)), DefaultSwathM: request.DefaultSwathM, OwnerTeam: strings.TrimSpace(request.OwnerTeam), Status: constants.AreaDraft, Version: 1}
	if err := s.repository.Create(&item); err != nil {
		return SurveyAreaView{}, mapDatabaseError(err, "测区编号")
	}
	if err := s.audit.Record(actor, "area.create", "survey_area", item.ID, nil, item, map[string]any{"coordinate_system": item.CoordinateSystem}); err != nil {
		return SurveyAreaView{}, err
	}
	return s.Get(item.ID)
}

func (s *SurveyAreaService) Update(id uint, request dto.UpdateSurveyAreaRequest, actor Actor) (SurveyAreaView, error) {
	before, err := s.repository.Get(id)
	if err != nil {
		return SurveyAreaView{}, mapDatabaseError(err, "测区")
	}
	if before.Status == constants.AreaArchived {
		return SurveyAreaView{}, api.Conflict("AREA_ARCHIVED", "已归档测区不可修改", nil)
	}
	if request.Status != constants.AreaDraft && request.Status != constants.AreaActive && request.Status != constants.AreaArchived {
		return SurveyAreaView{}, api.Unprocessable("AREA_STATUS_INVALID", "测区状态无效", nil)
	}
	if err := validateAreaGeometry(request.BoundaryGeoJSON, request.CoordinateSystem); err != nil {
		return SurveyAreaView{}, err
	}
	updated, err := s.repository.Update(id, request.ExpectedVersion, map[string]any{"name": strings.TrimSpace(request.Name), "boundary_geojson": request.BoundaryGeoJSON, "target_resolution_m": request.TargetResolutionM, "coordinate_system": strings.ToUpper(strings.TrimSpace(request.CoordinateSystem)), "default_swath_m": request.DefaultSwathM, "owner_team": strings.TrimSpace(request.OwnerTeam), "status": request.Status})
	if err != nil {
		return SurveyAreaView{}, mapDatabaseError(err, "测区")
	}
	if err := s.audit.Record(actor, "area.update", "survey_area", id, before, updated, map[string]any{"expected_version": request.ExpectedVersion}); err != nil {
		return SurveyAreaView{}, err
	}
	return s.Get(id)
}

func validateAreaGeometry(data []byte, crs string) error {
	if err := geometry.ValidateProjectedCRS(crs); err != nil {
		return api.Unprocessable("COORDINATE_SYSTEM_INVALID", "坐标系必须是以米为单位的投影坐标系", err)
	}
	polygon, err := geometry.ParsePolygon(data)
	if err != nil {
		return api.Unprocessable("GEOJSON_INVALID", "测区边界 GeoJSON 无效", err)
	}
	if geometry.PolygonArea(polygon) > 1_000_000_000 {
		return api.Unprocessable("AREA_TOO_LARGE", "测区面积超出离线分析上限", nil)
	}
	return nil
}

func mapDatabaseError(err error, entity string) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return api.NotFound(entity, err)
	case errors.Is(err, repository.ErrVersionConflict):
		return api.Conflict("VERSION_CONFLICT", entity+"版本已变化，请刷新后重试", err)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return api.Conflict("DUPLICATE_RESOURCE", entity+"已存在", err)
	default:
		return err
	}
}
