package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/paulmach/orb"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/geometry"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type SonarRunService struct {
	repository *repository.SonarRunRepository
	plans      *repository.TransectPlanRepository
	audit      *AuditService
}

func NewSonarRunService(repository *repository.SonarRunRepository, plans *repository.TransectPlanRepository, audit *AuditService) *SonarRunService {
	return &SonarRunService{repository: repository, plans: plans, audit: audit}
}

func (s *SonarRunService) List(query dto.SonarRunQuery) ([]model.SonarRun, int64, error) {
	return s.repository.List(query)
}
func (s *SonarRunService) Get(id uint) (model.SonarRun, error) {
	item, err := s.repository.Get(id)
	if err != nil {
		return item, mapDatabaseError(err, "声呐运行")
	}
	return item, nil
}

func (s *SonarRunService) Import(request dto.ImportSonarRunRequest, actor Actor) (model.SonarRun, bool, error) {
	plan, err := s.plans.Get(request.TransectPlanID)
	if err != nil {
		return model.SonarRun{}, false, mapDatabaseError(err, "测线规划")
	}
	if plan.PlanState != constants.PlanLocked {
		return model.SonarRun{}, false, api.Conflict("PLAN_NOT_LOCKED", "仅可向已锁定规划导入航迹", nil)
	}
	lines, err := geometry.ParseLines(request.TrackGeoJSON)
	if err != nil {
		return model.SonarRun{}, false, api.Unprocessable("GEOJSON_INVALID", "航迹 GeoJSON 无效", err)
	}
	if !request.EndedAt.After(request.StartedAt) {
		return model.SonarRun{}, false, api.Unprocessable("RUN_TIME_INVALID", "结束时间必须晚于开始时间", nil)
	}
	checksum := checksumRun(request)
	existing, lookupErr := s.repository.ByChecksum(checksum)
	if lookupErr == nil {
		return existing, true, nil
	}
	if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
		return model.SonarRun{}, false, lookupErr
	}
	item := model.SonarRun{TransectPlanID: request.TransectPlanID, RunCode: request.RunCode, TrackGeoJSON: datatypes.JSON(request.TrackGeoJSON), ActualSwathM: request.ActualSwathM, StartedAt: request.StartedAt.UTC(), EndedAt: request.EndedAt.UTC(), NavigationQuality: request.NavigationQuality, RunState: string(constants.RunImported), SourceChecksum: checksum, ImportedBy: actor.UserID, Version: 1}
	if err := s.repository.Create(&item); err != nil {
		return item, false, mapDatabaseError(err, "运行编号或航迹来源")
	}
	evidence := qualityEvidence(lines, item.StartedAt, item.EndedAt)
	if err := s.audit.Record(actor, "run.import", "sonar_run", item.ID, nil, item, map[string]any{"checksum": checksum, "quality_evidence": evidence}); err != nil {
		return item, false, err
	}
	created, err := s.Get(item.ID)
	return created, false, err
}

func (s *SonarRunService) Quality(id uint) (dto.RunQualityEvidence, error) {
	item, err := s.Get(id)
	if err != nil {
		return dto.RunQualityEvidence{}, err
	}
	lines, err := geometry.ParseLines(item.TrackGeoJSON)
	if err != nil {
		return dto.RunQualityEvidence{}, api.Unprocessable("GEOJSON_INVALID", "已导入航迹无法解析", err)
	}
	return qualityEvidence(lines, item.StartedAt, item.EndedAt), nil
}

func (s *SonarRunService) Transition(id uint, request dto.RunTransitionRequest, actor Actor) (model.SonarRun, error) {
	before, err := s.Get(id)
	if err != nil {
		return model.SonarRun{}, err
	}
	from, err := constants.ParseRunState(before.RunState)
	if err != nil {
		return model.SonarRun{}, fmt.Errorf("stored run state: %w", err)
	}
	target, err := constants.ParseRunState(request.TargetState)
	if err != nil {
		return model.SonarRun{}, api.Unprocessable("RUN_STATE_INVALID", "目标状态无效", err)
	}
	if !from.CanTransition(target) {
		return model.SonarRun{}, api.Conflict("RUN_TRANSITION_INVALID", fmt.Sprintf("不能从 %s 迁移到 %s", from, target), nil)
	}
	if target == constants.RunQualityChecked && before.NavigationQuality == constants.NavInvalid {
		return model.SonarRun{}, api.Conflict("NAVIGATION_INVALID", "导航质量无效的航迹必须驳回", nil)
	}
	if target == constants.RunRejected && len(request.Reason) < 4 {
		return model.SonarRun{}, api.Unprocessable("REJECTION_REASON_REQUIRED", "驳回原因至少 4 个字符", nil)
	}
	updated, err := s.repository.Transition(id, request.ExpectedVersion, before.RunState, request.TargetState)
	if err != nil {
		return updated, mapDatabaseError(err, "声呐运行")
	}
	if err := s.audit.Record(actor, "run.transition", "sonar_run", id, before, updated, map[string]any{"reason": request.Reason}); err != nil {
		return updated, err
	}
	return updated, nil
}

func checksumRun(request dto.ImportSonarRunRequest) string {
	hash := sha256.New()
	hash.Write(request.TrackGeoJSON)
	hash.Write([]byte(fmt.Sprintf("|%d|%.3f|%s|%s", request.TransectPlanID, request.ActualSwathM, request.StartedAt.UTC().Format(time.RFC3339Nano), request.EndedAt.UTC().Format(time.RFC3339Nano))))
	return hex.EncodeToString(hash.Sum(nil))
}

func qualityEvidence(lines []orb.LineString, startedAt, endedAt time.Time) dto.RunQualityEvidence {
	coordinates := 0
	for _, line := range lines {
		coordinates += len(line)
	}
	length := geometry.TrackLength(lines)
	minutes := endedAt.Sub(startedAt).Minutes()
	speed := 0.0
	if minutes > 0 {
		speed = length / (minutes * 60)
	}
	warnings := []string{}
	if coordinates < 4 {
		warnings = append(warnings, "航迹采样点偏少")
	}
	if speed > 12 {
		warnings = append(warnings, "平均航速超过离线质量阈值")
	}
	if length < 100 {
		warnings = append(warnings, "有效航迹长度不足 100 米")
	}
	return dto.RunQualityEvidence{CoordinateCount: coordinates, TrackLengthM: length, DurationMinutes: minutes, AverageSpeedMPS: speed, Warnings: warnings}
}
