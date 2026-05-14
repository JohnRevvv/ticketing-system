package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"ticketing-be-dev/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

	err := middleware.DBConn.
		Where("email = ?", body.Email).
		First(&user).Error

	// ❌ Email NOT found
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "No email registered",
		})
	}

	// ✅ Generate 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// ✅ Delete old OTPs
	middleware.DBConn.
		Where("user_id = ?", user.UserID).
		Delete(&models.PasswordResetToken{})

	// ✅ Create new OTP
	resetToken := models.PasswordResetToken{
		UserID:    user.UserID,
		Token:     otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := middleware.DBConn.Create(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create reset token",
		})
	}

	// ✅ Send email ONCE (async)
	go func() {
		if err := services.SendResetPasswordEmail(user.Email, otp); err != nil {
			log.Println("Email sending failed:", err)
		}
	}()

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Verification code has been sent to your email",
		Data: fiber.Map{
			"reset_token": otp, // ⚠️ remove in production
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
func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString
	hasSpecial := regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};:'",.<>\/?\\|]`).MatchString

	return hasNumber(password) && hasSpecial(password)
}

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

	// ✅ PASSWORD VALIDATION (added)
	if !validatePassword(body.NewPassword) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Password must be at least 8 characters long and include at least 1 number and 1 special character",
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

	// Get user email
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, resetToken.UserID).Error; err != nil {
		log.Println("Failed to fetch user for email:", err)
	} else if user.Email != "" {
		go services.SendPasswordResetSuccessEmail(user.Email, user.Username)
	}

	middleware.DBConn.Delete(&resetToken)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Password reset successful",
	})
}
