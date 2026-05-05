package main

import (
	"fmt"
	"log"
	"os"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	log.Println("Environment variables loaded")

	// Connect DB
	if middleware.ConnectDB() {
		log.Fatal("❌ Failed to connect to database")
	}

	// Fiber app
	app := fiber.New()
	app.Use(logger.New())

	// CORS
	app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
    AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))


	// ── Static file serving ───────────────────────────────────────────────────
	// Files are saved to ./upload/attachments/xxx on disk.
	// This maps GET /uploads/* → ./upload/* so the frontend can access them.
	// Example: http://localhost:8080/uploads/attachments/SR000003_xxx.png
	app.Static("/uploads", "/var/www/ticketing/uploads")

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
