package middleware

import (
	"fmt"
	"log"
	"os"
	"ticketing-be-dev/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

func GetEnv(key string) string {
	return os.Getenv(key)
}

func ConnectDB() bool {
	dsn := fmt.Sprintf(
	"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=%s",
	os.Getenv("DB_HOST"),
	os.Getenv("DB_PORT"),
	os.Getenv("DB_NAME"),
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_SSLMODE"),   // MUST be exactly as in .env
	os.Getenv("DB_TIMEZONE"),  // MUST be exactly as in .env
)

	fmt.Println("DEBUG DSN:", dsn) // <--- check DSN

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("❌ Database connection error:", err)
		return true
	}

	DBConn = db

	if err := DBConn.AutoMigrate(
		&models.AdminAccount{},
		&models.UserAccount{},
		&models.CreateTicket{},
		&models.PasswordResetToken{});
		err != nil {
		log.Println("❌ Auto-migration failed:", err)
		return true
	}

	log.Println("✅ Database connected and migrated")
	return false
}