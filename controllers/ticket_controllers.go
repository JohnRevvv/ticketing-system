package controllers

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	// Handle attachments
	form, err := c.MultipartForm()
	if err == nil && form.File != nil {
		files := form.File["attachments"]

		baseUploadPath := os.Getenv("UPLOAD_PATH")
		if baseUploadPath == "" {
			baseUploadPath = "upload/attachments"
		}

		if err := os.MkdirAll(baseUploadPath, os.ModePerm); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to create upload directory",
			})
		}

		for _, file := range files {
			if file.Size > 5*1024*1024 {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
					RetCode: "400",
					Message: "File too large (max 5MB)",
				})
			}

			cleanFileName := filepath.Base(file.Filename)
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
				FilePath:   filePath,
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

	// Get endorser email
	var endorser models.UserAccount
	if err := middleware.DBConn.
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

func GetAllTickets(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	// Fetch all tickets ordered by creation date
	if err := middleware.DBConn.Order("created_at desc").Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch tickets",
		})
	}

	// Prepare response with attachments
	var responseData []fiber.Map

	for _, ticket := range tickets {
		var attachments []models.TicketAttachment
		_ = middleware.DBConn.Where("ticket_id = ?", ticket.TicketID).Find(&attachments).Error

		responseData = append(responseData, fiber.Map{
			"ticket":      ticket,
			"attachments": attachments,
		})
	}

	// If no tickets found
	if len(tickets) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No tickets found",
			Data:    []string{},
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

	// ✅ If no tickets
	if len(tickets) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No tickets found",
			Data:    []TicketWithAttachments{},
		})
	}

	// ✅ Attach attachments per ticket
	var result []TicketWithAttachments

	for _, ticket := range tickets {
		var attachments []models.TicketAttachment

		middleware.DBConn.
			Where("ticket_id = ?", ticket.TicketID).
			Find(&attachments)

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

	// 🔹 Convert file paths to URL (for frontend)
	baseURL := "http://localhost:8080/uploads"

	for i := range attachments {
		cleanPath := strings.TrimPrefix(attachments[i].FilePath, "upload/")
		attachments[i].FilePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
	}

	// 🔹 Response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket fetched successfully",
		Data: fiber.Map{
			"ticket":      ticket,
			"attachments": attachments,
		},
	})
}

func ExportTicketsCSV(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	// Optional query params
	month := c.Query("month") // e.g., "4"
	year := c.Query("year")   // e.g., "2026"

	db := middleware.DBConn

	// Filter by month/year if provided
	if month != "" && year != "" {
		m, err1 := strconv.Atoi(month)
		y, err2 := strconv.Atoi(year)
		if err1 == nil && err2 == nil {
			start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
			end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
			db = db.Where("created_at BETWEEN ? AND ?", start, end)
		}
	}

	// Preload users involved (if you have separate user table)
	if err := db.Order("created_at desc").Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch tickets")
	}

	// Prepare CSV headers
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment;filename="tickets_report.csv"`)
	writer := csv.NewWriter(c.Response().BodyWriter())
	defer writer.Flush()

	headers := []string{
		"Ticket ID", "Creator", "Category", "Subject", "Institution",
		"Type", "Description", "Purpose", "Priority",
		"Assignee", "Endorser", "Approver", "Status",
		"Created At", "Updated At", "Cancelled By", "Cancelled At",
	}
	writer.Write(headers)

	// Write ticket rows
	for _, t := range tickets {
		row := []string{
			t.TicketID,
			t.Username, // Creator
			t.Category,
			t.Subject,
			t.Institution,
			t.Tickettype,
			t.Description,
			t.Priority,
			t.Assignee,
			t.Endorser,
			t.Approver,
			t.Status,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.UpdatedAt.Format("2006-01-02 15:04:05"),
			t.CancelledBy,
			"",
		}
		if t.CancelledAt != nil {
			row[17] = t.CancelledAt.Format("2006-01-02 15:04:05")
		}
		writer.Write(row)
	}

	return nil
}

func ViewAttachment(c *fiber.Ctx) error {
	// Get attachment ID or filename from params
	attachmentID := c.Params("id")

	var attachment models.TicketAttachment

	// Find attachment in DB
	if err := middleware.DBConn.
		Where("id = ?", attachmentID).
		First(&attachment).Error; err != nil {

		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Attachment not found",
		})
	}

	// Check if file exists
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "File not found on server",
		})
	}

	// Option 1: Display in browser (for images, pdf, etc.)
	return c.SendFile(attachment.FilePath)

	// Option 2 (force download instead):
	/*
		return c.Download(attachment.FilePath, attachment.FileName)
	*/
}

// ============================================
// TICKET REMARKS FUNCTION!!
// ============================================
func CreateTicketRemark(c *fiber.Ctx) error {
	var input struct {
		TicketID string `json:"ticket_id"`
		UserID   string `json:"user_id"`
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
	if input.TicketID == "" || input.UserID == "" || input.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "ticket_id, user_id, and message are required",
		})
	}

	remark := models.TicketRemark{
		RemarkID:  uuid.New().String(),
		TicketID:  input.TicketID,
		UserID:    input.UserID,
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
