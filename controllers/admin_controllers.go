package controllers

import (
	"strings"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"

	"github.com/gofiber/fiber/v2"
)


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