package model

import (
	"time"

	"gorm.io/datatypes"
)

type SurveyArea struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	AreaCode          string         `json:"area_code" gorm:"size:40;not null;uniqueIndex"`
	Name              string         `json:"name" gorm:"size:160;not null"`
	BoundaryGeoJSON   datatypes.JSON `json:"boundary_geojson" gorm:"column:boundary_geojson;type:jsonb;not null"`
	TargetResolutionM float64        `json:"target_resolution_m" gorm:"not null;check:target_resolution_m > 0"`
	CoordinateSystem  string         `json:"coordinate_system" gorm:"size:80;not null"`
	DefaultSwathM     float64        `json:"default_swath_m" gorm:"not null;check:default_swath_m > 0"`
	OwnerTeam         string         `json:"owner_team" gorm:"size:120;not null"`
	Status            string         `json:"status" gorm:"size:20;not null;index"`
	Version           uint           `json:"version" gorm:"not null;default:1"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (SurveyArea) TableName() string { return "survey_areas" }
