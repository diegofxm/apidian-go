package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS retorna middleware CORS configurado desde variables de entorno
// - allowOrigins: Dominios permitidos (separados por coma) o "*" para todos
func CORS(allowOrigins string) fiber.Handler {
	config := cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400,
	}

	// Si es "*", no se pueden usar credentials
	if allowOrigins == "*" {
		config.AllowCredentials = false
	}

	return cors.New(config)
}
