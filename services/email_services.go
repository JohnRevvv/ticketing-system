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

	subject := "🔔 New Ticket For Endorsement 🔔"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f6f8;
      margin: 0;
      padding: 20px;
    }
    .container {
      max-width: 600px;
      margin: auto;
      background: #ffffff;
      padding: 25px;
      border-radius: 10px;
      box-shadow: 0 4px 10px rgba(0,0,0,0.08);
    }
    .header {
      text-align: center;
      padding-bottom: 10px;
      border-bottom: 1px solid #eee;
    }
    .header h2 {
      color: #2c3e50;
      margin: 0;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #3498db;
    }
    .label {
      font-weight: bold;
      color: #555;
    }
    .value {
      color: #222;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
    .note {
      font-size: 12px;
      color: #999;
      margin-top: 5px;
    }
  </style>
</head>
<body>
  <div class="container">
    
    <div class="header">
      <h2>New Ticket For Endorsement</h2>
      <p>You have received a ticket that requires your action.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Ticket ID:</span> %s</p>
      <p><span class="label">Subject:</span> %s</p>
      <p><span class="label">Category:</span> %s</p>
      <p><span class="label">Priority:</span> %s</p>
      <p><span class="label">Submitted By:</span> %s</p>
    </div>

    <div class="footer">
      <p><b>Note:</b> This message is auto-generated.</p>
      <p>Please do not reply to this email.</p>
      <div class="note">If you have concerns, contact the system administrator.</div>
    </div>

  </div>
</body>
</html>
`, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Username)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send endorser email:", err)
	}

	return err
}

// SendApproverNotification — called after EndorseTicket succeeds
func SendApproverNotification(ticket models.CreateTicket, approverUsername string, toEmail string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "🔔 Ticket Ready for Approval 🔔"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f6f8;
      margin: 0;
      padding: 20px;
    }
    .container {
      max-width: 600px;
      margin: auto;
      background: #ffffff;
      padding: 25px;
      border-radius: 10px;
      box-shadow: 0 4px 10px rgba(0,0,0,0.08);
    }
    .header {
      text-align: center;
      padding-bottom: 10px;
      border-bottom: 1px solid #eee;
    }
    .header h2 {
      color: #2c3e50;
      margin: 0;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #f39c12;
    }
    .label {
      font-weight: bold;
      color: #555;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
    .note {
      font-size: 12px;
      color: #999;
      margin-top: 5px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>Ticket Ready for Approval</h2>
      <p>A ticket has been endorsed and requires your action.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> %s</p>
      <p><span class="label">Ticket ID:</span> %s</p>
      <p><span class="label">Subject:</span> %s</p>
      <p><span class="label">Category:</span> %s</p>
      <p><span class="label">Priority:</span> %s</p>
      <p><span class="label">Endorsed By:</span> %s</p>
    </div>

    <div class="footer">
      <p><b>Note:</b> This message is auto-generated.</p>
      <p>Please do not reply to this email.</p>
      <div class="note">For concerns, contact your system administrator.</div>
    </div>

  </div>

</body>
</html>
`, approverUsername, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Endorser)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" + // ✅ FIXED
			body,
	)

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

	subject := "🔔 Ticket Available for Resolution 🔔"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f6f8;
      margin: 0;
      padding: 20px;
    }
    .container {
      max-width: 600px;
      margin: auto;
      background: #ffffff;
      padding: 25px;
      border-radius: 10px;
      box-shadow: 0 4px 10px rgba(0,0,0,0.08);
    }
    .header {
      text-align: center;
      padding-bottom: 10px;
      border-bottom: 1px solid #eee;
    }
    .header h2 {
      color: #2c3e50;
      margin: 0;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #27ae60;
    }
    .label {
      font-weight: bold;
      color: #555;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
    .note {
      font-size: 12px;
      color: #999;
      margin-top: 5px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>Ticket Ready for Resolution</h2>
      <p>A new ticket is now assigned for action.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> %s</p>
      <p><span class="label">Ticket ID:</span> %s</p>
      <p><span class="label">Subject:</span> %s</p>
      <p><span class="label">Category:</span> %s</p>
      <p><span class="label">Priority:</span> %s</p>
      <p><span class="label">Approved By:</span> %s</p>
    </div>

    <div class="footer">
      <p><b>Note:</b> This message is auto-generated.</p>
      <p>Please do not reply to this email.</p>
      <div class="note">For concerns, contact your system administrator.</div>
    </div>

  </div>

</body>
</html>
`, resolverUsername, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Approver)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

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
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f6f8;
      margin: 0;
      padding: 20px;
    }
    .container {
      max-width: 600px;
      margin: auto;
      background: #ffffff;
      padding: 25px;
      border-radius: 10px;
      box-shadow: 0 4px 10px rgba(0,0,0,0.08);
    }
    .header {
      text-align: center;
      padding-bottom: 10px;
      border-bottom: 1px solid #eee;
    }
    .header h2 {
      color: #155724;
      margin: 0;
    }
    .badge {
      display: inline-block;
      background-color: #d4edda;
      color: #155724;
      font-size: 16px;
      font-weight: bold;
      padding: 10px 18px;
      border-radius: 8px;
      margin: 20px 0;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #28a745;
    }
    .label {
      font-weight: bold;
      color: #555;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
    .note {
      font-size: 12px;
      color: #999;
      margin-top: 5px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>Ticket Resolved Successfully</h2>
      <p>Good news! Your ticket has been completed.</p>
    </div>

    <div style="text-align:center;">
      <div class="badge">Resolved ✓</div>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> %s</p>
      <p><span class="label">Ticket ID:</span> %s</p>
      <p><span class="label">Subject:</span> %s</p>
      <p><span class="label">Category:</span> %s</p>
      <p><span class="label">Priority:</span> %s</p>
      <p><span class="label">Resolved By:</span> %s</p>
    </div>

    <div class="footer">
      <p><b>Note:</b> This message is auto-generated.</p>
      <p>Please do not reply to this email.</p>
      <div class="note">If you have further concerns, please contact the resolver.</div>
    </div>

  </div>

</body>
</html>
`, submitterUsername, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, ticket.Assignee)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send resolved email:", err)
	}

	return err
}
