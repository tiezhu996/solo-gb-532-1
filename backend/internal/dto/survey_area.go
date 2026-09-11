package dto

import "encoding/json"

type CreateSurveyAreaRequest struct {
	AreaCode          string          `json:"area_code" binding:"required,min=3,max=40"`
	Name              string          `json:"name" binding:"required,min=3,max=160"`
	BoundaryGeoJSON   json.RawMessage `json:"boundary_geojson" binding:"required"`
	TargetResolutionM float64         `json:"target_resolution_m" binding:"required,gt=0,lte=100"`
	CoordinateSystem  string          `json:"coordinate_system" binding:"required,min=4,max=80"`
	DefaultSwathM     float64         `json:"default_swath_m" binding:"required,gt=0,lte=2000"`
	OwnerTeam         string          `json:"owner_team" binding:"required,min=2,max=120"`
}

type UpdateSurveyAreaRequest struct {
	Name              string          `json:"name" binding:"required,min=3,max=160"`
	BoundaryGeoJSON   json.RawMessage `json:"boundary_geojson" binding:"required"`
	TargetResolutionM float64         `json:"target_resolution_m" binding:"required,gt=0,lte=100"`
	CoordinateSystem  string          `json:"coordinate_system" binding:"required,min=4,max=80"`
	DefaultSwathM     float64         `json:"default_swath_m" binding:"required,gt=0,lte=2000"`
	OwnerTeam         string          `json:"owner_team" binding:"required,min=2,max=120"`
	Status            string          `json:"status" binding:"required,oneof=draft active archived"`
	ExpectedVersion   uint            `json:"expected_version" binding:"required,gt=0"`
}

type SurveyAreaQuery struct {
	Status   string
	Search   string
	Page     int
	PageSize int
}

type AreaSummary struct {
	PlanCount      int64   `json:"plan_count"`
	ProcessedRuns  int64   `json:"processed_runs"`
	OpenGapCount   int64   `json:"open_gap_count"`
	LatestCoverage float64 `json:"latest_coverage_ratio"`
}
