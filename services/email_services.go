package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"ticketing-be-dev/models"
)

func SendResetPasswordEmail(toEmail string, code string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "Ticket System Password Reset Code"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; color: #333333; padding: 20px; }
    .container { background-color: #ffffff; padding: 30px; border-radius: 10px; text-align: center; max-width: 500px; margin: auto; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
    .code { display: inline-block; font-size: 32px; font-weight: bold; margin: 20px 0; padding: 15px 25px; background-color: #e0e0e0; border-radius: 8px; letter-spacing: 5px; }
    .footer { font-size: 14px; color: #666666; margin-top: 20px; }
  </style>
</head>
<body>
  <div class="container">
    <h2>Reset Password Code</h2>
    <p>Your verification code is:</p>
    <div class="code">%s</div>
    <p>This code will expire in 15 minutes.</p>
    <p class="footer">If you did not request this, please ignore this email.<br>Thank you.</p>
  </div>
</body>
</html>
`, code)

	msg := []byte("Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send reset email:", err)
	}
	return err
}

func SendPasswordResetSuccessEmail(toEmail string, username string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "✅ Password Reset Successful"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; color: #333333; padding: 20px; }
    h2 { font-weight: bold; }
    .container { background-color: #ffffff; padding: 30px; border-radius: 10px; text-align: center; max-width: 500px; margin: auto; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
    .success { display: inline-block; font-size: 28px; font-weight: bold; margin: 20px 0; padding: 15px 25px; background-color: #d4edda; color: #155724; border-radius: 8px; }
    .footer { font-size: 14px; color: #666666; margin-top: 20px; }
  </style>
</head>
<body>
  <div class="container">
    <h2>Password Reset Successful</h2>
    <p>Hello %s,</p>
    <div class="success">Your password has been successfully reset!</div>
    <p>You can now log in using your new password.</p>
    <p class="footer">If you did not perform this action, please contact support immediately.</p>
  </div>
</body>
</html>
`, username)

	msg := []byte("Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send password reset success email:", err)
	}
	return err
}

// SendEndorserNotification — called on ticket creation
func SendEndorserNotification(ticket models.CreateTicket, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 New Ticket For Endorsement! 🔔"
	body := fmt.Sprintf(`
Hello %s,

You have a new ticket that requires your endorsement.

---------------------------------------
Ticket ID  : %s
Subject    : %s
Category   : %s
Priority   : %s
Submitted by: %s
---------------------------------------

Please log in to the system to review and endorse this ticket.

Thank you.
`, ticket.Endorser, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Username)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send endorser email:", err)
	}
	return err
}

// SendApproverNotification — called after EndorseTicket succeeds
func SendApproverNotification(ticket models.CreateTicket, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 Ticket Ready for Your Approval 🔔"
	body := fmt.Sprintf(`
Hello %s,

A ticket has been endorsed and now requires your approval.

---------------------------------------
Ticket ID  : %s
Subject    : %s
Category   : %s
Priority   : %s
Endorsed by: %s
---------------------------------------

Please log in to the system to review and approve this ticket.

Thank you.
`, ticket.Approver, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Endorser)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send approver email:", err)
	}
	return err
}

// SendResolverNotification — called after ApproveTicket succeeds.
// Sent to ALL resolvers. resolverUsername passed separately because
// ticket.Assignee is still empty at approval time.
func SendResolverNotification(ticket models.CreateTicket, resolverUsername string, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 Ticket Available for Resolution"
	body := fmt.Sprintf(`
Hello %s,

A ticket has been approved and is now available for you to grab and resolve.

---------------------------------------
Ticket ID  : %s
Subject    : %s
Category   : %s
Priority   : %s
Approved by: %s
---------------------------------------

Please log in to the system and grab this ticket to start working on it.

Thank you.
`, resolverUsername, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Approver)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send resolver email:", err)
	}
	return err
}

// SendTicketResolvedEmail — called after ResolveTicket succeeds.
// Notifies the person who originally filed the ticket that it is now resolved.
func SendTicketResolvedEmail(ticket models.CreateTicket, submitterUsername string, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "✅ Your Ticket Has Been Resolved"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; background-color: #f5f5f5; color: #333333; padding: 20px; }
    .container { background-color: #ffffff; padding: 30px; border-radius: 10px; max-width: 500px; margin: auto; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
    h2 { color: #155724; }
    .badge { display: inline-block; background-color: #d4edda; color: #155724; font-size: 18px; font-weight: bold; padding: 10px 20px; border-radius: 8px; margin: 16px 0; }
    .details { background-color: #f8f9fa; border-left: 4px solid #28a745; padding: 12px 16px; border-radius: 4px; margin: 16px 0; }
    .details p { margin: 6px 0; font-size: 14px; }
    .label { color: #666666; font-weight: bold; }
    .footer { font-size: 13px; color: #888888; margin-top: 24px; }
  </style>
</head>
<body>
  <div class="container">
    <h2>✅ Ticket Resolved</h2>
    <p>Hello <strong>%s</strong>,</p>
    <p>Great news! Your ticket has been successfully resolved.</p>

    <div class="badge">Resolved ✓</div>

    <div class="details">
      <p><span class="label">Ticket ID  :</span> %s</p>
      <p><span class="label">Subject    :</span> %s</p>
      <p><span class="label">Category   :</span> %s</p>
      <p><span class="label">Priority   :</span> %s</p>
      <p><span class="label">Resolved by:</span> %s</p>
    </div>

    <p>If you have any further concerns, please feel free to file a new ticket.</p>
    <p class="footer">Thank you for using our support system.</p>
  </div>
</body>
</html>
`, submitterUsername, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Assignee)

	msg := []byte("Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send resolved email to submitter:", err)
	}
	return err
}
