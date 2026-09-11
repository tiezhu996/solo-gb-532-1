package service

import (
	"fmt"
	"math"
	"time"

	"github.com/paulmach/orb"
	"gorm.io/datatypes"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/geometry"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type TransectPlanService struct {
	repository *repository.TransectPlanRepository
	areas      *repository.SurveyAreaRepository
	audit      *AuditService
}

func NewTransectPlanService(repository *repository.TransectPlanRepository, areas *repository.SurveyAreaRepository, audit *AuditService) *TransectPlanService {
	return &TransectPlanService{repository: repository, areas: areas, audit: audit}
}

func (s *TransectPlanService) List(query dto.TransectPlanQuery) ([]model.TransectPlan, int64, error) {
	return s.repository.List(query)
}
func (s *TransectPlanService) Get(id uint) (model.TransectPlan, error) {
	item, err := s.repository.Get(id)
	if err != nil {
		return item, mapDatabaseError(err, "测线规划")
	}
	return item, nil
}

func (s *TransectPlanService) Create(request dto.CreateTransectPlanRequest, actor Actor) (model.TransectPlan, error) {
	area, err := s.areas.Get(request.SurveyAreaID)
	if err != nil {
		return model.TransectPlan{}, mapDatabaseError(err, "测区")
	}
	if area.Status == constants.AreaArchived {
		return model.TransectPlan{}, api.Conflict("AREA_ARCHIVED", "归档测区不能新增规划", nil)
	}
	if _, err := geometry.ParseLines(request.LineGeoJSON); err != nil {
		return model.TransectPlan{}, api.Unprocessable("GEOJSON_INVALID", "测线 GeoJSON 无效", err)
	}
	item := model.TransectPlan{SurveyAreaID: request.SurveyAreaID, Name: request.Name, LineGeoJSON: datatypes.JSON(request.LineGeoJSON), PlannedHeading: request.PlannedHeading, PlannedSwathM: request.PlannedSwathM, LineSpacingM: request.LineSpacingM, PlanState: constants.PlanDraft, Version: 1, CreatedBy: actor.UserID}
	if err := s.repository.Create(&item); err != nil {
		return item, err
	}
	if err := s.audit.Record(actor, "plan.create", "transect_plan", item.ID, nil, item, map[string]any{"source": "manual"}); err != nil {
		return item, err
	}
	return s.Get(item.ID)
}

func (s *TransectPlanService) Generate(request dto.GenerateLinesRequest, actor Actor) (model.TransectPlan, error) {
	area, err := s.areas.Get(request.SurveyAreaID)
	if err != nil {
		return model.TransectPlan{}, mapDatabaseError(err, "测区")
	}
	polygon, err := geometry.ParsePolygon(area.BoundaryGeoJSON)
	if err != nil {
		return model.TransectPlan{}, api.Unprocessable("GEOJSON_INVALID", "测区边界无法生成测线", err)
	}
	minX, minY, maxX, maxY := bounds(polygon)
	if request.LineSpacingM > math.Max(maxX-minX, maxY-minY) {
		return model.TransectPlan{}, api.Unprocessable("LINE_SPACING_INVALID", "测线间距超过测区跨度", nil)
	}
	lines := orb.MultiLineString{}
	horizontal := request.Heading >= 45 && request.Heading < 135 || request.Heading >= 225 && request.Heading < 315
	if horizontal {
		for y := minY + request.LineSpacingM/2; y < maxY; y += request.LineSpacingM {
			lines = append(lines, orb.LineString{{minX, y}, {maxX, y}})
		}
	} else {
		for x := minX + request.LineSpacingM/2; x < maxX; x += request.LineSpacingM {
			lines = append(lines, orb.LineString{{x, minY}, {x, maxY}})
		}
	}
	if len(lines) == 0 {
		return model.TransectPlan{}, api.Unprocessable("NO_TRANSECTS", "当前参数无法在测区内生成测线", nil)
	}
	encoded, err := geometry.MarshalFeature(lines)
	if err != nil {
		return model.TransectPlan{}, fmt.Errorf("encode generated transects: %w", err)
	}
	return s.Create(dto.CreateTransectPlanRequest{SurveyAreaID: request.SurveyAreaID, Name: request.Name, LineGeoJSON: encoded, PlannedHeading: request.Heading, PlannedSwathM: request.PlannedSwathM, LineSpacingM: request.LineSpacingM}, actor)
}

func (s *TransectPlanService) Update(id uint, request dto.UpdateTransectPlanRequest, actor Actor) (model.TransectPlan, error) {
	before, err := s.Get(id)
	if err != nil {
		return model.TransectPlan{}, err
	}
	if before.PlanState != constants.PlanDraft {
		return model.TransectPlan{}, api.Conflict("PLAN_LOCKED", "已锁定规划只能复制新版本", nil)
	}
	if _, err := geometry.ParseLines(request.LineGeoJSON); err != nil {
		return model.TransectPlan{}, api.Unprocessable("GEOJSON_INVALID", "测线 GeoJSON 无效", err)
	}
	updated, err := s.repository.Update(id, request.ExpectedVersion, map[string]any{"name": request.Name, "line_geojson": request.LineGeoJSON, "planned_heading": request.PlannedHeading, "planned_swath_m": request.PlannedSwathM, "line_spacing_m": request.LineSpacingM})
	if err != nil {
		return model.TransectPlan{}, mapDatabaseError(err, "测线规划")
	}
	if err := s.audit.Record(actor, "plan.update", "transect_plan", id, before, updated, map[string]any{"expected_version": request.ExpectedVersion}); err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *TransectPlanService) Lock(id uint, request dto.PlanTransitionRequest, actor Actor) (model.TransectPlan, error) {
	before, err := s.Get(id)
	if err != nil {
		return model.TransectPlan{}, err
	}
	if before.PlanState != constants.PlanDraft || request.TargetState != constants.PlanLocked {
		return model.TransectPlan{}, api.Conflict("PLAN_STATE_INVALID", "规划状态迁移无效", nil)
	}
	updated, err := s.repository.Transition(id, request.ExpectedVersion, constants.PlanDraft, constants.PlanLocked)
	if err != nil {
		return updated, mapDatabaseError(err, "测线规划")
	}
	if err := s.audit.Record(actor, "plan.lock", "transect_plan", id, before, updated, nil); err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *TransectPlanService) Copy(id uint, actor Actor) (model.TransectPlan, error) {
	before, err := s.Get(id)
	if err != nil {
		return model.TransectPlan{}, err
	}
	copy, err := s.repository.Copy(before, actor.UserID)
	if err != nil {
		return copy, err
	}
	copy.CreatedAt = time.Now().UTC()
	copy.UpdatedAt = copy.CreatedAt
	if err := s.audit.Record(actor, "plan.copy", "transect_plan", copy.ID, before, copy, map[string]any{"source_plan_id": id}); err != nil {
		return copy, err
	}
	return copy, nil
}

func bounds(polygon orb.Polygon) (float64, float64, float64, float64) {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, ring := range polygon {
		for _, point := range ring {
			minX = math.Min(minX, point[0])
			minY = math.Min(minY, point[1])
			maxX = math.Max(maxX, point[0])
			maxY = math.Max(maxY, point[1])
		}
	}
	return minX, minY, maxX, maxY
}
