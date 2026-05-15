package models

import "time"

type UserAccount struct {
	UserID   uint   `gorm:"primaryKey" json:"user_id"`
	StaffID  string `gorm:"unique;not null;index" json:"staff_id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`

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

// ============================================
// RESERVED FOR PHASE 2!!
// ============================================

type Institution struct {
	InstitutionID uint   `gorm:"primaryKey" json:"institution_id"`
	Name          string `gorm:"unique;not null" json:"name"`
	Description   string `json:"description"`
	Status        string `gorm:"default:'active';not null" json:"status"`

	CreatedBy uint      `gorm:"not null" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Institution) TableName() string {
	return "institution"
}
