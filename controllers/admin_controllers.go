package controllers

import (
	"strings"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"ticketing-be-dev/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Allowed values for Role and Status
var validRoles = []string{"approver", "endorser", "user", "resolver"}
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

	// ✅ Store old status before update
	oldStatus := user.Status

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

	// ✅ Send email notification if status changed
	if oldStatus != user.Status {

		fullName := user.FirstName + " " + user.LastName

		switch user.Status {

		case "approved":

			if user.Email != "" {
				go services.SendAccountApprovedNotification(
					user.Email,
					fullName,
					user.Role,
				)
			}

		case "rejected":

			if user.Email != "" {
				go services.SendAccountRejectedNotification(
					user.Email,
					fullName,
				)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User updated successfully",
		Data:    user,
	})
}

// UpdateUserProfile allows admin to update a user's full profile
func UpdateUserProfile(c *fiber.Ctx) error {
	userID := c.Params("id")

	var body struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Institution string `json:"institution"`
		Position   string `json:"position"`
		Role       string `json:"role"`
		Status     string `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
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

	// ✅ Store old status before update
	oldStatus := user.Status

	// Check email uniqueness if changed
	if body.Email != "" && body.Email != user.Email {
		var existing models.UserAccount

		if err := middleware.DBConn.
			Where("email = ?", body.Email).
			First(&existing).Error; err == nil {

			return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
				RetCode: "409",
				Message: "Email already exists",
			})
		}

		user.Email = body.Email
	}

	// Update fields if provided
	if body.FirstName != "" {
		user.FirstName = body.FirstName
	}

	if body.LastName != "" {
		user.LastName = body.LastName
	}

	if body.Institution != "" {
		user.Institution = body.Institution
	}

	if body.Position != "" {
		user.Position = body.Position
	}

	if body.Role != "" && contains(validRoles, strings.ToLower(body.Role)) {
		user.Role = strings.ToLower(body.Role)
	}

	if body.Status != "" && contains(validStatuses, strings.ToLower(body.Status)) {
		user.Status = strings.ToLower(body.Status)
	}

	// Hash new password if provided
	if body.Password != "" {

		hashed, err := bcrypt.GenerateFromPassword(
			[]byte(body.Password),
			bcrypt.DefaultCost,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to hash password",
			})
		}

		user.Password = string(hashed)
	}

	// Save changes
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update user",
		})
	}

	// ✅ Send status notification only if status changed
	if oldStatus != user.Status {

		fullName := user.FirstName + " " + user.LastName

		switch user.Status {

		case "approved":

			if user.Email != "" {
				go services.SendAccountApprovedNotification(
					user.Email,
					fullName,
					user.Role,
				)
			}

		case "rejected":

			if user.Email != "" {
				go services.SendAccountRejectedNotification(
					user.Email,
					fullName,
				)
			}
		}
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


// ============================================
// CATEGORY FUNCTIONS!
// ============================================

func AddCategory(c *fiber.Ctx) error {
	var input models.Category

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Category name is required",
		})
	}

	// ✅ Check if category already exists
	var existing models.Category
	if err := middleware.DBConn.Where("name = ?", input.Name).First(&existing).Error; err == nil {
		return c.Status(409).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Category already exists",
		})
	}

	input.CreatedAt = time.Now()

	if err := middleware.DBConn.Create(&input).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create category",
		})
	}

	return c.Status(201).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Category created successfully",
		Data:    input,
	})
}

func AddSubCategory(c *fiber.Ctx) error {
	var input models.SubCategory

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if input.Name == "" || input.CategoryID == 0 {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "CategoryID and name are required",
		})
	}

	// ✅ Check category exists
	var category models.Category
	if err := middleware.DBConn.First(&category, input.CategoryID).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Category not found",
		})
	}

	// ✅ Check duplicate subcategory under same category
	var existing models.SubCategory
	if err := middleware.DBConn.
		Where("name = ? AND category_id = ?", input.Name, input.CategoryID).
		First(&existing).Error; err == nil {
		return c.Status(409).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Subcategory already exists under this category",
		})
	}

	input.CreatedAt = time.Now()

	if err := middleware.DBConn.Create(&input).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create subcategory",
		})
	}

	return c.Status(201).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Subcategory created successfully",
		Data:    input,
	})
}

func UpdateSubCategoryDescription(c *fiber.Ctx) error {
	id := c.Params("id")

	var input struct {
		Description string `json:"description"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var subcategory models.SubCategory

	if err := middleware.DBConn.First(&subcategory, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Subcategory not found",
		})
	}

	subcategory.Description = input.Description

	if err := middleware.DBConn.Save(&subcategory).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update description",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Description updated successfully",
		Data:    subcategory,
	})
}

