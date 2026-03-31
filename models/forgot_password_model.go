package models

import "time"

type PasswordResetToken struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint
    Token     string    `gorm:"unique"`
    ExpiresAt time.Time
    CreatedAt time.Time
}

func (PasswordResetToken) TableName() string {
    return "password_reset_tokens"
}