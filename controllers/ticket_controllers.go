package controllers

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"ticketing-be-dev/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var ticketIDMutex = &sync.Mutex{}

// generateTicketID returns the next ticket code like SR00001
func generateTicketID() string {
	ticketIDMutex.Lock()
	defer ticketIDMutex.Unlock()

	var lastTicket models.CreateTicket
	if err := middleware.DBConn.Order("created_at desc").First(&lastTicket).Error; err != nil {
		// No tickets yet
		return "SR000001"
	}

	// Extract numeric part
	var num int
	fmt.Sscanf(lastTicket.TicketID, "SR%06d", &num)
	num++
	return fmt.Sprintf("SR%06d", num)
}

func CreateTicket(c *fiber.Ctx) error {
	// Start DB transaction
	tx := middleware.DBConn.Begin()

	// Parse ticket fields
	ticket := models.CreateTicket{
		TicketID:    generateTicketID(),
		Subject:     c.FormValue("subject"),
		Category:    c.FormValue("category"),
		Institution: c.FormValue("institution"),
		Tickettype:  c.FormValue("tickettype"),
		Description: c.FormValue("description"),
		Assignee:    c.FormValue("assignee"),
		Priority:    c.FormValue("priority"),
		Endorser:    c.FormValue("endorser"),
		Approver:    c.FormValue("approver"),
		Status:      "for endorsement",
	}

	// Get user info from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized: User ID not found",
		})
	}

	var user models.UserAccount
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user info",
		})
	}

	ticket.Username = user.Username

	// Save ticket
	if err := tx.Create(&ticket).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create ticket",
		})
	}

	// Handle attachments (S3 only)
	form, err := c.MultipartForm()
	if err == nil && form.File != nil {
		files := form.File["attachments"]

		for _, file := range files {

			// ✅ File size validation (5MB)
			if file.Size > 5*1024*1024 {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
					RetCode: "400",
					Message: "File too large (max 5MB)",
				})
			}

			// ✅ File type validation
			contentType := file.Header.Get("Content-Type")
			allowedTypes := map[string]bool{
				"image/jpeg":      true,
				"image/png":       true,
				"application/pdf": true,
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true, // xlsx
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // docx
			}

			if !allowedTypes[contentType] {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
					RetCode: "400",
					Message: "Invalid file type",
				})
			}

			// ✅ Upload to S3
			fileName, fileURL, err := services.UploadToS3(file, ticket.TicketID)
			if err != nil {
				log.Println("❌ S3 UPLOAD FAILED")
				log.Println("TicketID:", ticket.TicketID)
				log.Println("Filename:", file.Filename)
				log.Println("Error:", err)

				tx.Rollback()

				return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
					RetCode: "500",
					Message: "S3 Upload failed",
					Data: fiber.Map{
						"error":     err.Error(),
						"ticket_id": ticket.TicketID,
						"file_name": file.Filename,
					},
				})
			}

			// ✅ Save metadata
			attachment := models.TicketAttachment{
				TicketID:   ticket.TicketID,
				FileName:   fileName,
				FileURL:    fileURL, // consider renaming to FileURL
				UploadedBy: user.Username,
			}

			if err := tx.Create(&attachment).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
					RetCode: "500",
					Message: "Failed to save attachment metadata",
				})
			}
		}
	}

	// Get endorser email (using SAME transaction)
	var endorser models.UserAccount
	if err := tx.
		Where("username = ?", ticket.Endorser).
		First(&endorser).Error; err != nil {

		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Endorser not found",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to finalize transaction",
		})
	}

	// Send email asynchronously
	go func() {
		if err := services.SendEndorserNotification(ticket, endorser.Email); err != nil {
			log.Println("Error sending endorser email:", err)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Ticket created successfully",
		Data: fiber.Map{
			"ticket_code": ticket.TicketID,
			"ticket":      ticket,
		},
	})
}

