package service

import (
	"errors"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

func newResurveyFixture(t *testing.T) (*CoverageGapService, *TransectPlanService, *gorm.DB) {
	t.Helper()
	configuration := config.Config{DBDriver: "sqlite", DBDSN: "file:" + t.Name() + "?mode=memory&cache=shared", JWTSecret: "test-secret-with-more-than-thirty-two-characters", AutoMigrate: true, SeedData: true}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	support := repository.NewSupportRepository(db)
	audit := NewAuditService(support)
	service := NewCoverageGapService(repository.NewCoverageGapRepository(db), repository.NewSurveyAreaRepository(db), repository.NewSonarRunRepository(db), repository.NewTransectPlanRepository(db), audit)
	plans := NewTransectPlanService(repository.NewTransectPlanRepository(db), repository.NewSurveyAreaRepository(db), audit)
	return service, plans, db
}

func reviewedGap(t *testing.T, service *CoverageGapService, actor Actor) model.CoverageGap {
	t.Helper()
	result, err := service.Detect(dto.DetectCoverageRequest{SurveyAreaID: 1, SourceRunIDs: []uint{1}, AlgorithmVersion: "grid-cover-v1.0.0", ResolutionM: 20}, "resurvey-test-idempotency", actor)
	if err != nil {
		t.Fatalf("detect coverage: %v", err)
	}
	gap, err := service.Transition(result.Gap.ID, dto.GapTransitionRequest{TargetState: string(constants.GapReviewed), ExpectedVersion: result.Gap.Version, ReviewNote: "测试复核意见记录"}, actor)
	if err != nil {
		t.Fatalf("review gap: %v", err)
	}
	return gap
}

func appErrorCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *api.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	return appErr.Code
}

func TestGenerateResurveyPlanFromReviewedGap(t *testing.T) {
	service, plans, _ := newResurveyFixture(t)
	actor := Actor{RequestID: "req-resurvey", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	gap := reviewedGap(t, service, actor)

	plan, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor)
	if err != nil {
		t.Fatalf("generate resurvey plan: %v", err)
	}
	if plan.PlanState != constants.PlanDraft || plan.PlanSource != constants.PlanSourceResurvey {
		t.Fatalf("plan state/source = %s/%s, want draft/resurvey", plan.PlanState, plan.PlanSource)
	}
	if plan.SourceGapID == nil || *plan.SourceGapID != gap.ID {
		t.Fatalf("source gap link missing: %+v", plan.SourceGapID)
	}
	if plan.SourcePlanID == nil || *plan.SourcePlanID != 1 {
		t.Fatalf("source plan link = %+v, want original plan 1", plan.SourcePlanID)
	}
	if plan.SourceInputHash != gap.InputHash {
		t.Fatal("input hash association not preserved")
	}
	if string(plan.LineGeoJSON) != string(gap.RecommendedLineGeoJSON) {
		t.Fatal("resurvey plan must contain only the recommended lines")
	}
	if plan.SourceGapVersion == nil || *plan.SourceGapVersion != gap.Version {
		t.Fatalf("frozen gap version = %+v, want %d", plan.SourceGapVersion, gap.Version)
	}
	if plan.SourceGapState != gap.GapState {
		t.Fatalf("frozen gap state = %s, want %s", plan.SourceGapState, gap.GapState)
	}
	if plan.PlannedHeading != 90 {
		t.Fatalf("heading = %.1f, want 90 for east-west recommendation", plan.PlannedHeading)
	}

	original, err := plans.Get(1)
	if err != nil {
		t.Fatalf("get original plan: %v", err)
	}
	if original.PlanState != constants.PlanLocked || original.Version != 1 {
		t.Fatal("original plan must stay locked and untouched")
	}
}

func TestResurveyPlanFreezesSourceSnapshot(t *testing.T) {
	service, plans, _ := newResurveyFixture(t)
	actor := Actor{RequestID: "req-freeze", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	gap := reviewedGap(t, service, actor)

	plan, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor)
	if err != nil {
		t.Fatalf("generate resurvey plan: %v", err)
	}
	// 缺口继续复核迁移：reviewed -> accepted，快照版本与状态随之变化。
	moved, err := service.Transition(gap.ID, dto.GapTransitionRequest{TargetState: string(constants.GapAccepted), ExpectedVersion: gap.Version, ReviewNote: "接受补测建议的复核记录"}, actor)
	if err != nil {
		t.Fatalf("transition gap after generation: %v", err)
	}
	if moved.Version == gap.Version || moved.GapState == gap.GapState {
		t.Fatal("gap transition should advance version and state")
	}

	reloaded, err := plans.Get(plan.ID)
	if err != nil {
		t.Fatalf("reload resurvey plan: %v", err)
	}
	if reloaded.SourceGapVersion == nil || *reloaded.SourceGapVersion != gap.Version {
		t.Fatalf("frozen gap version drifted to %+v, want %d", reloaded.SourceGapVersion, gap.Version)
	}
	if reloaded.SourceGapState != gap.GapState {
		t.Fatalf("frozen gap state drifted to %s, want %s", reloaded.SourceGapState, gap.GapState)
	}
	if reloaded.SourceInputHash != gap.InputHash {
		t.Fatal("frozen input hash must not change")
	}
	if reloaded.SourceGap == nil || reloaded.SourceGap.Version != moved.Version {
		t.Fatal("live association should still track the current snapshot for reference")
	}
}

