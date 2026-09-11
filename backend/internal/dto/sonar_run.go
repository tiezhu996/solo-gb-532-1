package dto

import (
	"encoding/json"
	"time"
)

type ImportSonarRunRequest struct {
	TransectPlanID    uint            `json:"transect_plan_id" binding:"required,gt=0"`
	RunCode           string          `json:"run_code" binding:"required,min=3,max=48"`
	TrackGeoJSON      json.RawMessage `json:"track_geojson" binding:"required"`
	ActualSwathM      float64         `json:"actual_swath_m" binding:"required,gt=0,lte=2000"`
	StartedAt         time.Time       `json:"started_at" binding:"required"`
	EndedAt           time.Time       `json:"ended_at" binding:"required"`
	NavigationQuality string          `json:"navigation_quality" binding:"required,oneof=good degraded invalid"`
}

type RunTransitionRequest struct {
	TargetState     string `json:"target_state" binding:"required,oneof=quality_checked processing processed rejected superseded"`
	ExpectedVersion uint   `json:"expected_version" binding:"required,gt=0"`
	Reason          string `json:"reason" binding:"omitempty,max=500"`
}

type SonarRunQuery struct {
	TransectPlanID uint
	State          string
	Quality        string
	Page           int
	PageSize       int
}

type RunQualityEvidence struct {
	CoordinateCount int      `json:"coordinate_count"`
	TrackLengthM    float64  `json:"track_length_m"`
	DurationMinutes float64  `json:"duration_minutes"`
	AverageSpeedMPS float64  `json:"average_speed_mps"`
	Warnings        []string `json:"warnings"`
}
