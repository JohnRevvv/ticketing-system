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

	// HTML body with centered, bold code
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f5f5f5;
      color: #333333;
      padding: 20px;
    }
    .container {
      background-color: #ffffff;
      padding: 30px;
      border-radius: 10px;
      text-align: center;
      max-width: 500px;
      margin: auto;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    }
    .code {
      display: inline-block;
      font-size: 32px;
      font-weight: bold;
      margin: 20px 0;
      padding: 15px 25px;
      background-color: #e0e0e0;
      border-radius: 8px;
      letter-spacing: 5px;
    }
    .footer {
      font-size: 14px;
      color: #666666;
      margin-top: 20px;
    }
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

	// Include Content-Type header for HTML emails
	msg := []byte("Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	to := []string{toEmail}

	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)
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

	// HTML body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f5f5f5;
      color: #333333;
      padding: 20px;
    }
    .container {
      background-color: #ffffff;
      padding: 30px;
      border-radius: 10px;
      text-align: center;
      max-width: 500px;
      margin: auto;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    }
    .success {
      display: inline-block;
      font-size: 28px;
      font-weight: bold;
      margin: 20px 0;
      padding: 15px 25px;
      background-color: #d4edda;
      color: #155724;
      border-radius: 8px;
    }
    .footer {
      font-size: 14px;
      color: #666666;
      margin-top: 20px;
    }
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

	// Include Content-Type header for HTML emails
	msg := []byte("Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	to := []string{toEmail}

	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)
	if err != nil {
		log.Println("Failed to send password reset success email:", err)
	}

	return err
}

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