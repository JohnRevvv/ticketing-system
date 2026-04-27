package controllers

import (
	"fmt"
	"log"
	"strings"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"ticketing-be-dev/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Position string `json:"position"`
		Role     string `json:"role"`
		Email    string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Check if username already exists
	var existingUser models.UserAccount
	if err := middleware.DBConn.Where("username = ?", body.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Username already exists",
		})
	}

	// Check if email already exists
	if err := middleware.DBConn.Where("email = ?", body.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Email already exists",
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
		Username: body.Username,
		Password: string(hashedPassword),
		FullName: body.FullName,
		Email:    body.Email,
		Position: body.Position,
		Role:     body.Role,
		Status:   "active",
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

func Login(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"` // can be username OR email
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	var user models.UserAccount

	// 🔥 Allow login via username OR email
	if err := middleware.DBConn.
		Where("username = ? OR email = ?", body.Username, body.Username).
		First(&user).Error; err != nil {

		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username/email or password",
		})
	}

	// ✅ Check if user is active
	if user.Status != "active" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "User account is for approval. Wait for the admin to approve your account.",
		})
	}

	// 🔐 Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid username/email or password",
		})
	}

	// 🔑 Generate JWT
	token, err := middleware.GenerateJWT(user.UserID)
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
			"user_id":  user.UserID,
			"username": user.Username,
			"email":    user.Email, // optional but useful
			"token":    token,
		},
	})
}

func GetCurrentUser(c *fiber.Ctx) error {
	// Assume token is passed in Authorization header: "Bearer <token>"
	userID := c.Locals("userID") // This should be set by your JWT middleware

	var user models.UserAccount
	if err := middleware.DBConn.First(&user, "user_id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User fetched successfully",
		Data:    user,
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.UserAccount

	if err := middleware.DBConn.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch users",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Users fetched successfully",
		Data:    users,
	})
}

func EndorseTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id") // SR0000001

	// Get user ID from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// ✅ Check if user is endorser
	if user.Role != "endorser" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Access denied: Only endorsers can endorse tickets",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// ✅ Check current status (optional but recommended)
	if ticket.Status != "for endorsement" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket is not for endorsement",
		})
	}

	// ✅ Update status
	// ✅ Update status
	ticket.Status = "for approval"
	if err := middleware.DBConn.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update ticket",
		})
	}

	// Get approver email
	var approvers []models.UserAccount
	if err := middleware.DBConn.
		Where("role = ?", "approver").
		Find(&approvers).Error; err != nil {

		log.Println("Failed to fetch approvers:", err)
	} else {

		for _, approver := range approvers {
			if approver.Email != "" {

				// async email
				go services.SendApproverNotification(
					ticket,
					approver.Email,
					approver.FullName,
				)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket endorsed successfully",
		Data:    ticket,
	})
}

// ── REPLACE your ApproveTicket function with this ──────────────────────────

// ── Replace ONLY the ApproveTicket function in your user controller ──────────
// The rest of the file stays exactly the same.

func ApproveTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// Get user ID from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// Check role
	if user.Role != "approver" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Access denied: Only approvers can approve tickets",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// Cannot approve if still for endorsement
	if ticket.Status == "for endorsement" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket must be endorsed first before approval",
		})
	}

	// Prevent re-approval
	if ticket.Status != "for approval" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket is not for approval",
		})
	}

	// Save approver username + update status
	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"status":   "for assignment",
		"approver": user.Username,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update ticket",
		})
	}

	// ── Send email to every resolver ─────────────────────────────────────────
	// FIX: pass r.Username so each resolver gets a personalised greeting
	// instead of the empty ticket.Assignee field.
	var resolvers []models.UserAccount
	if err := middleware.DBConn.Where("role = ?", "resolver").Find(&resolvers).Error; err != nil {
		log.Println("Failed to fetch resolvers:", err)
	}

	for _, r := range resolvers {
		if r.Email != "" {
			resolverUsername := r.Username // capture loop variable for goroutine
			resolverEmail := r.Email
			go func() {
				if err := services.SendResolverNotification(ticket, resolverUsername, resolverEmail); err != nil {
					log.Println("Failed to send resolver email to", resolverEmail, ":", err)
				}
			}()
		}
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket approved successfully and resolvers notified",
		Data:    ticket,
	})
}

