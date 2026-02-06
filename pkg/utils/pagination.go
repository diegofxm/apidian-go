package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// NormalizePagination normalizes pagination parameters
// Ensures page >= 1 and pageSize is between 1 and 100 (default: 10)
func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}

// CalculateOffset calculates the SQL offset based on page and pageSize
func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// ParsePaginationParams extracts and normalizes pagination parameters from Fiber context
// Returns normalized page and pageSize values
func ParsePaginationParams(c *fiber.Ctx) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	pageSize, _ = strconv.Atoi(c.Query("page_size", "10"))
	return NormalizePagination(page, pageSize)
}
