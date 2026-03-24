package main

import (
	"fmt"
	"log"
	"os"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	} else {
		log.Println("✅ .env loaded")
	}

	// Optional: check if variables are loaded
	fmt.Println("DB_SSLMODE:", os.Getenv("DB_SSLMODE"))
	fmt.Println("DB_TIMEZONE:", os.Getenv("DB_TIMEZONE"))

	// Connect DB
	if middleware.ConnectDB() {
		log.Fatal("❌ Failed to connect to database")
	}

	// Fiber app
	app := fiber.New()
	app.Use(logger.New())

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running 🚀")
	})

	// Routes
	routes.AppRoutes(app)

	port := os.Getenv("PROJ_PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}