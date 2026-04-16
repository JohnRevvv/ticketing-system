package models

import (
	"time"
)

type UserAccount struct {
	UserID    uint   `gorm:"primaryKey" json:"user_id"`
	Username  string `gorm:"unique;not null" json:"username"`
	Password  string `json:"password"`
	FullName  string `json:"full_name"`
	Email     string `gorm:"unique" json:"email"`
	Position  string `json:"position"`
	Role      string `json:"role"`
	Status    string `gorm:"default:'active'" json:"status"`
	CreatedAt time.Time
}

func (UserAccount) TableName() string {
	return "useraccount"
}


type Category struct {
	CategoryID   uint      `gorm:"primaryKey" json:"category_id"`
	Name         string    `gorm:"unique;not null" json:"name"`
	Description  string    `json:"description"`
	Status       string    `gorm:"default:'active'" json:"status"`
	CreatedBy    string    `json:"created_by"` // admin username
	CreatedAt    time.Time `json:"created_at"`

	// Relationship
	SubCategories []SubCategory `gorm:"foreignKey:CategoryID" json:"sub_categories"`
}

func (Category) TableName() string {
	return "categories"
}
type SubCategory struct {
	SubCategoryID uint      `gorm:"primaryKey" json:"sub_category_id"`
	CategoryID    uint      `gorm:"not null" json:"category_id"` // FK to Category
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `json:"description"`
	Status        string    `gorm:"default:'active'" json:"status"`
	CreatedBy     string    `json:"created_by"` // admin username
	CreatedAt     time.Time `json:"created_at"`

	// Optional: preload parent
	Category Category `gorm:"foreignKey:CategoryID" json:"category"`
}

func (SubCategory) TableName() string {
	return "sub_categories"
}