package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// GetUserID extracts the authenticated user ID from the Fiber context
// Returns an error if the user is not authenticated or the ID is invalid
func GetUserID(c *fiber.Ctx) (int64, error) {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return 0, errors.New("user not authenticated")
	}
	return userID, nil
}
