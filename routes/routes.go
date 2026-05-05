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

	// ── Static file serving ───────────────────────────────────────────────────
	// Custom handler: URL-decodes the path before serving,
	// fixing filenames with spaces or special characters.
	app.Get("/uploads/*", func(c *fiber.Ctx) error {
		rawPath := c.Params("*")

		// Decode %20, %C3%A2%C2%80%C2%AF, etc.
		decoded, err := url.PathUnescape(rawPath)
		if err != nil {
			decoded = rawPath
		}

		filePath := filepath.Join("./uploads", decoded)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("File not found: " + decoded)
		}

		return c.SendFile(filePath)
	})

	// ==============================
	// Public routes (no token needed)
	// ==============================
	public := app.Group("/api/public/v1")

	public.Post("/login", controllers.Login)
	public.Post("/register", controllers.Register)
	public.Post("/forgot-password", controllers.ForgotPassword)
	public.Post("/reset-password", controllers.ResetPassword)
	public.Post("/verify-code", controllers.VerifyCode)

	// ==============================
	// User routes (JWT required)
	// ==============================
	userRoutes := public.Group("/user", middleware.JWTMiddleware())

	// ── Ticket CRUD ───────────────────────────────────────────────────────────
	userRoutes.Post("/ticket/create", controllers.CreateTicket)
	userRoutes.Put("/ticket/update/:id", controllers.UpdateTicket)
	userRoutes.Get("/list/my/tickets", controllers.GetUserTickets)
	userRoutes.Get("/list/all/tickets", controllers.GetAllTickets)

	// ── Export — MUST be before /tickets/:id to avoid param conflict ──────────
	userRoutes.Get("/tickets/export", controllers.ExportTicketsCSV)

	// ── Get ticket by ID — registered AFTER /tickets/export ──────────────────
	userRoutes.Get("/tickets/:id", controllers.GetTicketByID)

	// ── Ticket actions ────────────────────────────────────────────────────────
	userRoutes.Put("/ticket/endorse/:id", controllers.EndorseTicket)
	userRoutes.Put("/ticket/approve/:id", controllers.ApproveTicket)
	userRoutes.Put("/ticket/grab/:id", controllers.GrabTicket)
	userRoutes.Put("/ticket/ungrab/:id", controllers.UnGrabTicket)
	userRoutes.Put("/ticket/resolve/:id", controllers.ResolveTicket)
	userRoutes.Put("/ticket/cancel/:id", controllers.CancelTicket)
	userRoutes.Patch("/ticket/hold/:id", controllers.HoldTicket)
	userRoutes.Patch("/ticket/unhold/:id", controllers.ResumeTicket)

	// ── Users ─────────────────────────────────────────────────────────────────
	userRoutes.Get("/list/all/users", controllers.GetAllUsers)
	userRoutes.Get("/get/me", controllers.GetCurrentUser)
	userRoutes.Put("/update/user/:id", controllers.UpdateUserRoleStatus)
	userRoutes.Put("/update/profile/:id", controllers.UpdateUserProfile)

	// ── Attachments ───────────────────────────────────────────────────────────
	userRoutes.Get("/attachments/:id", controllers.ViewAttachment)

	// ── Remarks ───────────────────────────────────────────────────────────────
	userRoutes.Post("/ticket/remark", controllers.CreateTicketRemark)
	userRoutes.Get("/ticket/:ticket_id/remarks", controllers.GetRemarksByTicket)

	// ── Categories & Subcategories ───────────────────────────────────────────
	userRoutes.Post("/add-category", controllers.AddCategory)
	userRoutes.Post("/add-sub-category", controllers.AddSubCategory)
	userRoutes.Get("/categories", controllers.GetCategories)
	userRoutes.Get("/categories/:id/sub-categories", controllers.GetSubCategoriesByCategory)
	userRoutes.Patch("/subcategories/:id/description", controllers.UpdateSubCategoryDescription)
	userRoutes.Put("/update-categories/:id", controllers.UpdateCategoryName)
	userRoutes.Put("/update-sub-categories/:id", controllers.UpdateSubCategoryName)
	userRoutes.Delete("/delete-category/:id", controllers.DeleteCategory)
	userRoutes.Delete("/delete-subcategory/:id", controllers.DeleteSubCategory)

}
