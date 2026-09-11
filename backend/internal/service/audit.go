package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
)

type Actor struct {
	RequestID string
	UserID    uint
	Username  string
	Role      string
}

type AuditService struct{ repository *repository.SupportRepository }

func NewAuditService(repository *repository.SupportRepository) *AuditService {
	return &AuditService{repository: repository}
}

// WithTx 返回在同一事务中写入审计事件的服务，配合业务写入一起提交或回滚。
func (s *AuditService) WithTx(tx *gorm.DB) *AuditService {
	return &AuditService{repository: s.repository.WithTx(tx)}
}

func (s *AuditService) Record(actor Actor, action, entityType string, entityID uint, before, after, metadata any) error {
	beforeJSON, err := encodeSnapshot(before)
	if err != nil {
		return fmt.Errorf("encode audit before snapshot: %w", err)
	}
	afterJSON, err := encodeSnapshot(after)
	if err != nil {
		return fmt.Errorf("encode audit after snapshot: %w", err)
	}
	metadataJSON, err := encodeSnapshot(metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}
	event := model.AuditEvent{RequestID: actor.RequestID, UserID: actor.UserID, Actor: actor.Username, Role: actor.Role, Action: action, EntityType: entityType, EntityID: entityID, BeforeJSON: beforeJSON, AfterJSON: afterJSON, Metadata: metadataJSON, CreatedAt: time.Now().UTC()}
	return s.repository.CreateAudit(&event)
}

func (s *AuditService) List(filter repository.AuditFilter) ([]model.AuditEvent, int64, error) {
	return s.repository.ListAudits(filter)
}

func (s *AuditService) Ready() error { return s.repository.Ping() }

func encodeSnapshot(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
