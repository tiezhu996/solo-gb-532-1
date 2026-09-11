package constants

const (
	RoleAdmin         = "admin"
	RoleSurveyPlanner = "survey_planner"
	RoleDataProcessor = "data_processor"
	RoleReviewer      = "reviewer"
	RoleAuditor       = "auditor"

	AreaDraft    = "draft"
	AreaActive   = "active"
	AreaArchived = "archived"

	PlanDraft  = "draft"
	PlanLocked = "locked"

	NavGood     = "good"
	NavDegraded = "degraded"
	NavInvalid  = "invalid"
)

func ValidRole(value string) bool {
	switch value {
	case RoleAdmin, RoleSurveyPlanner, RoleDataProcessor, RoleReviewer, RoleAuditor:
		return true
	default:
		return false
	}
}

func CanMutate(role string) bool {
	return role == RoleAdmin || role == RoleSurveyPlanner || role == RoleDataProcessor || role == RoleReviewer
}