func TestGenerateResurveyPlanGuards(t *testing.T) {
	service, _, _ := newResurveyFixture(t)
	actor := Actor{RequestID: "req-guard", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	gap := reviewedGap(t, service, actor)

	if _, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID + 9}, actor); appErrorCode(t, err) != "RESURVEY_AREA_MISMATCH" {
		t.Fatalf("area mismatch should be rejected, got %v", err)
	}
	if _, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor); err != nil {
		t.Fatalf("first generation should succeed: %v", err)
	}
	if _, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor); appErrorCode(t, err) != "RESURVEY_PLAN_EXISTS" {
		t.Fatalf("duplicate submission should be rejected, got %v", err)
	}
}

func TestGenerateResurveyPlanRequiresReviewAndLines(t *testing.T) {
	service, _, _ := newResurveyFixture(t)
	actor := Actor{RequestID: "req-state", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	result, err := service.Detect(dto.DetectCoverageRequest{SurveyAreaID: 1, SourceRunIDs: []uint{1}, AlgorithmVersion: "grid-cover-v1.0.0", ResolutionM: 20}, "resurvey-state-idempotency", actor)
	if err != nil {
		t.Fatalf("detect coverage: %v", err)
	}
	if _, err := service.GenerateResurveyPlan(result.Gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: 1}, actor); appErrorCode(t, err) != "GAP_NOT_REVIEWED" {
		t.Fatalf("unreviewed snapshot should be rejected, got %v", err)
	}

	degenerate := model.CoverageGap{SurveyAreaID: 1, SourceRunIDs: datatypes.JSON([]byte(`[1]`)), GapGeoJSON: datatypes.JSON([]byte(`{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}}`)), AreaSquareM: 1, GapRatio: 0.01, Severity: string(constants.SeverityMinor), RecommendedLineGeoJSON: datatypes.JSON([]byte(`{"type":"Feature","properties":{},"geometry":{"type":"LineString","coordinates":[[500,300],[500,300]]}}`)), AlgorithmVersion: "grid-cover-v1.0.0", InputHash: "degenerate-lines-hash", GapState: string(constants.GapReviewed), Explanation: "测试空建议线快照", Version: 1, DetectedAt: time.Now().UTC()}
	if err := service.repository.Create(&degenerate); err != nil {
		t.Fatalf("insert degenerate gap: %v", err)
	}
	if _, err := service.GenerateResurveyPlan(degenerate.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: 1}, actor); appErrorCode(t, err) != "RECOMMENDED_LINES_EMPTY" {
		t.Fatalf("empty recommended lines should be rejected, got %v", err)
	}
}

func TestGenerateResurveyPlanAuditFailureRollsBack(t *testing.T) {
	service, _, db := newResurveyFixture(t)
	actor := Actor{RequestID: "req-audit-failure", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	gap := reviewedGap(t, service, actor)

	// 模拟审计写入失败：审计表不可用，规划与审计必须一起回滚。
	if err := db.Exec("DROP TABLE audit_events").Error; err != nil {
		t.Fatalf("drop audit table: %v", err)
	}
	if _, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor); err == nil {
		t.Fatal("audit failure must abort resurvey plan generation")
	}
	if _, err := service.plans.BySourceGap(gap.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("rolled-back plan must not remain, lookup err = %v", err)
	}
	var count int64
	if err := db.Model(&model.TransectPlan{}).Where("plan_source = ?", constants.PlanSourceResurvey).Count(&count).Error; err != nil {
		t.Fatalf("count resurvey plans: %v", err)
	}
	if count != 0 {
		t.Fatalf("found %d leftover resurvey plans after rollback", count)
	}

	// 恢复审计能力后重试不得被判成重复提交。
	if err := db.AutoMigrate(&model.AuditEvent{}); err != nil {
		t.Fatalf("restore audit table: %v", err)
	}
	plan, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor)
	if err != nil {
		t.Fatalf("retry after recovery should succeed, got %v", err)
	}
	if plan.SourceGapID == nil || *plan.SourceGapID != gap.ID {
		t.Fatal("recovered generation must keep the source gap link")
	}
}
