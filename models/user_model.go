package models

import (
	"time"
)

type UserAccount struct {
	UserID      uint   `gorm:"primaryKey" json:"user_id"`
	Username    string `gorm:"unique;not null" json:"username"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `gorm:"unique" json:"email"`
	Position    string `json:"position"`
	Institution string `json:"institution"`
	Role        string `gorm:"default:'user'" json:"role"`
	Status      string `gorm:"default:'active'" json:"status"`
	CreatedAt   time.Time
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
