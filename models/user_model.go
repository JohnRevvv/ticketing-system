package models

import (
	"time"
)

type UserAccount struct {
	UserID    uint   `gorm:"primaryKey" json:"user_id"`
	Username  string `gorm:"unique;not null" json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `gorm:"unique" json:"email"`
	Position  string `json:"position"`
	Role      string `json:"role"`
	Status    string `gorm:"default:'active'" json:"status"`
	CreatedAt time.Time
}

func (UserAccount) TableName() string {
	return "useraccount"
}
