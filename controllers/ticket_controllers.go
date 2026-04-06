package controllers

import (
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
)

var ticketIDMutex = &sync.Mutex{}

// generateTicketID returns the next ticket code like SR0000001
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
	fmt.Sscanf(lastTicket.TicketID, "SR%07d", &num)
	num++
	return fmt.Sprintf("SR%07d", num)
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
		Purpose:     c.FormValue("purpose"),
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

func GetTicketsReport(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	// Optional query params
	month := c.Query("month")   // "4"
	year := c.Query("year")     // "2026"

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

	if err := db.Order("created_at desc").Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch tickets",
		})
	}

	// Build table report (same as previous example)
	report := make([]map[string]interface{}, len(tickets))
	for i, t := range tickets {
		report[i] = map[string]interface{}{
			"Ticket ID":    t.TicketID,
			"Username":     t.Username,
			"Category":     t.Category,
			"Subject":      t.Subject,
			"Institution":  t.Institution,
			"Type":         t.Tickettype,
			"Description":  t.Description,
			"Purpose":      t.Purpose,
			"Priority":     t.Priority,
			"Assignee":     t.Assignee,
			"Endorser":     t.Endorser,
			"Approver":     t.Approver,
			"Remarks":      t.Remarks,
			"Status":       t.Status,
			"Created At":   t.CreatedAt.Format("2006-01-02 15:04:05"),
			"Updated At":   t.UpdatedAt.Format("2006-01-02 15:04:05"),
			"Cancelled By": t.CancelledBy,
			"Cancelled At": "",
		}
		if t.CancelledAt != nil {
			report[i]["Cancelled At"] = t.CancelledAt.Format("2006-01-02 15:04:05")
		}
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Ticket report generated successfully",
		Data:    report,
	})
}
