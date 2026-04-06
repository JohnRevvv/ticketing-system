package controllers

import (
	"crypto/tls"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// -----------------------------
// FORGOT PASSWORD: SEND CODE
// -----------------------------
func ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var user models.UserAccount
	if err := middleware.DBConn.
		Where("email = ?", body.Email).
		First(&user).Error; err != nil {

		// Don't reveal email existence
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "If the email exists, a verification code has been sent",
		})
	}

	// Generate token
	token := uuid.New().String()
	resetToken := models.PasswordResetToken{
		UserID:    user.UserID,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := middleware.DBConn.Create(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create reset token",
		})
	}

	// Send email async
	go func() {
		m := gomail.NewMessage()
		m.SetHeader("From", "no-reply@example.com") // sender
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Ticket System Password Reset Code")
		m.SetBody("text/html",
			"Hello,<br><br>"+
				"Your verification code is: <b>"+token+"</b><br>"+
				"It expires in 15 minutes.<br><br>"+
				"If you didn't request this, ignore this email.<br><br>"+
				"Thanks,<br>Ticket System Team",
		)

		d := gomail.NewDialer("smtp.gmail.com", 587, "yourgmail@gmail.com", "your-app-password")
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		if err := d.DialAndSend(m); err != nil {
			println("Email sending failed:", err.Error())
		}
	}()

	// For testing/dev only, include token in response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "If the email exists, a verification code has been sent",
		Data: fiber.Map{
			"reset_token": token, // remove in production
		},
	})
}

// -----------------------------
// VERIFY CODE
// -----------------------------
func VerifyCode(c *fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var resetToken models.PasswordResetToken
	if err := middleware.DBConn.
		Where("token = ?", body.Code).
		First(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid or expired code",
		})
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Code expired",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Code verified",
		Data: fiber.Map{
			"token": resetToken.Token,
		},
	})
}

// -----------------------------
// RESET PASSWORD
// -----------------------------
func ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var resetToken models.PasswordResetToken
	if err := middleware.DBConn.
		Where("token = ?", body.Token).
		First(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid or expired token",
		})
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Token expired",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash password",
		})
	}

	if err := middleware.DBConn.
		Model(&models.UserAccount{}).
		Where("user_id = ?", resetToken.UserID).
		Update("password", string(hashedPassword)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update password",
		})
	}

	middleware.DBConn.Delete(&resetToken)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Password reset successful",
	})
}
