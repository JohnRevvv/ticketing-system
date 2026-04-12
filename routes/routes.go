package routes

import (
	"ticketing-be-dev/controllers"
	"ticketing-be-dev/middleware"

	"github.com/gofiber/fiber/v2"
)

func AppRoutes(app *fiber.App) {
	api := app.Group("/api")

	// ── Static file serving ───────────────────────────────────────────────────
	// Files are saved to ./upload/attachments/ on disk
	// Served at GET /uploads/attachments/filename
	app.Static("/uploads", "./upload")

	// ==============================
	// Public routes (no token needed)
	// ==============================
	api.Post("/user/register", controllers.Register)
	api.Post("/user/login", controllers.Login)
	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password", controllers.ResetPassword)
	api.Post("/verify-code", controllers.VerifyCode)

	// ==============================
	// User routes (JWT required)
	// ==============================
	userRoutes := api.Group("/user", middleware.JWTMiddleware())

	// ── Ticket CRUD ───────────────────────────────────────────────────────────
	userRoutes.Post("/ticket/create", controllers.CreateTicket)
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
	userRoutes.Put("/ticket/resolve/:id", controllers.ResolveTicket)
	userRoutes.Put("/ticket/cancel/:id", controllers.CancelTicket)

	// ── Users ─────────────────────────────────────────────────────────────────
	userRoutes.Get("/list/all/users", controllers.GetAllUsers)
	userRoutes.Get("/get/me", controllers.GetCurrentUser)
	userRoutes.Put("/update/user/:id", controllers.UpdateUserRoleStatus)
	userRoutes.Put("/update/profile/:id", controllers.UpdateUserProfile)

	// ── Attachments ───────────────────────────────────────────────────────────
	userRoutes.Get("/attachments/:id", controllers.ViewAttachment)
}
