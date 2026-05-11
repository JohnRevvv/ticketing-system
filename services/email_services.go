package services

import (
	"encoding/base64"
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
    .btn { display: inline-block; margin-top: 20px; padding: 12px 25px; font-size: 16px;color: #ffffff; background-color: #007bff; text-decoration: none; border-radius: 6px; }
    .btn:hover { background-color: #0056b3; }
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

    <!-- LOGIN BUTTON -->
    <a href="https://ideyanale.bakawan-ai.com/login" class="btn">Go to Login</a>

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
    .btn {
  display: block;
  text-align: center;
  margin: 25px auto 10px;
  padding: 12px;
  font-size: 16px;
  color: #ffffff !important;
  background-color: #007bff;
  text-decoration: none !important;
  border-radius: 6px;
  width: 200px;
}
    .btn:hover {
      background-color: #0056b3;
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

    <!-- LOGIN BUTTON -->
    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">Go to Login</a>

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
func SendApproverNotification(ticket models.CreateTicket, toEmail string, fullName string) error {
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
    .btn {
  display: block;
  text-align: center;
  margin: 25px auto 10px;
  padding: 12px;
  font-size: 16px;
  color: #ffffff !important;
  background-color: #007bff;
  text-decoration: none !important;
  border-radius: 6px;
  width: 200px;
}
    .btn:hover {
      background-color: #0056b3;
}
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
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

    <!-- LOGIN BUTTON -->
    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">Go to Login</a>

    <div class="footer">
      <p><b>Note:</b> This message is auto-generated.</p>
      <p>Please do not reply to this email.</p>
    </div>
  </div>

</body>
</html>
`,
		fullName, // ✅ Hello name
		ticket.TicketID,
		ticket.Subject,
		ticket.Category,
		ticket.Priority,
		ticket.Endorser,
	)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(smtpHost+":587", auth, from, []string{toEmail}, msg)
	if err != nil {
		log.Println("Failed to send approver email:", err)
	}

	return err
}

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
  <meta charset="UTF-8">
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
      border: 1px solid #eaeaea;
    }
    .header {
      text-align: center;
      border-bottom: 1px solid #eee;
      padding-bottom: 15px;
    }
    .header h2 {
      margin: 0;
      color: #2c3e50;
    }
    .header p {
      color: #777;
      font-size: 14px;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #007bff;
    }
    .label {
      font-weight: bold;
      color: #333;
    }
    .value {
      color: #555;
    }
    .priority {
      font-weight: bold;
      color: #e74c3c;
    }
    .btn {
  display: block;
  text-align: center;
  margin: 25px auto 10px;
  padding: 12px;
  font-size: 16px;
  color: #ffffff !important;
  background-color: #007bff;
  text-decoration: none !important;
  border-radius: 6px;
  width: 200px;
}
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>🔔 Ticket Ready for Resolution</h2>
      <p>A new ticket has been assigned to you.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> <span class="value">%s</span></p>
      <p><span class="label">Ticket ID:</span> <span class="value">%s</span></p>
      <p><span class="label">Subject:</span> <span class="value">%s</span></p>
      <p><span class="label">Category:</span> <span class="value">%s</span></p>
      <p><span class="label">Priority:</span> <span class="priority">%s</span></p>
      <p><span class="label">Approved By:</span> <span class="value">%s</span></p>
    </div>

    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">Open Ticket System</a>

    <div class="footer">
      <p><b>Note:</b> This is an automated message.</p>
      <p>Please do not reply to this email.</p>
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

// ==============================
// Submitter email notifications
// ==============================

func SendEndorsedNotification(ticket models.CreateTicket, submitterName string, toEmail string, endorserName string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "📌 Your Ticket Has Been Endorsed"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
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
      border: 1px solid #eaeaea;
    }
    .header {
      text-align: center;
      border-bottom: 1px solid #eee;
      padding-bottom: 15px;
    }
    .header h2 {
      margin: 0;
      color: #2c3e50;
    }
    .header p {
      color: #777;
      font-size: 14px;
    }
    .ticket-box {
      background: #f9fafb;
      padding: 15px;
      margin-top: 20px;
      border-radius: 8px;
      border-left: 5px solid #007bff;
    }
    .label {
      font-weight: bold;
      color: #333;
    }
    .value {
      color: #555;
    }
    .status {
      font-weight: bold;
      color: #007bff;
    }
    .btn {
      display: block;
      text-align: center;
      margin: 25px auto 10px;
      padding: 12px;
      font-size: 16px;
      color: #ffffff !important;
      background-color: #007bff;
      text-decoration: none !important;
      border-radius: 6px;
      width: 220px;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>📌 Ticket Endorsed</h2>
      <p>Your ticket has been endorsed and forwarded for approval.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> <span class="value">%s</span></p>
      <p><span class="label">Ticket ID:</span> <span class="value">%s</span></p>
      <p><span class="label">Subject:</span> <span class="value">%s</span></p>
      <p><span class="label">Category:</span> <span class="value">%s</span></p>
      <p><span class="label">Priority:</span> <span class="value">%s</span></p>
      <p><span class="label">Endorsed By:</span> <span class="value">%s</span></p>
      <p><span class="label">Status:</span> <span class="status">For Approval</span></p>
    </div>

    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">
      View Ticket
    </a>

    <div class="footer">
      <p><b>Note:</b> This is an automated message.</p>
      <p>Please wait for the approver's action on your request.</p>
    </div>

  </div>

</body>
</html>
`,
		submitterName,
		ticket.TicketID,
		ticket.Subject,
		ticket.Category,
		ticket.Priority,
		endorserName,
	)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	if err != nil {
		log.Println("Failed to send endorsed email:", err)
	}

	return err
}

func SendApprovedNotification(ticket models.CreateTicket, submitterName string, toEmail string, approverName string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "✅ Your Ticket Has Been Approved"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
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
      border: 1px solid #eaeaea;
    }
    .header {
      text-align: center;
      border-bottom: 1px solid #eee;
      padding-bottom: 15px;
    }
    .header h2 {
      margin: 0;
      color: #2c3e50;
    }
    .header p {
      color: #777;
      font-size: 14px;
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
      color: #333;
    }
    .value {
      color: #555;
    }
    .status {
      font-weight: bold;
      color: #28a745;
    }
    .btn {
      display: block;
      text-align: center;
      margin: 25px auto 10px;
      padding: 12px;
      font-size: 16px;
      color: #ffffff !important;
      background-color: #28a745;
      text-decoration: none !important;
      border-radius: 6px;
      width: 220px;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>✅ Ticket Approved</h2>
      <p>Your ticket has been approved and is now waiting for assignment.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> <span class="value">%s</span></p>
      <p><span class="label">Ticket ID:</span> <span class="value">%s</span></p>
      <p><span class="label">Subject:</span> <span class="value">%s</span></p>
      <p><span class="label">Category:</span> <span class="value">%s</span></p>
      <p><span class="label">Priority:</span> <span class="value">%s</span></p>
      <p><span class="label">Approved By:</span> <span class="value">%s</span></p>
      <p><span class="label">Status:</span> <span class="status">For Assignment</span></p>
    </div>

    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">
      View Ticket
    </a>

    <div class="footer">
      <p><b>Note:</b> This is an automated message.</p>
      <p>Your request is now pending resolver assignment.</p>
    </div>

  </div>

</body>
</html>
`,
		submitterName,
		ticket.TicketID,
		ticket.Subject,
		ticket.Category,
		ticket.Priority,
		approverName,
	)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	if err != nil {
		log.Println("Failed to send approved email:", err)
	}

	return err
}

func SendResolvedNotification(ticket models.CreateTicket, submitterName string, toEmail string, resolverName string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "✅ Your Ticket Has Been Resolved"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
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
      border: 1px solid #eaeaea;
    }
    .header {
      text-align: center;
      border-bottom: 1px solid #eee;
      padding-bottom: 15px;
    }
    .header h2 {
      margin: 0;
      color: #2c3e50;
    }
    .header p {
      color: #777;
      font-size: 14px;
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
      color: #333;
    }
    .value {
      color: #555;
    }
    .status {
      font-weight: bold;
      color: #28a745;
    }
    .btn {
      display: block;
      text-align: center;
      margin: 25px auto 10px;
      padding: 12px;
      font-size: 16px;
      color: #ffffff !important;
      background-color: #28a745;
      text-decoration: none !important;
      border-radius: 6px;
      width: 220px;
    }
    .footer {
      margin-top: 25px;
      font-size: 12px;
      text-align: center;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 15px;
    }
  </style>
</head>
<body>

  <div class="container">

    <div class="header">
      <h2>✅ Ticket Resolved</h2>
      <p>Your request has been successfully completed.</p>
    </div>

    <div class="ticket-box">
      <p><span class="label">Hello:</span> <span class="value">%s</span></p>
      <p><span class="label">Ticket ID:</span> <span class="value">%s</span></p>
      <p><span class="label">Subject:</span> <span class="value">%s</span></p>
      <p><span class="label">Category:</span> <span class="value">%s</span></p>
      <p><span class="label">Priority:</span> <span class="value">%s</span></p>
      <p><span class="label">Resolved By:</span> <span class="value">%s</span></p>
      <p><span class="label">Status:</span> <span class="status">Resolved</span></p>
    </div>

    <a href="https://idiyanale.bakawan-ai.com/login" class="btn">View Ticket</a>

    <div class="footer">
      <p><b>Note:</b> This is an automated message.</p>
      <p>If the issue persists, please create a new ticket.</p>
    </div>

  </div>

</body>
</html>
`, submitterName, ticket.TicketID, ticket.Subject, ticket.Category, ticket.Priority, resolverName)

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

// ==============================
//      Account Status
// ==============================

func SendAccountApprovedNotification(toEmail string, fullName string, role string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "✅ Account Approved"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial; background:#f4f6f8; padding:20px;">

<div style="max-width:600px; margin:auto; background:#fff; padding:25px; border-radius:10px;">

<h2 style="color:#28a745;">✅ Account Approved</h2>

<p>Hello <b>%s</b>,</p>

<p>Your account has been approved successfully.</p>

<p>You may now log in to the system.</p>

<div style="background:#f9fafb; padding:15px; border-radius:8px;">
  <p><b>Role:</b> %s</p>
  <p><b>Status:</b> Approved</p>
</div>

<a href="https://idiyanale.bakawan-ai.com/login"
style="
display:inline-block;
margin-top:20px;
padding:12px 20px;
background:#28a745;
color:white;
text-decoration:none;
border-radius:6px;
">
Login Now
</a>

<p style="margin-top:30px; font-size:12px; color:#777;">
This is an automated email.
</p>

</div>

</body>
</html>
`, fullName, role)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	if err != nil {
		log.Println("Failed to send approval email:", err)
	}

	return err
}

func SendAccountRejectedNotification(toEmail string,fullName string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "❌ Account Rejected"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial; background:#f4f6f8; padding:20px;">

<div style="max-width:600px; margin:auto; background:#fff; padding:25px; border-radius:10px;">

<h2 style="color:#dc3545;">❌ Account Rejected</h2>

<p>Hello <b>%s</b>,</p>

<p>We regret to inform you that your account registration has been rejected.</p>

<p>Please contact the administrator for more information.</p>

<div style="background:#f9fafb; padding:15px; border-radius:8px;">
  <p><b>Status:</b> Rejected</p>
</div>

<p style="margin-top:30px; font-size:12px; color:#777;">
This is an automated email.
</p>

</div>

</body>
</html>
`, fullName)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	if err != nil {
		log.Println("Failed to send rejection email:", err)
	}

	return err
}

// ==============================
//      Remark notification
// ==============================

func SendTicketRemark1Notification(toEmail string, submitterName string, ticket models.CreateTicket, message string, senderUsername string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "💬 New Remark on Your Ticket"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial; background:#f4f6f8; padding:20px;">

<div style="max-width:600px; margin:auto; background:#fff; padding:25px; border-radius:10px;">

<h2 style="color:#007bff;">💬 New Ticket Remark</h2>

<p>Hello <b>%s</b>,</p>

<p>A new remark has been added to your ticket.</p>

<div style="background:#f9fafb; padding:15px; border-radius:8px;">
  <p><b>Ticket ID:</b> %s</p>
  <p><b>Subject:</b> %s</p>
  <p><b>Message:</b> %s</p>
  <p><b>Sent By:</b> %s</p>
</div>

<p style="margin-top:20px;">
Please log in to view full details.
</p>

</div>

</body>
</html>
`,
		submitterName,
		ticket.TicketID,
		ticket.Subject,
		message,
		senderUsername,
	)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	return err
}

func SendTicketRemarkNotification(toEmail string, submitterName string, ticket models.CreateTicket, message string, senderUsername string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "💬 New Remark on Your Ticket"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial; background:#f4f6f8; padding:20px;">

<div style="max-width:600px; margin:auto; background:#fff; padding:25px; border-radius:10px;">

<h2 style="color:#007bff;">New Ticket Remark</h2>

<p>Hello <b>%s</b>,</p>

<p>A new remark has been added to your ticket.</p>

<hr>

<p><b>Ticket ID:</b> %s</p>
<p><b>Subject:</b> %s</p>

<hr>

<h3 style="color:#333;">Message</h3>

<div style="
	background:#f1f3f5;
	padding:15px;
	border-radius:8px;
	font-size:16px;
	color:#222;
	border-left:4px solid #007bff;
">
	%s
</div>

<hr>

<p><b>Sent By:</b> %s</p>

<p style="margin-top:20px; font-size:12px; color:#777;">
Please log in to view full ticket details.
</p>

</div>

</body>
</html>
`,
		submitterName,
		ticket.TicketID,
		ticket.Subject,
		message,
		senderUsername,
	)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	err := smtp.SendMail(
		smtpHost+":587",
		auth,
		from,
		[]string{toEmail},
		msg,
	)

	return err
}

func encodeSubject(subject string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	return "=?UTF-8?B?" + encoded + "?="
}
