package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter retorna middleware de rate limiting
// - Development: 1000 requests/minuto
// - Production: 100 requests/minuto
func RateLimiter(env string) fiber.Handler {
	max := 100
	if env == "development" {
		max = 1000 // Más permisivo en desarrollo
	}

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Limitar por IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many requests, please try again later",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Usa memoria por defecto
	})
}
