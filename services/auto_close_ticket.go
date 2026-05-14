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

	// ✅ Every minute
	// _, err := c.AddFunc("* * * * *", func() {

	// every day
	_, err := c.AddFunc("0 0 * * *", func() {

		log.Println("================================================")
		log.Println("Running auto closer job...")
		log.Println("Current time:", time.Now())

		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered panic:", r)
			}
		}()

		AutoCloseResolvedTickets()
	})

	if err != nil {
		log.Fatal("Cron error:", err)
	}

	c.Start()

	log.Println("✅ Ticket auto closer started")
}

func AutoCloseResolvedTickets() {

	var tickets []models.CreateTicket

	// ✅ 1 minute ago
	// cutoff := time.Now().Add(-1 * time.Minute)

	// 7 days
	cutoff := time.Now().AddDate(0, 0, -7)

	log.Println("Cutoff time:", cutoff)

	// ✅ TEMPORARY TEST
	// remove updated_at first to verify logic works
	err := middleware.DBConn.
		Where("status = ?", "resolved").
		Find(&tickets).Error

	if err != nil {
		log.Println("DB query failed:", err)
		return
	}

	log.Println("Resolved tickets found:", len(tickets))

	now := time.Now()

	for _, ticket := range tickets {

		log.Println("Closing ticket:", ticket.TicketID)

		err := middleware.DBConn.
			Model(&models.CreateTicket{}).
			Where("ticket_id = ?", ticket.TicketID).
			Updates(map[string]interface{}{
				"status":    "closed",
				"closed_at": now,
			}).Error

		if err != nil {
			log.Println("Failed closing ticket:", ticket.TicketID, err)
			continue
		}

		log.Println("✅ Ticket closed:", ticket.TicketID)
	}
}