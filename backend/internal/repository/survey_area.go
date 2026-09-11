package repository

import (
	"fmt"

	"gorm.io/gorm"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/model"
)

type SurveyAreaRepository struct{ db *gorm.DB }

func NewSurveyAreaRepository(db *gorm.DB) *SurveyAreaRepository { return &SurveyAreaRepository{db: db} }

func (r *SurveyAreaRepository) List(query dto.SurveyAreaQuery) ([]model.SurveyArea, int64, error) {
	db := r.db.Model(&model.SurveyArea{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		term := "%" + query.Search + "%"
		db = db.Where("area_code LIKE ? OR name LIKE ? OR owner_team LIKE ?", term, term, term)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count survey areas: %w", err)
	}
	var items []model.SurveyArea
	if err := db.Order("updated_at DESC, id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list survey areas: %w", err)
	}
	return items, total, nil
}

func (r *SurveyAreaRepository) Get(id uint) (model.SurveyArea, error) {
	var item model.SurveyArea
	if err := r.db.First(&item, id).Error; err != nil {
		return model.SurveyArea{}, fmt.Errorf("get survey area: %w", err)
	}
	return item, nil
}

func (r *SurveyAreaRepository) Create(item *model.SurveyArea) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("create survey area: %w", err)
	}
	return nil
}

func (r *SurveyAreaRepository) Update(id, expectedVersion uint, values map[string]any) (model.SurveyArea, error) {
	values["version"] = gorm.Expr("version + 1")
	result := r.db.Model(&model.SurveyArea{}).Where("id = ? AND version = ?", id, expectedVersion).Updates(values)
	if result.Error != nil {
		return model.SurveyArea{}, fmt.Errorf("update survey area: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.SurveyArea{}, ErrVersionConflict
	}
	return r.Get(id)
}

func (r *SurveyAreaRepository) Summary(id uint) (dto.AreaSummary, error) {
	var summary dto.AreaSummary
	if err := r.db.Model(&model.TransectPlan{}).Where("survey_area_id = ?", id).Count(&summary.PlanCount).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&model.SonarRun{}).Joins("JOIN transect_plans p ON p.id = sonar_runs.transect_plan_id").Where("p.survey_area_id = ? AND sonar_runs.run_state = ?", id, "processed").Count(&summary.ProcessedRuns).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&model.CoverageGap{}).Where("survey_area_id = ? AND gap_state <> ?", id, "closed").Count(&summary.OpenGapCount).Error; err != nil {
		return summary, err
	}
	var gap model.CoverageGap
	if err := r.db.Where("survey_area_id = ?", id).Order("detected_at DESC").First(&gap).Error; err == nil {
		summary.LatestCoverage = gap.CoverageRatio
	} else if err != gorm.ErrRecordNotFound {
		return summary, err
	}
	return summary, nil
}
