package service

import (
	"errors"
	"testing"
	"time"

	"gorm.io/datatypes"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

func newResurveyFixture(t *testing.T) (*CoverageGapService, *TransectPlanService) {
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
	return service, plans
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
	service, plans := newResurveyFixture(t)
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
	if plan.SourceGap == nil || plan.SourceGap.Version != gap.Version {
		t.Fatal("source gap version should be preloaded for provenance display")
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

func TestGenerateResurveyPlanGuards(t *testing.T) {
	service, _ := newResurveyFixture(t)
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
	service, _ := newResurveyFixture(t)
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