func UpdateTicket(c *fiber.Ctx) error {
	tx := middleware.DBConn.Begin()

	// Get ticket ID from params
	ticketID := c.Params("id")

	// Get user from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized",
		})
	}

	var user models.UserAccount
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user",
		})
	}

	// Get existing ticket
	var ticket models.CreateTicket
	if err := tx.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// 🔒 Check ownership
	if ticket.Username != user.Username {
		tx.Rollback()
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "You are not allowed to edit this ticket",
		})
	}

	// 🔒 Check status
	if ticket.Status != "for endorsement" {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Ticket can no longer be edited",
		})
	}

	// Update fields
	ticket.Subject = c.FormValue("subject")
	ticket.Category = c.FormValue("category")
	ticket.Institution = c.FormValue("institution")
	ticket.Tickettype = c.FormValue("tickettype")
	ticket.Description = c.FormValue("description")
	ticket.Assignee = c.FormValue("assignee")
	ticket.Priority = c.FormValue("priority")
	ticket.Endorser = c.FormValue("endorser")
	ticket.Approver = c.FormValue("approver")

	// Save updates
	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update ticket",
		})
	}

	// OPTIONAL: Handle new attachments (append only)
	form, err := c.MultipartForm()
	if err == nil && form.File != nil {
		files := form.File["attachments"]

		baseUploadPath := os.Getenv("UPLOAD_PATH")
		if baseUploadPath == "" {
			baseUploadPath = "uploads/attachments"
		}

		os.MkdirAll(baseUploadPath, os.ModePerm)

		for _, file := range files {
			if file.Size > 5*1024*1024 {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
					RetCode: "400",
					Message: "File too large (max 5MB)",
				})
			}

			cleanFileName := sanitizeFileName(file.Filename)
			savedFileName := fmt.Sprintf("%s_%d_%s", ticket.TicketID, time.Now().UnixNano(), cleanFileName)
			filePath := fmt.Sprintf("%s/%s", baseUploadPath, savedFileName)

			if err := c.SaveFile(file, filePath); err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
					RetCode: "500",
					Message: "Failed to save attachment",
				})
			}

			attachment := models.TicketAttachment{
				TicketID:   ticket.TicketID,
				FileName:   cleanFileName,
				FileURL:    filePath,
				UploadedBy: user.Username,
			}

			if err := tx.Create(&attachment).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
					RetCode: "500",
					Message: "Failed to save attachment metadata",
				})
			}
		}
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to commit update",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket updated successfully",
		Data:    ticket,
	})
}

func GetTicketByID(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	// 🔹 Get ticket
	var ticket models.CreateTicket
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		First(&ticket).Error; err != nil {

		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Ticket not found",
		})
	}

	// 🔹 Get attachments
	var attachments []models.TicketAttachment
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		Find(&attachments).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch attachments",
		})
	}

	// 🔹 NO MORE LOCAL PATH FIXING
	// Just return the S3 URL as-is

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket fetched successfully",
		Data: fiber.Map{
			"ticket":      ticket,
			"attachments": attachments,
		},
	})
}

func ViewAttachment(c *fiber.Ctx) error {
	attachmentID := c.Params("id")

	var attachment models.TicketAttachment

	if err := middleware.DBConn.
		Where("id = ?", attachmentID).
		First(&attachment).Error; err != nil {

		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Attachment not found",
		})
	}

	// 🔥 Redirect to S3 URL directly
	return c.Redirect(attachment.FileURL, 302)
}

