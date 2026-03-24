package routes

import (
	"ticketing-be-dev/controllers"
	"github.com/gofiber/fiber/v2"
)

func AppRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Admin registration (public, no token)
	api.Post("/admin/register", controllers.RegisterAdmin)
    api.Post("/admin/login", controllers.LoginAdmin)


    	// User routes
	api.Post("/user/register", controllers.RegisterUser)
	api.Post("/user/login", controllers.LoginUser)
    
}