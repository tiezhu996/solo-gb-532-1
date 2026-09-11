package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/geometry"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type CoverageResultView struct {
	Gap        model.CoverageGap    `json:"gap"`
	Evidence   dto.CoverageEvidence `json:"evidence"`
	Idempotent bool                 `json:"idempotent"`
}

type CoverageGapService struct {
	repository *repository.CoverageGapRepository
	areas      *repository.SurveyAreaRepository
	runs       *repository.SonarRunRepository
	plans      *repository.TransectPlanRepository
	audit      *AuditService
}

func NewCoverageGapService(repository *repository.CoverageGapRepository, areas *repository.SurveyAreaRepository, runs *repository.SonarRunRepository, plans *repository.TransectPlanRepository, audit *AuditService) *CoverageGapService {
	return &CoverageGapService{repository: repository, areas: areas, runs: runs, plans: plans, audit: audit}
}

func (s *CoverageGapService) List(query dto.CoverageGapQuery) ([]model.CoverageGap, int64, error) {
	return s.repository.List(query)
}
func (s *CoverageGapService) Get(id uint) (model.CoverageGap, error) {
	item, err := s.repository.Get(id)
	if err != nil {
		return item, mapDatabaseError(err, "覆盖缺口")
	}
	return item, nil
}

func (s *CoverageGapService) Detect(request dto.DetectCoverageRequest, idempotencyKey string, actor Actor) (CoverageResultView, error) {
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 120 {
		return CoverageResultView{}, api.BadRequest("IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key 必须为 8 到 120 个字符", nil)
	}
	area, err := s.areas.Get(request.SurveyAreaID)
	if err != nil {
		return CoverageResultView{}, mapDatabaseError(err, "测区")
	}
	if err := geometry.ValidateProjectedCRS(area.CoordinateSystem); err != nil {
		return CoverageResultView{}, api.Unprocessable("COORDINATE_SYSTEM_INVALID", "覆盖计算只支持米制投影坐标", err)
	}
	runs, err := s.runs.ByIDs(request.SourceRunIDs)
	if err != nil {
		return CoverageResultView{}, err
	}
	if len(runs) != len(uniqueIDs(request.SourceRunIDs)) {
		return CoverageResultView{}, api.Unprocessable("RUN_SET_INCOMPLETE", "部分运行不存在或重复", nil)
	}
	checksums := make([]string, 0, len(runs))
	tracks := make([]geometry.Track, 0, len(runs))
	for _, run := range runs {
		if run.TransectPlan == nil || run.TransectPlan.SurveyAreaID != area.ID {
			return CoverageResultView{}, api.Unprocessable("RUN_AREA_MISMATCH", "运行不属于所选测区", nil)
		}
		if run.RunState != string(constants.RunProcessed) {
			return CoverageResultView{}, api.Conflict("RUN_NOT_PROCESSED", "仅已处理运行可用于覆盖计算", nil)
		}
		lines, parseErr := geometry.ParseLines(run.TrackGeoJSON)
		if parseErr != nil {
			return CoverageResultView{}, api.Unprocessable("GEOJSON_INVALID", "运行航迹无法解析", parseErr)
		}
		checksums = append(checksums, run.SourceChecksum)
		tracks = append(tracks, geometry.Track{RunID: run.ID, Lines: lines, SwathM: run.ActualSwathM})
	}
	resolution := request.ResolutionM
	if resolution == 0 {
		resolution = area.TargetResolutionM
	}
	hash := geometry.StableInputHash(area.ID, area.CoordinateSystem, request.AlgorithmVersion, checksums, resolution)
	if existing, lookupErr := s.repository.ByInputHash(hash); lookupErr == nil {
		return CoverageResultView{Gap: existing, Evidence: evidenceFromGap(existing, area.CoordinateSystem, len(runs)), Idempotent: true}, nil
	} else if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
		return CoverageResultView{}, lookupErr
	}
	boundary, err := geometry.ParsePolygon(area.BoundaryGeoJSON)
	if err != nil {
		return CoverageResultView{}, api.Unprocessable("GEOJSON_INVALID", "测区边界无法计算", err)
	}
	started := time.Now()
	result, err := geometry.CalculateCoverage(boundary, tracks, resolution)
	if err != nil {
		return CoverageResultView{}, api.Unprocessable("COVERAGE_CALCULATION_INVALID", "覆盖计算参数或几何无效", err)
	}
	ids, _ := json.Marshal(uniqueIDs(request.SourceRunIDs))
	elapsed := time.Since(started).Milliseconds()
	item := model.CoverageGap{SurveyAreaID: area.ID, SourceRunIDs: ids, GapGeoJSON: result.GapGeoJSON, AreaSquareM: result.AreaSquareM * result.GapRatio, GapRatio: result.GapRatio, Severity: string(constants.SeverityForRatio(result.GapRatio)), RecommendedLineGeoJSON: result.RecommendedGeoJSON, AlgorithmVersion: request.AlgorithmVersion, InputHash: hash, GapState: string(constants.GapDetected), Explanation: fmt.Sprintf("以 %.1f 米网格分析 %d 个单元；过滤 %d 个低于分辨率阈值的碎片。补测线仅供人工规划参考。", resolution, result.SampleCells, result.FilteredFragments), CoverageRatio: result.CoverageRatio, OverlapRatio: result.OverlapRatio, ProcessingMillis: elapsed, Version: 1, DetectedAt: time.Now().UTC()}
	if err := s.repository.Create(&item); err != nil {
		return CoverageResultView{}, mapDatabaseError(err, "覆盖输入")
	}
	evidence := evidenceFromGap(item, area.CoordinateSystem, len(runs))
	if err := s.audit.Record(actor, "coverage.detect", "coverage_gap", item.ID, nil, item, map[string]any{"input_checksum": hash, "idempotency_key": idempotencyKey, "coordinate_system": area.CoordinateSystem, "resolution_m": resolution, "source_run_count": len(runs), "processing_millis": elapsed, "result_summary": evidence}); err != nil {
		return CoverageResultView{}, err
	}
	return CoverageResultView{Gap: item, Evidence: evidence}, nil
}