func GetAllTickets(c *fiber.Ctx) error {
	var tickets []models.CreateTicket
	if err := middleware.DBConn.Order("created_at desc").Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch tickets",
		})
	}

	if len(tickets) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No tickets found",
			Data:    []string{},
		})
	}

	var responseData []fiber.Map
	for _, ticket := range tickets {
		var attachments []models.TicketAttachment
		_ = middleware.DBConn.Where("ticket_id = ?", ticket.TicketID).Find(&attachments).Error

		responseData = append(responseData, fiber.Map{
			"ticket_id":          ticket.TicketID,
			"username":           ticket.Username,
			"category":           ticket.Category,
			"subject":            ticket.Subject,
			"institution":        ticket.Institution,
			"tickettype":         ticket.Tickettype,
			"description":        ticket.Description,
			"priority":           ticket.Priority,
			"assignee":           ticket.Assignee,
			"endorser":           ticket.Endorser,
			"approver":           ticket.Approver,
			"status":             ticket.Status,
			"created_at":         ticket.CreatedAt,
			"updated_at":         ticket.UpdatedAt,
			"cancelled_by":       ticket.CancelledBy,
			"cancelled_at":       ticket.CancelledAt,
			"started_at":         ticket.StartedAt,
			"resolved_at":        ticket.ResolvedAt,
			"resolution_minutes": ticket.ResolutionMinutes,
			"resolution_time":    ticket.ResolutionTime,
			"onhold":             ticket.OnHold,
			"hold_started_at":    ticket.HoldStartedAt,
			"total_hold_seconds": ticket.TotalHoldSeconds,
			"attachments":        attachments,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Tickets fetched successfully",
		Data:    responseData,
	})
}

type TicketWithAttachments struct {
	models.CreateTicket
	Attachments []models.TicketAttachment `json:"attachments"`
}

// GetUserTickets returns tickets for the logged-in user
func GetUserTickets(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized: User ID not found",
		})
	}

	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user info",
		})
	}

	if err := middleware.DBConn.
		Where("username = ?", user.Username).
		Order("created_at desc").
		Find(&tickets).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch tickets",
		})
	}

	if len(tickets) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No tickets found",
			Data:    []TicketWithAttachments{},
		})
	}

	// 🔥 Base URL (change this!)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // fallback
	}

	var result []TicketWithAttachments

	for _, ticket := range tickets {
		var attachments []models.TicketAttachment

		middleware.DBConn.
			Where("ticket_id = ?", ticket.TicketID).
			Find(&attachments)

		// ✅ Fix paths here
		for i := range attachments {
			attachments[i].FileURL = fmt.Sprintf(
				"%s/uploads/%s",
				baseURL,
				attachments[i].FileURL,
			)
		}

		result = append(result, TicketWithAttachments{
			CreateTicket: ticket,
			Attachments:  attachments,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Tickets fetched successfully",
		Data:    result,
	})
}

