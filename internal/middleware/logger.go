package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Procesar request
		err := c.Next()

		// Calcular latencia
		latency := time.Since(start)

		// Obtener información del request
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Determinar color según status code
		statusColor := getStatusColor(status)
		methodColor := getMethodColor(method)

		// Formato profesional con separadores
		fmt.Printf(
			"%s | %s%-6s%s | %s%3d%s | %12v | %-15s | %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			methodColor,
			method,
			resetColor(),
			statusColor,
			status,
			resetColor(),
			latency,
			ip,
			path,
		)

		// Log adicional para errores
		if status >= 400 {
			fmt.Printf("     └─ User-Agent: %s\n", truncateString(userAgent, 80))
		}

		return err
	}
}

// getStatusColor retorna el color ANSI según el status code
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\033[32m" // Verde
	case status >= 300 && status < 400:
		return "\033[36m" // Cyan
	case status >= 400 && status < 500:
		return "\033[33m" // Amarillo
	case status >= 500:
		return "\033[31m" // Rojo
	default:
		return "\033[37m" // Blanco
	}
}

// getMethodColor retorna el color ANSI según el método HTTP
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // Azul
	case "POST":
		return "\033[32m" // Verde
	case "PUT":
		return "\033[33m" // Amarillo
	case "DELETE":
		return "\033[31m" // Rojo
	case "PATCH":
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // Blanco
	}
}

// resetColor resetea el color ANSI
func resetColor() string {
	return "\033[0m"
}

// truncateString trunca un string a una longitud máxima
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
