package controllers

import (
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser registers a new user with default status "active"
func RegisterUser(c *fiber.Ctx) error {
	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Position  string `json:"position"`
		Email     string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash password",
		})
	}

	user := models.UserAccount{
		Username:  body.Username,
		Password:  string(hashedPassword),
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Position:  body.Position,
		Role:      "end-user",
		Status:    "pending", // default status
	}

	if err := middleware.DBConn.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "User registered successfully",
		Data:    user,
	})
}

// LoginUser logs in only if status is active
func LoginUser(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var user models.UserAccount
	if err := middleware.DBConn.Where("username = ?", body.Username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username or password",
		})
	}

	// Check if user is active
	if user.Status != "active" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "User account is for approval. Wait for the admin to approve your account.",
		})
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username or password",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Login successful",
		Data: fiber.Map{
			"user_id":   user.UserID,
			"username":  user.Username,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"email":     user.Email,
			"position":  user.Position,
			"role":      user.Role,
			"status":    user.Status,
		},
	})
}
