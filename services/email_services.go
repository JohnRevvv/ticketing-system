package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"ticketing-be-dev/models"
)

// SendEndorserNotification sends an email to the endorser
func SendEndorserNotification(ticket models.CreateTicket, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST") // just the host

	// Gmail default port is 587
	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 New Ticket For Endorsement! 🔔"
	body := fmt.Sprintf(`
Hello %s,

You have a new ticket that requires your endorsement.

---------------------------------------
Ticket ID: %s
Subject: %s
Category: %s
Priority: %s
Submitted by: %s
---------------------------------------

Please log in to the system to review and endorse this ticket.

Thank you.
`, ticket.Endorser, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Username)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)
	to := []string{toEmail}

	// Combine host and default port 587 directly here
	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)
	if err != nil {
		log.Println("Failed to send endorser email:", err)
	}

	return err
}


// SendApproverNotification sends an email to the approver after endorsement
func SendApproverNotification(ticket models.CreateTicket, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST") // Gmail host

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 Ticket Ready for Your Approval"
	body := fmt.Sprintf(`
Hello %s,

A ticket has been endorsed and now requires your approval.

---------------------------------------
Ticket ID: %s
Subject: %s
Category: %s
Priority: %s
Endorsed by: %s
---------------------------------------

Please log in to the system to review and approve this ticket.

Thank you.
`, ticket.Approver, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Endorser)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)
	to := []string{toEmail}

	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)
	if err != nil {
		log.Println("Failed to send approver email:", err)
	}

	return err
}

// SendResolverNotification sends an email to the resolver when a ticket is approved
func SendResolverNotification(ticket models.CreateTicket, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 Ticket Assigned for Resolution"
	body := fmt.Sprintf(`
Hello %s,

A ticket has been approved and is now assigned to you for resolution.

---------------------------------------
Ticket ID: %s
Subject: %s
Category: %s
Priority: %s
Approved by: %s
---------------------------------------

Please log in to the system to review and resolve this ticket.

Thank you.
`, ticket.Assignee, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Approver)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)
	to := []string{toEmail}

	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)
	if err != nil {
		log.Println("Failed to send resolver email:", err)
	}

	return err
}