package models

import "time"

type CreateTicket struct {
	TicketID    string `json:"ticket_id" gorm:"primaryKey"` // <-- change from uint to string
	Username    string `json:"username"`
	Subject     string `json:"subject"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	Assignee    string `json:"assignee"`
	Endorser    string `json:"endorser"`
	Approver    string `json:"approver"`
	Remarks     string `json:"remarks"`
	Status      string `json:"status" gorm:"default:'for endorsement'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CancelledBy string     `json:"cancelled_by"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

// Table name
func (CreateTicket) TableName() string {
	return "tickets"
}
