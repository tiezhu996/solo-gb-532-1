package model

import (
	"time"

	"gorm.io/datatypes"
)

type TransectPlan struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	SurveyAreaID    uint           `json:"survey_area_id" gorm:"not null;index"`
	Name            string         `json:"name" gorm:"size:160;not null"`
	LineGeoJSON     datatypes.JSON `json:"line_geojson" gorm:"column:line_geojson;type:jsonb;not null"`
	PlannedHeading  float64        `json:"planned_heading" gorm:"not null"`
	PlannedSwathM   float64        `json:"planned_swath_m" gorm:"not null;check:planned_swath_m > 0"`
	LineSpacingM    float64        `json:"line_spacing_m" gorm:"not null;check:line_spacing_m > 0"`
	PlanState       string         `json:"plan_state" gorm:"size:20;not null;index"`
	PlanSource      string         `json:"plan_source" gorm:"size:20;not null;default:manual;index"`
	SourceGapID     *uint          `json:"source_gap_id" gorm:"uniqueIndex:idx_plan_source_gap"`
	SourcePlanID    *uint          `json:"source_plan_id" gorm:"index"`
	SourceInputHash string         `json:"source_input_hash" gorm:"size:64"`
	// 生成时冻结的快照版本与状态，后续缺口复核不再改写。
	SourceGapVersion *uint         `json:"source_gap_version"`
	SourceGapState   string        `json:"source_gap_state" gorm:"size:24"`
	Version          uint          `json:"version" gorm:"not null;default:1"`
	CreatedBy        uint          `json:"created_by" gorm:"not null"`
	UpdatedAt        time.Time     `json:"updated_at"`
	CreatedAt        time.Time     `json:"created_at"`
	SurveyArea       *SurveyArea   `json:"survey_area,omitempty" gorm:"foreignKey:SurveyAreaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	SourceGap        *CoverageGap  `json:"source_gap,omitempty" gorm:"foreignKey:SourceGapID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	SourcePlan       *TransectPlan `json:"source_plan,omitempty" gorm:"foreignKey:SourcePlanID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (TransectPlan) TableName() string { return "transect_plans" }
