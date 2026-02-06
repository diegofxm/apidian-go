package middleware

import (
	"apidian-go/internal/config"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/pkg/response"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	CompanyID int64  `json:"company_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware valida JWT
func AuthMiddleware(cfg *config.JWTConfig, db *database.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Unauthorized(c, "Invalid authorization header format")
		}

		tokenString := parts[1]

		// Validar JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return response.Unauthorized(c, "Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.Unauthorized(c, "Invalid token claims")
		}

		// Extraer user_id y email de los claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return response.Unauthorized(c, "Invalid user_id in token")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return response.Unauthorized(c, "Invalid email in token")
		}

		// Guardar información del usuario en el contexto
		c.Locals("user_id", int64(userID))
		c.Locals("email", email)

		return c.Next()
	}
}

func GenerateToken(userID, companyID int64, email string, cfg *config.JWTConfig) (string, error) {
	claims := Claims{
		UserID:    userID,
		CompanyID: companyID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.Expiration))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
