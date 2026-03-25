package controllers

import (
	"fmt"
	"os"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/models"
	"ticketing-be-dev/models/response"
	"time"

	"github.com/gofiber/fiber/v2"
)

// generateTicketID returns the next ticket code like SR0000001
func generateTicketID() string {
	var lastTicket models.CreateTicket
	if err := middleware.DBConn.Order("ticket_id desc").First(&lastTicket).Error; err != nil {
		// No tickets yet
		return "SR0000001"
	}

	// Extract numeric part
	var num int
	fmt.Sscanf(lastTicket.TicketID, "SR%07d", &num)
	num++
	return fmt.Sprintf("SR%07d", num)
}

// CreateTicket handles new ticket creation
// func CreateTicket(c *fiber.Ctx) error {
// 	// Parse request body
// 	var body struct {
// 		Subject     string `json:"subject"`
// 		Title       string `json:"title"`
// 		Description string `json:"description"`
// 		Purpose     string `json:"purpose"`
// 		Assignee    string `json:"assignee"`
// 		Endorser    string `json:"endorser"`
// 		Approver    string `json:"approver"`
// 	}

// 	if err := c.BodyParser(&body); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
// 			RetCode: "400",
// 			Message: "Invalid request body",
// 		})
// 	}

// 	// Get user ID from JWT middleware
// 	userID, err := middleware.GetUserIDFromJWT(c)
// 	if err != nil {
// 		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
// 			RetCode: "401",
// 			Message: "Unauthorized: User ID not found in token",
// 		})
// 	}

// 	// Get user info
// 	var user models.UserAccount
// 	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to fetch user info",
// 		})
// 	}

// 	// Generate ticket ID
// 	ticketID := generateTicketID()

// 	// Create new ticket
// 	ticket := models.CreateTicket{
// 		TicketID:    ticketID,
// 		Username:    user.Username, // store username for reference
// 		Subject:     body.Subject,
// 		Title:       body.Title,
// 		Description: body.Description,
// 		Purpose:     body.Purpose,
// 		Assignee:    body.Assignee,
// 		Endorser:    body.Endorser,
// 		Approver:    body.Approver,
// 		Status:      "for endorsement",
// 	}

// 	// Save to DB
// 	if err := middleware.DBConn.Create(&ticket).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to create ticket",
// 		})
// 	}

// 	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
// 		RetCode: "201",
// 		Message: "Ticket created successfully",
// 		Data: fiber.Map{
// 			"ticket_code": ticket.TicketID,
// 			"ticket":      ticket,
// 		},
// 	})
// }

func CreateTicket(c *fiber.Ctx) error {
    // Parse ticket fields from form
    ticket := models.CreateTicket{
        TicketID:    generateTicketID(),
        Subject:     c.FormValue("subject"),
        Title:       c.FormValue("title"),
        Description: c.FormValue("description"),
        Purpose:     c.FormValue("purpose"),
        Assignee:    c.FormValue("assignee"),
        Endorser:    c.FormValue("endorser"),
        Approver:    c.FormValue("approver"),
        Status:      "for endorsement",
    }

    // Get user info from JWT
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
    ticket.Username = user.Username

    // Save ticket to DB
    if err := middleware.DBConn.Create(&ticket).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to create ticket",
        })
    }

    // Handle attachments
    form, err := c.MultipartForm()
    if err == nil && form.File != nil {
        files := form.File["attachments"]
        for _, file := range files {
            // Create folder for this ticket if not exists
            uploadDir := fmt.Sprintf("upload/attachments/%s", ticket.TicketID)
            os.MkdirAll(uploadDir, os.ModePerm)

            // Save file to backend folder
            savedFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
            filePath := fmt.Sprintf("%s/%s", uploadDir, savedFileName)
            if err := c.SaveFile(file, filePath); err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
                    RetCode: "500",
                    Message: "Failed to save attachment",
                })
            }

            // Save metadata to TicketAttachment table
            attachment := models.TicketAttachment{
                TicketID:   ticket.TicketID,
                FileName:   file.Filename,
                FilePath:   filePath,
                UploadedBy: user.Username,
            }
            middleware.DBConn.Create(&attachment)
        }
    }

    return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
        RetCode: "201",
        Message: "Ticket created successfully",
        Data: fiber.Map{
            "ticket_code": ticket.TicketID,
            "ticket":      ticket,
        },
    })
}

// GetAllTickets returns all tickets (Admin view)
// func GetAllTickets(c *fiber.Ctx) error {
// 	var tickets []models.CreateTicket
// 	if err := middleware.DBConn.Order("created_at desc").Find(&tickets).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to fetch tickets",
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "Tickets fetched successfully",
// 		Data:    tickets,
// 	})
// }

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

// GetUserTickets returns tickets for the logged-in user
func GetUserTickets(c *fiber.Ctx) error {
	var tickets []models.CreateTicket

	// Get user ID from JWT
	userID, err := middleware.GetUserIDFromJWT(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized: User ID not found",
		})
	}

	// Get user info (to get username)
	var user models.UserAccount
	if err := middleware.DBConn.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user info",
		})
	}

	// Fetch tickets by username
	if err := middleware.DBConn.
		Where("username = ?", user.Username).
		Order("created_at desc").
		Find(&tickets).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch tickets",
		})
	}

	// ✅ Check if no tickets found
	if len(tickets) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "No tickets found",
			Data:    []models.CreateTicket{}, // empty array (best practice)
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Tickets fetched successfully",
		Data:    tickets,
	})
}