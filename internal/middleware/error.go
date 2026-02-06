package middleware

import (
	"apidian-go/pkg/response"
	"log"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			log.Printf("Error: %v", err)

			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(response.Response{
					Success: false,
					Error:   e.Message,
				})
			}

			return response.InternalServerError(c, "An unexpected error occurred")
		}

		return nil
	}
}
