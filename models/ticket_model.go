package models

import "time"

type CreateTicket struct {
	TicketID    string `json:"ticket_id" gorm:"primaryKey"`
	Username    string `json:"username"`
	Category    string `json:"category"`
	Subject     string `json:"subject"`
	Institution string `json:"institution"`
	Tickettype  string `json:"tickettype"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Assignee    string `json:"assignee"`
	Endorser    string `json:"endorser"`
	Approver    string `json:"approver"`
	Status      string `json:"status" gorm:"default:'for endorsement'"`

	CreatedAt   time.Time
	UpdatedAt   time.Time

	CancelledBy string     `json:"cancelled_by"`
	CancelledAt *time.Time `json:"cancelled_at"`

	ResolvedAt  *time.Time `json:"resolved_at"` // ✅ add this
}

// Table name
func (CreateTicket) TableName() string {
	return "tickets"
}

type TicketAttachment struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TicketID   string `json:"ticket_id"` // same ticket ID as CreateTicket
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`   // saved in backend folder
	UploadedBy string `json:"uploaded_by"` // username or userID
}

func (TicketAttachment) TableName() string {
	return "ticket_attachments"
}

type TicketRemark struct {
    RemarkID  string    `gorm:"primaryKey"`
    TicketID  string
    UserID    string
    Message   string
    CreatedAt time.Time
}

func (TicketRemark) TableName() string {
	return "ticketremark"
}