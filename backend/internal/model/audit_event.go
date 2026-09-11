package model

import "time"

type AuditEvent struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	RequestID  string    `json:"request_id" gorm:"size:64;not null;index"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	Actor      string    `json:"actor" gorm:"size:80;not null;index"`
	Role       string    `json:"role" gorm:"size:32;not null"`
	Action     string    `json:"action" gorm:"size:80;not null;index"`
	EntityType string    `json:"entity_type" gorm:"size:60;not null;index"`
	EntityID   uint      `json:"entity_id" gorm:"not null;index"`
	BeforeJSON string    `json:"before_json" gorm:"type:text;not null"`
	AfterJSON  string    `json:"after_json" gorm:"type:text;not null"`
	Metadata   string    `json:"metadata" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;index"`
}

func (AuditEvent) TableName() string { return "audit_events" }
