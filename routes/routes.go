package routes

import (
	"net/url"
	"os"
	"path/filepath"

	"ticketing-be-dev/controllers"
	"ticketing-be-dev/middleware"

	"github.com/gofiber/fiber/v2"
)

func AppRoutes(app *fiber.App) {

	// =========================
	// API VERSIONING
	// =========================
	api := app.Group("/api/v1")

	// =========================
	// STATIC FILES
	// =========================
	app.Get("/uploads/*", func(c *fiber.Ctx) error {
		rawPath := c.Params("*")

		decoded, err := url.PathUnescape(rawPath)
		if err != nil {
			decoded = rawPath
		}

		filePath := filepath.Join("./uploads", decoded)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).
				SendString("File not found: " + decoded)
		}

		return c.SendFile(filePath)
	})

	// =========================
	// PUBLIC ROUTES (NO AUTH)
	// =========================
	auth := api.Group("/auth")

	auth.Post("/login", controllers.Login)
	auth.Post("/register", controllers.Register)
	auth.Post("/forgot-password", controllers.ForgotPassword)
	auth.Post("/reset-password", controllers.ResetPassword)
	auth.Post("/verify-code", controllers.VerifyCode)

	// =========================
	// PROTECTED ROUTES (JWT)
	// =========================
	protected := api.Group("", middleware.JWTMiddleware())

	// =========================
	// USERS
	// =========================
	users := protected.Group("/users")

	users.Get("/me", controllers.GetCurrentUser)
	users.Get("/", controllers.GetAllUsers)
	users.Put("/:id/role", controllers.UpdateUserRoleStatus)
	users.Put("/:id/profile", controllers.UpdateUserProfile)

	// =========================
	// TICKETS (CORE MODULE)
	// =========================
	tickets := protected.Group("/tickets")

	tickets.Post("/", controllers.CreateTicket)
	tickets.Get("/my", controllers.GetUserTickets)
	tickets.Get("/", controllers.GetAllTickets)
	tickets.Get("/export", controllers.ExportTicketsCSV)

	tickets.Get("/:id", controllers.GetTicketByID)

	tickets.Put("/:id", controllers.UpdateTicket)
	tickets.Put("/:id/endorse", controllers.EndorseTicket)
	tickets.Put("/:id/approve", controllers.ApproveTicket)
	tickets.Put("/:id/grab", controllers.GrabTicket)
	tickets.Put("/:id/ungrab", controllers.UnGrabTicket)
	tickets.Put("/:id/resolve", controllers.ResolveTicket)
	tickets.Put("/:id/cancel", controllers.CancelTicket)
	tickets.Patch("/:id/hold", controllers.HoldTicket)
	tickets.Patch("/:id/unhold", controllers.ResumeTicket)

	// =========================
	// TICKET REMARKS
	// =========================
	remarks := tickets.Group("/:id/remarks")

	remarks.Post("/", controllers.CreateTicketRemark)
	remarks.Get("/", controllers.GetRemarksByTicket)

	// =========================
	// ATTACHMENTS
	// =========================
	protected.Get("/attachments/:id", controllers.ViewAttachment)

	// =========================
	// CATEGORIES
	// =========================
	categories := protected.Group("/categories")

	categories.Post("/", controllers.AddCategory)
	categories.Get("/", controllers.GetCategories)
	categories.Put("/:id", controllers.UpdateCategoryName)
	categories.Delete("/:id", controllers.DeleteCategory)

	categories.Get("/:id/subcategories", controllers.GetSubCategoriesByCategory)

	subcategories := protected.Group("/subcategories")

	subcategories.Post("/", controllers.AddSubCategory)
	subcategories.Put("/:id", controllers.UpdateSubCategoryName)
	subcategories.Patch("/:id/description", controllers.UpdateSubCategoryDescription)
	subcategories.Delete("/:id", controllers.DeleteSubCategory)
}