func GetCategories(c *fiber.Ctx) error {
	var categories []models.Category

	if err := middleware.DBConn.
		Preload("SubCategories").
		Find(&categories).Error; err != nil {

		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch categories",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Categories fetched successfully",
		Data:    categories,
	})
}

func GetSubCategoriesByCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")

	if categoryID == "" {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Category ID is required",
		})
	}

	// ✅ Check if category exists
	var category models.Category
	if err := middleware.DBConn.First(&category, categoryID).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Category not found",
		})
	}

	var subcategories []models.SubCategory

	if err := middleware.DBConn.
		Where("category_id = ?", categoryID).
		Find(&subcategories).Error; err != nil {

		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch subcategories",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Subcategories fetched successfully",
		Data:    subcategories,
	})
}

func UpdateCategoryName(c *fiber.Ctx) error {
	id := c.Params("id")

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Name is required",
		})
	}

	var category models.Category

	// ✅ Check if exists
	if err := middleware.DBConn.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Category not found",
		})
	}

	// ✅ Optional: check duplicate
	var existing models.Category
	if err := middleware.DBConn.
		Where("name = ? AND category_id != ?", input.Name, id).
		First(&existing).Error; err == nil {

		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Category name already exists",
		})
	}

	category.Name = input.Name

	if err := middleware.DBConn.Save(&category).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update category",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Category updated successfully",
		Data:    category,
	})
}

func UpdateSubCategoryName(c *fiber.Ctx) error {
	id := c.Params("id")

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Name is required",
		})
	}

	var sub models.SubCategory

	// ✅ Check if exists
	if err := middleware.DBConn.First(&sub, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Subcategory not found",
		})
	}

	// ✅ Prevent duplicate in same category
	var existing models.SubCategory
	if err := middleware.DBConn.
		Where("name = ? AND category_id = ? AND sub_category_id != ?",
			input.Name, sub.CategoryID, id).
		First(&existing).Error; err == nil {

		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Subcategory name already exists in this category",
		})
	}

	sub.Name = input.Name

	if err := middleware.DBConn.Save(&sub).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update subcategory",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Subcategory name updated successfully",
		Data:    sub,
	})
}

func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category models.Category

	// ✅ Check if exists
	if err := middleware.DBConn.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Category not found",
		})
	}

	// ✅ Delete (cascade will handle subcategories)
	if err := middleware.DBConn.Delete(&category).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete category",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Category and its subcategories deleted successfully",
	})
}

func DeleteSubCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var subcategory models.SubCategory

	// ✅ Check if exists
	if err := middleware.DBConn.First(&subcategory, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Subcategory not found",
		})
	}

	// ✅ Delete
	if err := middleware.DBConn.Delete(&subcategory).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete subcategory",
		})
	}

	return c.Status(200).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Subcategory deleted successfully",
	})
}


// ============================================
// INSTITUTIONS!!
// ============================================

func CreateInstitution(c *fiber.Ctx) error {
	var input models.Institution

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if err := middleware.DBConn.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create institution",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Institution created successfully",
		Data:    input,
	})
}

func UpdateInstitutionStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	type Request struct {
		Status string `json:"status"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var institution models.Institution
	if err := middleware.DBConn.First(&institution, "institution_id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Institution not found",
		})
	}

	institution.Status = req.Status

	if err := middleware.DBConn.Save(&institution).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Status updated successfully",
		Data:    institution,
	})
}

func GetInstitutions(c *fiber.Ctx) error {
	var institutions []models.Institution

	if err := middleware.DBConn.Preload("Positions").Find(&institutions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch institutions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Institutions fetched successfully",
		Data:    institutions,
	})
}

func CreatePosition(c *fiber.Ctx) error {
	var input models.Position

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// check if institution exists
	var inst models.Institution
	if err := middleware.DBConn.First(&inst, "institution_id = ?", input.InstitutionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Institution not found",
		})
	}

	if err := middleware.DBConn.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create position",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Position created successfully",
		Data:    input,
	})
}

func UpdatePositionName(c *fiber.Ctx) error {
	id := c.Params("id")

	type Request struct {
		Name string `json:"name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var position models.Position
	if err := middleware.DBConn.First(&position, "position_id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Position not found",
		})
	}

	position.Name = req.Name

	if err := middleware.DBConn.Save(&position).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update position",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Position updated successfully",
		Data:    position,
	})
}

func GetPositions(c *fiber.Ctx) error {
	var positions []models.Position

	if err := middleware.DBConn.Preload("Institution").Find(&positions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch positions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Positions fetched successfully",
		Data:    positions,
	})
}

func UpdatePositionStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	type Request struct {
		Status string `json:"status"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var position models.Position
	if err := middleware.DBConn.First(&position, "position_id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Position not found",
		})
	}

	position.Status = req.Status

	if err := middleware.DBConn.Save(&position).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Status updated successfully",
		Data:    position,
	})
}