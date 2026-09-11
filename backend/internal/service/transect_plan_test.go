package service

import (
	"testing"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/dto"
)

func TestCopyResurveyPlanClearsProvenance(t *testing.T) {
	service, plans, _ := newResurveyFixture(t)
	actor := Actor{RequestID: "req-copy-resurvey", UserID: 4, Username: "reviewer", Role: constants.RoleReviewer}
	gap := reviewedGap(t, service, actor)

	plan, err := service.GenerateResurveyPlan(gap.ID, dto.GenerateResurveyPlanRequest{SurveyAreaID: gap.SurveyAreaID}, actor)
	if err != nil {
		t.Fatalf("generate resurvey plan: %v", err)
	}
	locked, err := plans.Lock(plan.ID, dto.PlanTransitionRequest{TargetState: constants.PlanLocked, ExpectedVersion: plan.Version}, actor)
	if err != nil {
		t.Fatalf("lock resurvey plan: %v", err)
	}

	copy, err := plans.Copy(locked.ID, actor)
	if err != nil {
		t.Fatalf("copy locked resurvey plan: %v", err)
	}
	if copy.PlanState != constants.PlanDraft || copy.PlanSource != constants.PlanSourceManual {
		t.Fatalf("copy state/source = %s/%s, want draft/manual", copy.PlanState, copy.PlanSource)
	}
	if copy.SourceGapID != nil || copy.SourcePlanID != nil {
		t.Fatalf("copy must not keep source links: gap=%+v plan=%+v", copy.SourceGapID, copy.SourcePlanID)
	}
	if copy.SourceInputHash != "" {
		t.Fatalf("copy must not keep input hash: %q", copy.SourceInputHash)
	}
	if copy.SourceGapVersion != nil || copy.SourceGapState != "" {
		t.Fatalf("copy must not keep frozen snapshot: version=%+v state=%q", copy.SourceGapVersion, copy.SourceGapState)
	}
	if copy.Version != locked.Version+1 {
		t.Fatalf("copy version = %d, want %d", copy.Version, locked.Version+1)
	}

	reloaded, err := plans.Get(locked.ID)
	if err != nil {
		t.Fatalf("reload original resurvey plan: %v", err)
	}
	if reloaded.PlanState != constants.PlanLocked || reloaded.PlanSource != constants.PlanSourceResurvey {
		t.Fatal("original plan must stay locked resurvey")
	}
	if reloaded.SourceGapID == nil || *reloaded.SourceGapID != gap.ID || reloaded.SourceInputHash != gap.InputHash {
		t.Fatal("original source links must stay intact")
	}
	if reloaded.SourceGapVersion == nil || *reloaded.SourceGapVersion != gap.Version || reloaded.SourceGapState != gap.GapState {
		t.Fatal("original frozen snapshot must stay intact")
	}
}
