package services

import (
	"log"
	"time"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"

	"github.com/robfig/cron/v3"
)

func StartTicketAutoCloser() {
	c := cron.New()

	_, err := c.AddFunc("@daily", func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("AutoCloser panic recovered:", r)
			}
		}()

		AutoCloseResolvedTickets()
	})

	if err != nil {
		log.Fatal("Failed to start cron:", err)
	}

	c.Start()
}

func AutoCloseResolvedTickets() {
	var tickets []models.CreateTicket

	// 🕒 7 days cutoff
	cutoff := time.Now().Add(-10 * time.Minute)
	// cutoff := time.Now().AddDate(0, 0, -7)

	// 🔍 Find resolved tickets older than 7 days and not yet closed
	if err := middleware.DBConn.
		Where("status = ? AND updated_at <= ?", "resolved", cutoff).
		Find(&tickets).Error; err != nil {
		log.Println("Failed to fetch resolved tickets:", err)
		return
	}

	now := time.Now()

	for _, ticket := range tickets {
		err := middleware.DBConn.Model(&models.CreateTicket{}).
			Where("ticket_id = ?", ticket.TicketID).
			Updates(map[string]interface{}{
				"status":    "closed",
				"closed_at": now,
			}).Error

		if err != nil {
			log.Println("Failed to auto-close ticket:", ticket.TicketID, err)
			continue
		}

		log.Println("Auto-closed ticket:", ticket.TicketID)
	}
}