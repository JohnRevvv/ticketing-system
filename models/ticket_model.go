package models

import "time"

type CreateTicket struct {
	TicketID          string     `json:"ticket_id"          gorm:"primaryKey"`
	Username          string     `json:"username"`
	Category          string     `json:"category"`
	Subject           string     `json:"subject"`
	Institution       string     `json:"institution"`
	Tickettype        string     `json:"tickettype"`
	Description       string     `json:"description"`
	Priority          string     `json:"priority"`
	Assignee          string     `json:"assignee"`
	Endorser          string     `json:"endorser"`
	Approver          string     `json:"approver"`
	Status            string     `json:"status"             gorm:"default:'for endorsement'"`
	CreatedAt         time.Time  `json:"created_at"         gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at"         gorm:"autoUpdateTime"`
	CancelledBy       string     `json:"cancelled_by"`
	CancelledAt       *time.Time `json:"cancelled_at"`
	StartedAt         *time.Time `json:"started_at"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	ResolutionMinutes float64    `json:"resolution_minutes"`
	ResolutionTime    string     `json:"resolution_time" gorm:"column:resolution_time;default:''"`
	OnHold            bool       `json:"onhold" gorm:"column:on_hold;default:false"`
	HoldStartedAt     *time.Time `json:"hold_started_at"`
	TotalHoldSeconds  float64    `json:"total_hold_seconds"`
}

func (CreateTicket) TableName() string {
	return "tickets"
}

// ── TicketAttachment ──────────────────────────────────────────────────────────

type TicketAttachment struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TicketID   string `json:"ticket_id"`
	FileName   string `json:"file_name"`
	FileURL   string `json:"file_url"`
	UploadedBy string `json:"uploaded_by"`
}

func (TicketAttachment) TableName() string {
	return "ticket_attachments"
}

// ── TicketRemark ──────────────────────────────────────────────────────────────

type TicketRemark struct {
	RemarkID  string    `gorm:"primaryKey" json:"remark_id"`
	TicketID  string    `json:"ticket_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (TicketRemark) TableName() string {
	return "ticketremark"
}
