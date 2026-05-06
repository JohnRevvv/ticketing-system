package main

import (
	"fmt"
	"log"
	"os"
	"ticketing-be-dev/middleware"
	"ticketing-be-dev/routes"
	"ticketing-be-dev/services"

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

	// Connect DB
	if middleware.ConnectDB() {
		log.Fatal("❌ Failed to connect to database")
	}

	// ✅ Init S3
	if err := services.InitS3(); err != nil {
		log.Fatal("❌ Failed to initialize S3:", err)
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