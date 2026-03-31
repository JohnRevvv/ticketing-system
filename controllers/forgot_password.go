package controllers

import (
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

    // Check if email exists
    if err := middleware.DBConn.
        Where("email = ?", body.Email).
        First(&user).Error; err != nil {

        // ⚠️ Do NOT reveal if email exists (security best practice)
        return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
            RetCode: "200",
            Message: "If the email exists, a reset link has been sent",
        })
    }

    // Generate token
    token := uuid.New().String()

    resetToken := models.PasswordResetToken{
        UserID:    user.UserID,
        Token:     token,
        ExpiresAt: time.Now().Add(15 * time.Minute), // expires in 15 mins
    }

    // Save token
    if err := middleware.DBConn.Create(&resetToken).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to create reset token",
        })
    }

    // 🔥 For now: return token (later you send via email)
    return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Reset token generated",
        Data: fiber.Map{
            "reset_token": token,
        },
    })
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

    var resetToken models.PasswordResetToken

    // Find token
    if err := middleware.DBConn.
        Where("token = ?", body.Token).
        First(&resetToken).Error; err != nil {

        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Invalid or expired token",
        })
    }

    // Check expiry
    if time.Now().After(resetToken.ExpiresAt) {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Token expired",
        })
    }

    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to hash password",
        })
    }

    // Update user password
    if err := middleware.DBConn.
        Model(&models.UserAccount{}).
        Where("user_id = ?", resetToken.UserID).
        Update("password", string(hashedPassword)).Error; err != nil {

        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to update password",
        })
    }

    // Delete used token
    middleware.DBConn.Delete(&resetToken)

    return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Password reset successful",
    })
}