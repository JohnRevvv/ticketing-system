package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ticketing-be-dev/models/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// DO NOT preload secret at global level
func getSecret() []byte {
	return []byte(os.Getenv("SECRET_KEY"))
}

// ─────────────────────────────────────────────
// Generate JWT
// ─────────────────────────────────────────────
func GenerateJWT(ID uint) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = ID
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
	// claims["exp"] = time.Now().Add(30 * time.Minute).Unix()

	return token.SignedString(getSecret())
}

// ─────────────────────────────────────────────
// JWT Middleware (FIXED)
// ─────────────────────────────────────────────
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(401).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "Missing Authorization header",
			})
		}

		// remove Bearer prefix safely
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			return c.Status(401).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "Invalid token format",
			})
		}

		// parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return getSecret(), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "Invalid token claims",
			})
		}

		idFloat, ok := claims["id"].(float64)
		if !ok {
			return c.Status(401).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "User ID missing in token",
			})
		}

		userID := uint(idFloat)

		fmt.Println("AUTH USER ID:", userID)

		c.Locals("user_id", userID)

		return c.Next()
	}
}

// ─────────────────────────────────────────────
// Helper function
// ─────────────────────────────────────────────
func GetUserIDFromJWT(c *fiber.Ctx) (uint, error) {
	id := c.Locals("user_id")

	userID, ok := id.(uint)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}

	return userID, nil
}