// ── REPLACE your CancelTicket function with this ────────────────────────────
// Now allows the ticket CREATOR (any role) to cancel their own ticket,
// in addition to the existing role-based rules.

func CancelTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// Already cancelled or resolved — nothing to do
	if ticket.Status == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket is already cancelled",
		})
	}
	if ticket.Status == "resolved" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Resolved tickets cannot be cancelled",
		})
	}

	// =========================
	// ✅ ROLE + OWNERSHIP CHECK
	// =========================
	canCancel := false

	switch user.Role {
	case "admin":
		// Admin can always cancel
		canCancel = true

	case "endorser":
		if ticket.Status == "for endorsement" {
			canCancel = true
		}

	case "approver":
		if ticket.Status == "for approval" {
			canCancel = true
		}

	case "resolver":
		// Resolver can cancel only tickets assigned to them
		if ticket.Assignee == user.Username {
			canCancel = true
		}
	}

	// ✅ FIXED: The ticket CREATOR can cancel regardless of their role,
	// as long as the ticket is not yet in progress or resolved.
	if ticket.Username == user.Username &&
		ticket.Status != "in progress" &&
		ticket.Status != "resolved" &&
		ticket.Status != "cancelled" {
		canCancel = true
	}

	if !canCancel {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "You are not allowed to cancel this ticket at this stage",
		})
	}

	// =========================
	// ✅ CANCEL TICKET
	// =========================
	now := time.Now()
	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"status":       "cancelled",
		"cancelled_by": user.Username,
		"cancelled_at": now,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to cancel ticket",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket cancelled successfully",
		Data:    ticket,
	})
}

func GrabTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// ✅ Role check
	if user.Role != "resolver" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only resolver can grab tickets",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// ❌ Must be approved first
	if ticket.Status != "for assignment" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket must be approved before grabbing",
		})
	}

	// ❌ Prevent double grab
	if ticket.Assignee != "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket already assigned",
		})
	}

	now := time.Now()

	// ✅ Assign + update status
	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"assignee":   user.Username,
		"status":     "in progress",
		"started_at": now,
	}).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to grab ticket",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket grabbed successfully",
		Data:    ticket,
	})
}

func UnGrabTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// Role check
	if user.Role != "resolver" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only resolver can ungrab tickets",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// Must be assigned to someone
	if ticket.Assignee == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket is not currently assigned",
		})
	}

	// Only allow the same resolver to ungrab it
	if ticket.Assignee != user.Username {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "You can only ungrab tickets assigned to you",
		})
	}

	// ✅ Select forces GORM to update empty string and nil fields
	if err := middleware.DBConn.Model(&ticket).
		Select("assignee", "status", "started_at").
		Updates(map[string]interface{}{
			"assignee":   "",
			"status":     "for assignment",
			"started_at": nil,
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to ungrab ticket",
		})
	}

	// Refresh ticket data
	middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket ungrabbed successfully",
		Data:    ticket,
	})
}

func ResolveTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// 🔐 Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// 👤 Get user info
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// 🚫 Role check
	if user.Role != "resolver" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only resolver can resolve tickets",
		})
	}

	// 🎫 Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// ⚠️ Must be in progress
	if ticket.Status != "in progress" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket is not in progress",
		})
	}

	// 🔒 Only assigned resolver can resolve
	if ticket.Assignee != user.Username {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "You are not assigned to this ticket",
		})
	}

	// ✅ Calculate net working duration (wall time − hold time)
	now := time.Now()
	var wallDuration time.Duration
	if ticket.StartedAt != nil {
		wallDuration = now.Sub(*ticket.StartedAt)
	} else {
		wallDuration = now.Sub(ticket.CreatedAt)
	}

	// Subtract accumulated hold seconds
	holdSeconds := ticket.TotalHoldSeconds
	if ticket.OnHold && ticket.HoldStartedAt != nil {
		holdSeconds += now.Sub(*ticket.HoldStartedAt).Seconds()
	}
	netDuration := wallDuration - time.Duration(holdSeconds)*time.Second
	if netDuration < 0 {
		netDuration = 0
	}

	resolutionStr := humanDuration(netDuration)

	// 💾 Update DB
	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"status":             "resolved",
		"resolved_at":        now,
		"resolution_minutes": netDuration.Minutes(), // clean net minutes, no hold
		"resolution_time":    resolutionStr,         // ← add this
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to resolve ticket",
		})
	}

	// 📧 SEND EMAIL
	var submitter models.UserAccount
	if err := middleware.DBConn.
		Where("username = ?", ticket.Username).
		First(&submitter).Error; err != nil {
		log.Println("❌ Could not find submitter:", err)
	} else if submitter.Email != "" {
		log.Println("📧 Sending resolved email to:", submitter.Email)
		go func(t models.CreateTicket, username, email string) {
			if err := services.SendResolverNotification(t, username, email); err != nil {
				log.Println("❌ Failed to send resolved email:", err)
			} else {
				log.Println("✅ Email sent to:", email)
			}
		}(ticket, submitter.Username, submitter.Email)
	}

	// ✅ RESPONSE
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket resolved successfully",
		Data: fiber.Map{
			"ticket":     ticket,
			"resolution": resolutionStr,
		},
	})
}

// ── humanDuration: formats duration as  1d 02h 30m 15s  ─────────────────────
func humanDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	days := total / 86400
	hours := (total % 86400) / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02dm %02ds", minutes, seconds)
}

func AddCategoryWithSubcategories(c *fiber.Ctx) error {
	// ── Get user from JWT ─────────────────────────────
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// ── Admin only ────────────────────────────────────
	if user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only admin can add category",
		})
	}

	// ── Request body ──────────────────────────────────
	var input struct {
		Name          string   `json:"name"`
		SubCategories []string `json:"subcategories"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Category name is required",
		})
	}

	// ── Start transaction ─────────────────────────────
	tx := middleware.DBConn.Begin()

	var category models.Category

	// ── CHECK IF CATEGORY EXISTS ──────────────────────
	err = tx.Where("LOWER(name) = LOWER(?)", input.Name).First(&category).Error

	if err != nil {
		// NOT FOUND → create new
		category = models.Category{
			Name:      input.Name,
			CreatedBy: user.Username,
		}

		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to create category",
			})
		}
	}

	// ── CREATE SUBCATEGORIES (NO DUPLICATES) ──────────
	for _, sub := range input.SubCategories {
		sub = strings.TrimSpace(sub)
		if sub == "" {
			continue
		}

		var existing models.SubCategory

		err := tx.Where("category_id = ? AND LOWER(name) = LOWER(?)",
			category.CategoryID, sub).First(&existing).Error

		if err == nil {
			// already exists → skip
			continue
		}

		subCategory := models.SubCategory{
			CategoryID: category.CategoryID,
			Name:       sub,
			CreatedBy:  user.Username,
		}

		if err := tx.Create(&subCategory).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to create subcategory",
			})
		}
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Category and subcategories saved successfully",
		Data:    category,
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

	return c.Status(200).JSON(categories)
}

func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category models.Category

	if err := middleware.DBConn.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Category not found",
		})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request",
		})
	}

	if input.Name == "" {
		return c.Status(400).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Name is required",
		})
	}

	category.Name = input.Name

	if err := middleware.DBConn.Save(&category).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update category",
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Category updated successfully",
		Data:    category,
	})
}

func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	tx := middleware.DBConn.Begin()

	// delete subcategories first
	if err := tx.Where("category_id = ?", id).
		Delete(&models.SubCategory{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete subcategories",
		})
	}

	// delete category
	if err := tx.Delete(&models.Category{}, id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete category",
		})
	}

	tx.Commit()

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Category deleted successfully",
	})
}

func DeleteSubCategories(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := middleware.DBConn.Delete(&models.SubCategory{}, id).Error; err != nil {
		return c.Status(500).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete subcategory",
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Subcategory deleted successfully",
	})
}
