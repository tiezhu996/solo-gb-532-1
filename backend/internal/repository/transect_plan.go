package repository

import (
	"fmt"

	"gorm.io/gorm"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/model"
)

type TransectPlanRepository struct{ db *gorm.DB }

func NewTransectPlanRepository(db *gorm.DB) *TransectPlanRepository {
	return &TransectPlanRepository{db: db}
}

func (r *TransectPlanRepository) List(query dto.TransectPlanQuery) ([]model.TransectPlan, int64, error) {
	db := r.db.Model(&model.TransectPlan{})
	if query.SurveyAreaID > 0 {
		db = db.Where("survey_area_id = ?", query.SurveyAreaID)
	}
	if query.State != "" {
		db = db.Where("plan_state = ?", query.State)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count transect plans: %w", err)
	}
	var items []model.TransectPlan
	if err := db.Preload("SurveyArea").Order("updated_at DESC, id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list transect plans: %w", err)
	}
	return items, total, nil
}

func (r *TransectPlanRepository) Get(id uint) (model.TransectPlan, error) {
	var item model.TransectPlan
	if err := r.db.Preload("SurveyArea").First(&item, id).Error; err != nil {
		return item, fmt.Errorf("get transect plan: %w", err)
	}
	return item, nil
}

func (r *TransectPlanRepository) Create(item *model.TransectPlan) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("create transect plan: %w", err)
	}
	return nil
}

func (r *TransectPlanRepository) Update(id, expectedVersion uint, values map[string]any) (model.TransectPlan, error) {
	values["version"] = gorm.Expr("version + 1")
	result := r.db.Model(&model.TransectPlan{}).Where("id = ? AND version = ?", id, expectedVersion).Updates(values)
	if result.Error != nil {
		return model.TransectPlan{}, fmt.Errorf("update transect plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.TransectPlan{}, ErrVersionConflict
	}
	return r.Get(id)
}

func (r *TransectPlanRepository) Transition(id, expectedVersion uint, from, to string) (model.TransectPlan, error) {
	result := r.db.Model(&model.TransectPlan{}).Where("id = ? AND version = ? AND plan_state = ?", id, expectedVersion, from).Updates(map[string]any{"plan_state": to, "version": gorm.Expr("version + 1")})
	if result.Error != nil {
		return model.TransectPlan{}, fmt.Errorf("transition transect plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.TransectPlan{}, ErrVersionConflict
	}
	return r.Get(id)
}

func (r *TransectPlanRepository) Copy(source model.TransectPlan, actorID uint) (model.TransectPlan, error) {
	copy := source
	copy.ID = 0
	copy.Name = source.Name + " / 复制版本"
	copy.PlanState = "draft"
	copy.Version = source.Version + 1
	copy.CreatedBy = actorID
	copy.CreatedAt = source.CreatedAt.Add(0)
	copy.UpdatedAt = source.UpdatedAt.Add(0)
	if err := r.db.Create(&copy).Error; err != nil {
		return model.TransectPlan{}, fmt.Errorf("copy transect plan: %w", err)
	}
	return copy, nil
}
