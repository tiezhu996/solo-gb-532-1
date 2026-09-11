package model

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:80;not null;uniqueIndex"`
	DisplayName  string    `json:"display_name" gorm:"size:120;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	Role         string    `json:"role" gorm:"size:32;not null;index"`
	Active       bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at"`
}

func (User) TableName() string { return "users" }
