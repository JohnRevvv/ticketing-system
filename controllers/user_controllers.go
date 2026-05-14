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
		Username    string `json:"username"`
		Password    string `json:"password"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Position    string `json:"position"`
		Email       string `json:"email"`
		Institution string `json:"institution"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if body.Username == "" ||
		body.Password == "" ||
		body.FirstName == "" ||
		body.LastName == "" ||
		body.Position == "" ||
		body.Email == "" ||
		body.Institution == "" {

		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "All fields are required",
		})
	}

	// Check if username already exists
	var existingUser models.UserAccount

	if err := middleware.DBConn.
		Where("username = ?", body.Username).
		First(&existingUser).Error; err == nil {

		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Username already exists",
		})
	}

	// Check if email already exists
	if err := middleware.DBConn.
		Where("email = ?", body.Email).
		First(&existingUser).Error; err == nil {

		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Email already exists",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(body.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash password",
		})
	}

	user := models.UserAccount{
		Username:    body.Username,
		Password:    string(hashedPassword),
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Email:       body.Email,
		Position:    body.Position,
		Institution: body.Institution,
		Status:      "inactive",
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
	ticketID := c.Params("id")

	// Get user ID from JWT (endorser)
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// Endorser info
	var endorser models.UserAccount
	if err := middleware.DBConn.First(&endorser, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	if endorser.Role != "endorser" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Access denied",
		})
	}

	// Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		First(&ticket).Error; err != nil {

		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	if ticket.Status != "for endorsement" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid status",
		})
	}

	// Update ticket
	ticket.Status = "for approval"
	if err := middleware.DBConn.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Update failed",
		})
	}

	// ============================================
	// GET SUBMITTER (FULL NAME ONLY)
	// ============================================
	var submitter models.UserAccount
	middleware.DBConn.
		Where("username = ?", ticket.Username).
		First(&submitter)

	submitterFullName := strings.TrimSpace(
		submitter.FirstName + " " + submitter.LastName,
	)

	endorserFullName := strings.TrimSpace(
		endorser.FirstName + " " + endorser.LastName,
	)

	// ============================================
	// APPROVERS NOTIFICATION (FULL NAME FIXED)
	// ============================================
	var approvers []models.UserAccount
	if err := middleware.DBConn.
		Where("role = ?", "approver").
		Find(&approvers).Error; err == nil {

		for _, approver := range approvers {

			if approver.Email == "" {
				continue
			}

			approverFullName := strings.TrimSpace(
				approver.FirstName + " " + approver.LastName,
			)

			go services.SendApproverNotification(
				ticket,
				approver.Email,
				submitterFullName,
				endorserFullName,
				approverFullName,
			)
		}
	}

	// ============================================
	// SUBMITTER NOTIFICATION
	// ============================================
	if submitter.Email != "" {

		go services.SendEndorsedNotification(
			ticket,
			submitterFullName,
			submitter.Email,
			endorserFullName,
		)
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket endorsed successfully",
		Data:    ticket,
	})
}

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
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		First(&ticket).Error; err != nil {

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

	// =========================================
	// FETCH ENDORSER
	// =========================================
	var endorser models.UserAccount
	middleware.DBConn.
		Where("username = ?", ticket.Endorser).
		First(&endorser)

	endorserName := strings.TrimSpace(endorser.FirstName + " " + endorser.LastName)

	// =========================================
	// FETCH SUBMITTER
	// =========================================
	var submitter models.UserAccount
	middleware.DBConn.
		Where("username = ?", ticket.Username).
		First(&submitter)

	submitterName := strings.TrimSpace(submitter.FirstName + " " + submitter.LastName)

	// =========================================
	// FETCH APPROVER (current user)
	// =========================================
	approverName := strings.TrimSpace(user.FirstName + " " + user.LastName)

	// =========================================
	// ✅ SEND EMAIL TO RESOLVERS
	// =========================================
	var resolvers []models.UserAccount

	if err := middleware.DBConn.
		Where("role = ?", "resolver").
		Find(&resolvers).Error; err != nil {

		log.Println("Failed to fetch resolvers:", err)
	}

	for _, r := range resolvers {

		r := r // safe copy

		if r.Email == "" {
			continue
		}

		resolverName := strings.TrimSpace(r.FirstName + " " + r.LastName)

		go func(resolver models.UserAccount, resolverName string) {

			err := services.SendResolverNotification(
				ticket,
				resolverName, // ✅ FULL NAME
				resolver.Email,
				submitterName,
				approverName,
				endorserName,
			)

			if err != nil {
				log.Println("Failed resolver email:", resolver.Email, err)
			}

		}(r, resolverName)
	}

	// =========================================
	// ✅ NOTIFY SUBMITTER
	// =========================================
	if submitter.Email != "" {

		go func() {

			err := services.SendApprovedNotification(
				ticket,
				submitterName,
				submitter.Email,
				user.FirstName+" "+user.LastName,
			)

			if err != nil {
				log.Println("Failed to send submitter approval email:", err)
			}

		}()
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket approved successfully and resolvers notified",
		Data:    ticket,
	})
}

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

	// ✅ Submitter can ONLY cancel while ticket is still for endorsement
	if ticket.Username == user.Username &&
		ticket.Status == "for endorsement" {
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

	// 👤 Get resolver info
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
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		First(&ticket).Error; err != nil {

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

	// ✅ Calculate net working duration
	now := time.Now()

	var wallDuration time.Duration

	if ticket.StartedAt != nil {
		wallDuration = now.Sub(*ticket.StartedAt)
	} else {
		wallDuration = now.Sub(ticket.CreatedAt)
	}

	// Subtract hold duration
	holdSeconds := ticket.TotalHoldSeconds

	if ticket.OnHold && ticket.HoldStartedAt != nil {
		holdSeconds += now.Sub(*ticket.HoldStartedAt).Seconds()
	}

	netDuration := wallDuration - time.Duration(holdSeconds)*time.Second

	if netDuration < 0 {
		netDuration = 0
	}

	resolutionStr := humanDuration(netDuration)

	// ============================================
	// UPDATE TICKET
	// ============================================
	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"status":             "resolved",
		"resolved_at":        now,
		"resolution_minutes": netDuration.Minutes(),
		"resolution_time":    resolutionStr,
	}).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to resolve ticket",
		})
	}

	// ============================================
	// FETCH SUBMITTER
	// ============================================
	var submitter models.UserAccount

	if err := middleware.DBConn.
		Where("username = ?", ticket.Username).
		First(&submitter).Error; err != nil {

		log.Println("❌ Could not find submitter:", err)

	} else if submitter.Email != "" {

		// ✅ Full names
		submitterName := strings.TrimSpace(
			submitter.FirstName + " " + submitter.LastName,
		)

		resolverName := strings.TrimSpace(
			user.FirstName + " " + user.LastName,
		)

		log.Println("📧 Sending resolved email to:", submitter.Email)

		go func(
			t models.CreateTicket,
			submitterName string,
			email string,
			resolverName string,
		) {

			if err := services.SendResolvedNotification(
				t,
				submitterName,
				email,
				resolverName,
			); err != nil {

				log.Println("❌ Failed to send resolved email:", err)

			} else {

				log.Println("✅ Email sent to:", email)
			}

		}(
			ticket,
			submitterName,
			submitter.Email,
			resolverName,
		)
	}

	// ============================================
	// RESPONSE
	// ============================================
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
// RESERVED FOR PHASE 2!!
// ============================================