func ExportTicketsCSV(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	// Optional query params
	month := c.Query("month")
	year := c.Query("year")

	db := middleware.DBConn

	if month != "" && year != "" {
		m, err1 := strconv.Atoi(month)
		y, err2 := strconv.Atoi(year)
		if err1 == nil && err2 == nil {
			start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
			end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
			db = db.Where("created_at BETWEEN ? AND ?", start, end)
		}
	}

	if err := db.Order("created_at desc").Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch tickets")
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment;filename="tickets_report.csv"`)
	writer := csv.NewWriter(c.Response().BodyWriter())
	defer writer.Flush()

	// 17 headers → indices 0–16
	headers := []string{
		"Ticket ID",    // 0
		"Creator",      // 1
		"Category",     // 2
		"Subject",      // 3
		"Institution",  // 4
		"Type",         // 5
		"Description",  // 6
		"Priority",     // 7
		"Assignee",     // 8
		"Endorser",     // 9
		"Approver",     // 10
		"Status",       // 11
		"Created At",   // 12
		"Updated At",   // 13
		"Cancelled By", // 14
		"Cancelled At", // 15
	}
	writer.Write(headers)

	for _, t := range tickets {
		// 16 values → indices 0–15, matching headers above
		cancelledAt := ""
		if t.CancelledAt != nil {
			cancelledAt = t.CancelledAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			t.TicketID,    // 0
			t.Username,    // 1
			t.Category,    // 2
			t.Subject,     // 3
			t.Institution, // 4
			t.Tickettype,  // 5
			t.Description, // 6
			t.Priority,    // 7
			t.Assignee,    // 8
			t.Endorser,    // 9
			t.Approver,    // 10
			t.Status,      // 11
			t.CreatedAt.Format("2006-01-02 15:04:05"), // 12
			t.UpdatedAt.Format("2006-01-02 15:04:05"), // 13
			t.CancelledBy, // 14
			cancelledAt,   // 15
		}
		writer.Write(row)
	}

	return nil
}

func sanitizeFileName(name string) string {
	// Remove weird unicode spaces and normalize
	name = strings.ReplaceAll(name, " ", " ")      // narrow no-break space
	name = strings.ReplaceAll(name, "\u00A0", " ") // non-breaking space

	// Replace spaces with underscore (optional but safer)
	name = strings.ReplaceAll(name, " ", "_")

	// Remove any problematic characters
	name = strings.Map(func(r rune) rune {
		if r > 127 {
			return -1 // remove non-ASCII
		}
		return r
	}, name)

	return name
}

// ============================================
// TICKET REMARKS FUNCTION!!
// ============================================
func CreateTicketRemark(c *fiber.Ctx) error {
	var input struct {
		TicketID string `json:"ticket_id"`
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Message  string `json:"message"`
	}

	// Parse body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
		})
	}

	// Validate
	if input.TicketID == "" || input.UserID == "" || input.Username == "" || input.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "ticket_id, user_id, username, and message are required",
		})
	}

	remark := models.TicketRemark{
		RemarkID:  uuid.New().String(),
		TicketID:  input.TicketID,
		UserID:    input.UserID,
		Username:  input.Username,
		Message:   input.Message,
		CreatedAt: time.Now(),
	}

	// Save to DB
	if err := middleware.DBConn.Create(&remark).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create remark",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Remark added successfully",
		Data:    remark,
	})
}

func GetRemarksByTicket(c *fiber.Ctx) error {
	ticketID := c.Params("ticket_id")

	var remarks []models.TicketRemark

	// Fetch remarks by ticket ordered by time (chat flow)
	if err := middleware.DBConn.
		Where("ticket_id = ?", ticketID).
		Order("created_at asc").
		Find(&remarks).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch remarks",
		})
	}

	// Prepare response (optional join later if needed)
	var responseData []fiber.Map

	for _, remark := range remarks {
		responseData = append(responseData, fiber.Map{
			"remark": remark,
		})
	}

	// If no remarks found
	if len(remarks) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No remarks found",
			Data:    []string{},
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Remarks fetched successfully",
		Data:    responseData,
	})
}

func HoldTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"message": "Ticket not found",
		})
	}

	// prevent invalid states
	if ticket.Status == "resolved" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Cannot hold a resolved ticket",
		})
	}

	if ticket.OnHold {
		return c.Status(400).JSON(fiber.Map{
			"message": "Ticket already on hold",
		})
	}

	now := time.Now()

	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"on_hold":         true,
		"hold_started_at": now,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Failed to put ticket on hold",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Ticket placed on hold",
	})
}

func ResumeTicket(c *fiber.Ctx) error {
	ticketID := c.Params("id")

	var ticket models.CreateTicket
	if err := middleware.DBConn.Where("ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"message": "Ticket not found",
		})
	}

	if !ticket.OnHold {
		return c.Status(400).JSON(fiber.Map{
			"message": "Ticket is not on hold",
		})
	}

	now := time.Now()

	var holdDuration float64 = 0

	if ticket.HoldStartedAt != nil {
		holdDuration = now.Sub(*ticket.HoldStartedAt).Seconds()
	}

	totalHold := float64(ticket.TotalHoldSeconds) + holdDuration

	if err := middleware.DBConn.Model(&ticket).Updates(map[string]interface{}{
		"on_hold":            false,
		"hold_started_at":    nil,
		"total_hold_seconds": totalHold,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Failed to resume ticket",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Ticket resumed",
	})
}
