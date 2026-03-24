package models

type AdminAccount struct {
	AdminID  uint   `json:"admin_id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"password"`
}

// Table name
func (AdminAccount) TableName() string {
	return "adminaccount"
}