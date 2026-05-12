package models

import (
	"time"
)

type UserAccount struct {
	UserID      uint      `gorm:"primaryKey" json:"user_id"`
	Username    string    `gorm:"unique;not null" json:"username"`
	Password    string    `gorm:"not null" json:"password"`
	FirstName   string    `gorm:"not null" json:"first_name"`
	LastName    string    `gorm:"not null" json:"last_name"`
	Email       string    `gorm:"unique;not null" json:"email"`
	Position    string    `gorm:"not null" json:"position"`
	Institution string    `gorm:"not null" json:"institution"`
	Role        string    `gorm:"default:'user';not null" json:"role"`
	Status      string    `gorm:"default:'active';not null" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (UserAccount) TableName() string {
	return "useraccount"
}

type Category struct {
	CategoryID    uint          `json:"category_id" gorm:"primaryKey"`
	Name          string        `json:"name"`
	SubCategories []SubCategory `json:"subcategories" gorm:"foreignKey:CategoryID"`
	CreatedAt     time.Time     `json:"created_at"`
}

func (Category) TableName() string {
	return "categories"
}

type SubCategory struct {
	SubCategoryID uint      `json:"sub_category_id" gorm:"primaryKey"`
	CategoryID    uint      `json:"category_id"` // 🔥 REQUIRED
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

func (SubCategory) TableName() string {
	return "subcategories"
}
