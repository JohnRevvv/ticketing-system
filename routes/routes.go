package routes

import (
	"ticketing-be-dev/controllers"
	"ticketing-be-dev/middleware"

	"github.com/gofiber/fiber/v2"
)

func AppRoutes(app *fiber.App) {
	api := app.Group("/api")

	// ==============================
	// Public routes (no token needed)
	// ==============================
	api.Post("/user/register", controllers.Register)
	api.Post("/user/login", controllers.Login)

	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password", controllers.ResetPassword)

	// ==============================
	// Admin routes (JWT required)
	// ==============================
	adminRoutes := api.Group("/admin", middleware.JWTMiddleware())
	adminRoutes.Get("/list/all/users", controllers.GetAllUsers)
	adminRoutes.Put("/update/user/:id", controllers.UpdateUserRoleStatus)
	adminRoutes.Get("/list/all/tickets", controllers.GetAllTickets)
	adminRoutes.Get("/tickets/:id", controllers.GetTicketByID)

	// ==============================
	// User routes (JWT required)
	// ==============================
	userRoutes := api.Group("/user", middleware.JWTMiddleware())
	userRoutes.Post("/ticket/create", controllers.CreateTicket)
	userRoutes.Get("/list/my/tickets", controllers.GetUserTickets)
	userRoutes.Put("/ticket/endorse/:id", controllers.EndorseTicket)
	userRoutes.Put("/ticket/approve/:id", controllers.ApproveTicket)
	userRoutes.Put("/ticket/grab/:id", controllers.GrabTicket)
	userRoutes.Put("/ticket/resolve/:id", controllers.ResolveTicket)
	userRoutes.Put("/ticket/cancel/:id", controllers.CancelTicket)
	userRoutes.Get("/tickets/:id", controllers.GetTicketByID)
}
