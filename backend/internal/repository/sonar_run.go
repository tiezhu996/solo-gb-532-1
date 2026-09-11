package repository

import (
	"fmt"

	"gorm.io/gorm"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/model"
)

type SonarRunRepository struct{ db *gorm.DB }

func NewSonarRunRepository(db *gorm.DB) *SonarRunRepository { return &SonarRunRepository{db: db} }

func (r *SonarRunRepository) List(query dto.SonarRunQuery) ([]model.SonarRun, int64, error) {
	db := r.db.Model(&model.SonarRun{})
	if query.TransectPlanID > 0 {
		db = db.Where("transect_plan_id = ?", query.TransectPlanID)
	}
	if query.State != "" {
		db = db.Where("run_state = ?", query.State)
	}
	if query.Quality != "" {
		db = db.Where("navigation_quality = ?", query.Quality)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count sonar runs: %w", err)
	}
	var items []model.SonarRun
	if err := db.Preload("TransectPlan").Order("started_at DESC, id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list sonar runs: %w", err)
	}
	return items, total, nil
}

func (r *SonarRunRepository) Get(id uint) (model.SonarRun, error) {
	var item model.SonarRun
	if err := r.db.Preload("TransectPlan.SurveyArea").First(&item, id).Error; err != nil {
		return item, fmt.Errorf("get sonar run: %w", err)
	}
	return item, nil
}

func (r *SonarRunRepository) ByIDs(ids []uint) ([]model.SonarRun, error) {
	var items []model.SonarRun
	if err := r.db.Preload("TransectPlan.SurveyArea").Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("get sonar runs: %w", err)
	}
	return items, nil
}

func (r *SonarRunRepository) ByChecksum(checksum string) (model.SonarRun, error) {
	var item model.SonarRun
	if err := r.db.Where("source_checksum = ?", checksum).First(&item).Error; err != nil {
		return item, fmt.Errorf("find run checksum: %w", err)
	}
	return item, nil
}

func (r *SonarRunRepository) Create(item *model.SonarRun) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("create sonar run: %w", err)
	}
	return nil
}

func (r *SonarRunRepository) Transition(id, expectedVersion uint, from, to string) (model.SonarRun, error) {
	result := r.db.Model(&model.SonarRun{}).Where("id = ? AND version = ? AND run_state = ?", id, expectedVersion, from).Updates(map[string]any{"run_state": to, "version": gorm.Expr("version + 1")})
	if result.Error != nil {
		return model.SonarRun{}, fmt.Errorf("transition sonar run: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.SonarRun{}, ErrVersionConflict
	}
	return r.Get(id)
}
