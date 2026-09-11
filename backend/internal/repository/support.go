package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"sonar-survey-coverage-planner/backend/internal/model"
)

var ErrVersionConflict = errors.New("resource version or state conflict")

type SupportRepository struct{ db *gorm.DB }

func NewSupportRepository(db *gorm.DB) *SupportRepository { return &SupportRepository{db: db} }

func (r *SupportRepository) UserByUsername(username string) (model.User, error) {
	var user model.User
	if err := r.db.Where("username = ? AND active = ?", username, true).First(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("find active user: %w", err)
	}
	return user, nil
}

func (r *SupportRepository) UserByID(id uint) (model.User, error) {
	var user model.User
	if err := r.db.Where("id = ? AND active = ?", id, true).First(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("find active user by id: %w", err)
	}
	return user, nil
}

func (r *SupportRepository) CreateAudit(event *model.AuditEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}

type AuditFilter struct {
	RequestID, EntityType, Actor string
	Page, PageSize               int
}

func (r *SupportRepository) ListAudits(filter AuditFilter) ([]model.AuditEvent, int64, error) {
	query := r.db.Model(&model.AuditEvent{})
	if filter.RequestID != "" {
		query = query.Where("request_id = ?", filter.RequestID)
	}
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.Actor != "" {
		query = query.Where("actor LIKE ?", "%"+filter.Actor+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit events: %w", err)
	}
	var events []model.AuditEvent
	if err := query.Order("created_at DESC, id DESC").Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit events: %w", err)
	}
	return events, total, nil
}

func (r *SupportRepository) Ping() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("database handle: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping: %w", err)
	}
	return nil
}
