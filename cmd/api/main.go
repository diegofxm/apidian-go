package main

import (
	"apidian-go/internal/config"
	"apidian-go/internal/handler"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/middleware"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Configurar timezone de Colombia (America/Bogota = UTC-5)
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		log.Printf("Warning: Could not load America/Bogota timezone: %v. Using UTC-5 offset", err)
		// Fallback: crear timezone con offset fijo UTC-5
		loc = time.FixedZone("COT", -5*60*60)
	}
	time.Local = loc
	log.Printf("✓ Timezone configured: %s (UTC%s)", loc.String(), time.Now().Format("-07:00"))

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Conectar a base de datos
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database connected successfully")

	// Crear aplicación Fiber
	app := fiber.New(fiber.Config{
		AppName:      "APIDIAN API v0.1.0",
		ErrorHandler: customErrorHandler,
	})

	// Middleware globales
	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.CORS(cfg.Server.AllowOrigins))
	app.Use(middleware.RateLimiter(cfg.Server.Env))
	app.Use(middleware.ErrorHandler())

	// Rutas de sistema (health, metrics, etc.)
	handler.SetupSystemRoutes(app, cfg)

	// Rutas API
	api := app.Group("/api/v1")

	// Rutas públicas
	handler.SetupPublicRoutes(api, db, cfg)

	// Rutas protegidas (requieren autenticación)
	protected := api.Group("", middleware.AuthMiddleware(&cfg.JWT, db))
	handler.SetupProtectedRoutes(protected, db, cfg)

	// Iniciar servidor
	port := ":" + cfg.Server.Port
	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
