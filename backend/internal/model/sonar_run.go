package model

import (
	"time"

	"gorm.io/datatypes"
)

type SonarRun struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	TransectPlanID    uint           `json:"transect_plan_id" gorm:"not null;index"`
	RunCode           string         `json:"run_code" gorm:"size:48;not null;uniqueIndex"`
	TrackGeoJSON      datatypes.JSON `json:"track_geojson" gorm:"column:track_geojson;type:jsonb;not null"`
	ActualSwathM      float64        `json:"actual_swath_m" gorm:"not null;check:actual_swath_m > 0"`
	StartedAt         time.Time      `json:"started_at" gorm:"not null;index"`
	EndedAt           time.Time      `json:"ended_at" gorm:"not null"`
	NavigationQuality string         `json:"navigation_quality" gorm:"size:20;not null"`
	RunState          string         `json:"run_state" gorm:"size:24;not null;index"`
	SourceChecksum    string         `json:"source_checksum" gorm:"size:64;not null;uniqueIndex"`
	ImportedBy        uint           `json:"imported_by" gorm:"not null"`
	Version           uint           `json:"version" gorm:"not null;default:1"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	TransectPlan      *TransectPlan  `json:"transect_plan,omitempty" gorm:"foreignKey:TransectPlanID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (SonarRun) TableName() string { return "sonar_runs" }