// GenerateResurveyPlan 把已复核且有建议线的覆盖快照转为只含补测线的草稿测线规划，
// 保留原规划、快照与输入哈希的关联；每个快照最多生成一次。
func (s *CoverageGapService) GenerateResurveyPlan(gapID uint, request dto.GenerateResurveyPlanRequest, actor Actor) (model.TransectPlan, error) {
	gap, err := s.Get(gapID)
	if err != nil {
		return model.TransectPlan{}, err
	}
	if request.SurveyAreaID != gap.SurveyAreaID {
		return model.TransectPlan{}, api.Unprocessable("RESURVEY_AREA_MISMATCH", "所选测区与缺口快照所属测区不一致", nil)
	}
	if state := constants.GapState(gap.GapState); state != constants.GapReviewed && state != constants.GapAccepted {
		return model.TransectPlan{}, api.Conflict("GAP_NOT_REVIEWED", "覆盖快照尚未复核，不能生成补测规划", nil)
	}
	lines, err := geometry.ParseLines(gap.RecommendedLineGeoJSON)
	if err != nil {
		return model.TransectPlan{}, api.Unprocessable("GEOJSON_INVALID", "快照建议补测线无法解析", err)
	}
	if geometry.TrackLength(lines) <= 0 {
		return model.TransectPlan{}, api.Unprocessable("RECOMMENDED_LINES_EMPTY", "快照没有可用的建议补测线", nil)
	}
	if _, err := s.plans.BySourceGap(gapID); err == nil {
		return model.TransectPlan{}, api.Conflict("RESURVEY_PLAN_EXISTS", "该快照已生成补测规划，请勿重复提交", nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.TransectPlan{}, err
	}
	area, err := s.areas.Get(gap.SurveyAreaID)
	if err != nil {
		return model.TransectPlan{}, mapDatabaseError(err, "测区")
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = fmt.Sprintf("缺口 #%d 补测方案", gap.ID)
	}
	swath := request.PlannedSwathM
	if swath == 0 {
		swath = area.DefaultSwathM
	}
	spacing := request.LineSpacingM
	if spacing == 0 {
		spacing = area.DefaultSwathM
	}
	plan := model.TransectPlan{SurveyAreaID: gap.SurveyAreaID, Name: name, LineGeoJSON: gap.RecommendedLineGeoJSON, PlannedHeading: geometry.LineHeading(lines[0]), PlannedSwathM: swath, LineSpacingM: spacing, PlanState: constants.PlanDraft, PlanSource: constants.PlanSourceResurvey, SourceGapID: &gap.ID, SourcePlanID: s.sourcePlanID(gap), SourceInputHash: gap.InputHash, Version: 1, CreatedBy: actor.UserID}
	if err := s.plans.Create(&plan); err != nil {
		return plan, mapDatabaseError(err, "补测规划")
	}
	if err := s.audit.Record(actor, "coverage.resurvey_plan", "transect_plan", plan.ID, nil, plan, map[string]any{"source_gap_id": gap.ID, "gap_version": gap.Version, "input_hash": gap.InputHash, "source_plan_id": plan.SourcePlanID, "algorithm_version": gap.AlgorithmVersion}); err != nil {
		return plan, err
	}
	created, err := s.plans.Get(plan.ID)
	if err != nil {
		return created, mapDatabaseError(err, "补测规划")
	}
	return created, nil
}

// sourcePlanID 取快照来源运行所属的原测线规划，作为补测规划的原规划关联。
func (s *CoverageGapService) sourcePlanID(gap model.CoverageGap) *uint {
	var runIDs []uint
	if err := json.Unmarshal(gap.SourceRunIDs, &runIDs); err != nil || len(runIDs) == 0 {
		return nil
	}
	runs, err := s.runs.ByIDs(uniqueIDs(runIDs))
	if err != nil {
		return nil
	}
	for _, run := range runs {
		if run.TransectPlanID > 0 {
			id := run.TransectPlanID
			return &id
		}
	}
	return nil
}

func (s *CoverageGapService) Transition(id uint, request dto.GapTransitionRequest, actor Actor) (model.CoverageGap, error) {
	before, err := s.Get(id)
	if err != nil {
		return model.CoverageGap{}, err
	}
	from, err := constants.ParseGapState(before.GapState)
	if err != nil {
		return model.CoverageGap{}, fmt.Errorf("stored gap state: %w", err)
	}
	target, err := constants.ParseGapState(request.TargetState)
	if err != nil {
		return model.CoverageGap{}, api.Unprocessable("GAP_STATE_INVALID", "目标缺口状态无效", err)
	}
	if !from.CanTransition(target) {
		return model.CoverageGap{}, api.Conflict("GAP_TRANSITION_INVALID", fmt.Sprintf("不能从 %s 迁移到 %s", from, target), nil)
	}
	explanation := before.Explanation + " 复核记录：" + request.ReviewNote
	updated, err := s.repository.Transition(id, request.ExpectedVersion, before.GapState, request.TargetState, explanation)
	if err != nil {
		return updated, mapDatabaseError(err, "覆盖缺口")
	}
	if err := s.audit.Record(actor, "coverage.transition", "coverage_gap", id, before, updated, map[string]any{"review_note": request.ReviewNote}); err != nil {
		return updated, err
	}
	return updated, nil
}

func evidenceFromGap(gap model.CoverageGap, crs string, sourceCount int) dto.CoverageEvidence {
	return dto.CoverageEvidence{InputHash: gap.InputHash, CoordinateSystem: crs, AlgorithmVersion: gap.AlgorithmVersion, SourceRunCount: sourceCount, CoverageRatio: gap.CoverageRatio, OverlapRatio: gap.OverlapRatio, GapRatio: gap.GapRatio, ProcessingMillis: gap.ProcessingMillis, DecisionBoundaryNote: "离线网格近似用于识别疑似缺口，不构成船舶或声呐控制指令。"}
}

func uniqueIDs(values []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