func CreateInstitution(c *fiber.Ctx) error {

	// 🔐 Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// 👤 Get user
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "User not found",
		})
	}

	// 🚫 Only super admin allowed
	if user.Role != "super_admin" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only super admin can create institution",
		})
	}

	// 📥 Request body (no DTO)
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// 🔍 Validate
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Institution name is required",
		})
	}

	// 🚨 Check duplicate name
	var existing models.Institution
	if err := middleware.DBConn.
		Where("name = ?", req.Name).
		First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Institution already exists",
		})
	}

	// 🏗 Create institution
	institution := models.Institution{
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedBy:   user.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := middleware.DBConn.Create(&institution).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create institution",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Institution created successfully",
		Data:    institution,
	})
}

func GetInstitutions(c *fiber.Ctx) error {
	var institutions []models.Institution

	if err := middleware.DBConn.Find(&institutions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch institutions",
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Success",
		Data:    institutions,
	})
}

func AssignAdminRole(c *fiber.Ctx) error {

	// 🔐 Get requester
	requesterID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	var requester models.UserAccount
	if err := middleware.DBConn.First(&requester, requesterID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "User not found",
		})
	}

	// 🚫 Only super admin can assign roles
	if requester.Role != "super_admin" {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Only super admin can assign admin role",
		})
	}

	// 📥 Request body
	var req struct {
		UserID      uint   `json:"user_id"`
		Institution string `json:"institution"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	if req.UserID == 0 || req.Institution == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "user_id and institution are required",
		})
	}

	// 👤 Find target user
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, req.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
		})
	}

	// 🏢 Assign role + institution
	user.Role = "admin"
	user.Institution = req.Institution

	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to assign admin role",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User promoted to admin successfully",
		Data:    user,
	})
}

func GetAllUsers2(c *fiber.Ctx) error {

	// 🔐 Get requester from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	// 👤 Get user info
	var requester models.UserAccount
	if err := middleware.DBConn.First(&requester, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "User not found",
		})
	}

	var users []models.UserAccount

	query := middleware.DBConn.Model(&models.UserAccount{})

	// 🚫 If NOT super admin → restrict to same institution
	if requester.Role != "super_admin" {
		query = query.Where("institution = ?", requester.Institution)
	}

	// 📥 Execute query
	if err := query.Find(&users).Error; err != nil {
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