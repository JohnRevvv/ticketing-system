package controllers

import (
	"strings"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func RegisterAdmin(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Parse request body
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Check required fields
	if body.Username == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Username and password are required",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash password",
		})
	}

	admin := models.AdminAccount{
		Username: body.Username,
		Password: string(hashedPassword),
	}

	// Save admin to DB
	if err := middleware.DBConn.Create(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to register admin",
		})
	}

	// Success response
	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Admin registered successfully",
		Data:    admin,
	})
}

func LoginAdmin(c *fiber.Ctx) error {
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

	if body.Username == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Username and password are required",
		})
	}

	var admin models.AdminAccount
	if err := middleware.DBConn.Where("username = ?", body.Username).First(&admin).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username or password",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username or password",
		})
	}

	// ✅ Generate JWT
	token, err := middleware.GenerateJWT(admin.AdminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Login successful",
		Data: fiber.Map{
			"admin_id": admin.AdminID,
			"username": admin.Username,
			"token":    token, // ✅ IMPORTANT
		},
	})
}

// Allowed values for Role and Status
var validRoles = []string{"approver", "endorser", "enduser", "resolver"}
var validStatuses = []string{"active", "inactive"}

// UpdateUserRoleStatus allows admin to update a user's role and status
func UpdateUserRoleStatus(c *fiber.Ctx) error {
	userID := c.Params("id") // UserID from URL param

	// Parse request body
	var body struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Validate Role
	if body.Role != "" && !contains(validRoles, strings.ToLower(body.Role)) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid role value",
		})
	}

	// Validate Status
	if body.Status != "" && !contains(validStatuses, strings.ToLower(body.Status)) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid status value",
		})
	}

	// Find the user
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
		})
	}

	// Update fields if provided
	if body.Role != "" {
		user.Role = strings.ToLower(body.Role)
	}
	if body.Status != "" {
		user.Status = strings.ToLower(body.Status)
	}

	// Save changes
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User updated successfully",
		Data:    user,
	})
}

// Helper function
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}