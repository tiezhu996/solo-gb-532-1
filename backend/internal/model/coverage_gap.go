package model

import (
	"time"

	"gorm.io/datatypes"
)

type CoverageGap struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	SurveyAreaID           uint           `json:"survey_area_id" gorm:"not null;index"`
	SourceRunIDs           datatypes.JSON `json:"source_run_ids" gorm:"column:source_run_ids;type:jsonb;not null"`
	GapGeoJSON             datatypes.JSON `json:"gap_geojson" gorm:"column:gap_geojson;type:jsonb;not null"`
	AreaSquareM            float64        `json:"area_square_m" gorm:"not null"`
	GapRatio               float64        `json:"gap_ratio" gorm:"not null"`
	Severity               string         `json:"severity" gorm:"size:20;not null;index"`
	RecommendedLineGeoJSON datatypes.JSON `json:"recommended_line_geojson" gorm:"column:recommended_line_geojson;type:jsonb;not null"`
	AlgorithmVersion       string         `json:"algorithm_version" gorm:"size:40;not null"`
	InputHash              string         `json:"input_hash" gorm:"size:64;not null;uniqueIndex:idx_gap_idempotency"`
	GapState               string         `json:"gap_state" gorm:"size:24;not null;index"`
	Explanation            string         `json:"explanation" gorm:"size:1200;not null"`
	CoverageRatio          float64        `json:"coverage_ratio" gorm:"not null"`
	OverlapRatio           float64        `json:"overlap_ratio" gorm:"not null"`
	ProcessingMillis       int64          `json:"processing_millis" gorm:"not null"`
	Version                uint           `json:"version" gorm:"not null;default:1"`
	DetectedAt             time.Time      `json:"detected_at" gorm:"not null"`
	UpdatedAt              time.Time      `json:"updated_at"`
	SurveyArea             *SurveyArea    `json:"survey_area,omitempty" gorm:"foreignKey:SurveyAreaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (CoverageGap) TableName() string { return "coverage_gaps" }
