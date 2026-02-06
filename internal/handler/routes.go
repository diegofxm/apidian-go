package handler

import (
	"apidian-go/internal/config"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/repository"
	"apidian-go/internal/service"
	"apidian-go/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// SetupSystemRoutes configura rutas de sistema (health, metrics, etc.)
func SetupSystemRoutes(app *fiber.App, cfg *config.Config) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.Server.Env,
		})
	})
}

func SetupPublicRoutes(api fiber.Router, db *database.Database, cfg *config.Config) {
	// Auth routes
	auth := api.Group("/auth")
	authHandler := NewAuthHandler(db, cfg)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Public info
	api.Get("/ping", func(c *fiber.Ctx) error {
		return response.Success(c, "pong", fiber.Map{
			"timestamp": c.Context().Time(),
		})
	})
	
	// TEMPORAL: PDF preview sin autenticación para testing
	pdfHandler := NewPDFHandler(db, cfg)
	api.Get("/invoices/pdf/:number", pdfHandler.GenerateInvoicePDFByNumber) // Por número de factura (título visible en navegador)
}

func SetupProtectedRoutes(api fiber.Router, db *database.Database, cfg *config.Config) {
	// Companies CRUD
	companies := api.Group("/companies")
	companyHandler := NewCompanyHandler(db, cfg)
	companies.Get("/", companyHandler.GetAll)
	companies.Get("/:id", companyHandler.GetByID)
	companies.Post("/", companyHandler.Create)
	companies.Put("/:id", companyHandler.Update)
	companies.Delete("/:id", companyHandler.Delete)
	companies.Post("/:id/certificate", companyHandler.UploadCertificate) // Deprecated
	
	// TODO: Implementar certificación ante DIAN (SendTestSetAsync)
	// companies.Post("/:id/certification/testset", companyHandler.SubmitTestSet)     // Enviar set de pruebas (30 FV + 10 NC + 10 ND)
	// companies.Get("/:id/certification/status", companyHandler.GetCertificationStatus) // Consultar estado de certificación (GetStatusZip)

	// Customers (FLAT with company_id filter)
	customers := api.Group("/customers")
	customerHandler := NewCustomerHandler(db)
	customers.Get("/", customerHandler.GetAll)           // ?company_id=1
	customers.Get("/:id", customerHandler.GetByID)
	customers.Post("/", customerHandler.Create)          // company_id in JSON body
	customers.Put("/:id", customerHandler.Update)
	customers.Delete("/:id", customerHandler.Delete)

	// Products (FLAT with company_id filter)
	products := api.Group("/products")
	productHandler := NewProductHandler(db)
	products.Get("/", productHandler.GetAll)             // ?company_id=1
	products.Get("/:id", productHandler.GetByID)
	products.Post("/", productHandler.Create)            // company_id in JSON body
	products.Put("/:id", productHandler.Update)
	products.Delete("/:id", productHandler.Delete)

	// Invoices (FLAT with company_id filter)
	invoices := api.Group("/invoices")
	invoiceHandler := NewInvoiceHandler(db, cfg)
	pdfHandler := NewPDFHandler(db, cfg)
	invoices.Get("/", invoiceHandler.GetAll)                              // ?company_id=1&status=draft
	invoices.Get("/:id", invoiceHandler.GetByID)
	invoices.Post("/", invoiceHandler.Create)                             // company_id in JSON body
	invoices.Put("/:id", invoiceHandler.Update)
	invoices.Delete("/:id", invoiceHandler.Delete)
	invoices.Post("/:id/sign", invoiceHandler.Sign)                       // Firmar factura
	invoices.Post("/:id/send", invoiceHandler.SendToDIAN)                 // Enviar a DIAN (SendBillSync - individual)
	invoices.Post("/:id/status", invoiceHandler.GetInvoiceStatus)         // Consultar estado en DIAN
	api.Get("/invoices/pdf/:number", pdfHandler.GenerateInvoicePDFByNumber) // Por número de factura (título visible)
	invoices.Post("/:id/attached", invoiceHandler.GenerateAttachedDocument) // Generar AttachedDocument
	invoices.Get("/:id/download", invoiceHandler.DownloadZIP)             // Descargar ZIP final
	invoices.Get("/:id/xml", invoiceHandler.GetXML)                       // Obtener XML firmado
	
	// TODO: Implementar envío masivo de facturas (SendBillAsync)
	// invoices.Post("/batch/send", invoiceHandler.SendBatchToDIAN)       // Enviar lote de facturas (retorna ZipKey)
	// invoices.Get("/batch/:zipKey/status", invoiceHandler.GetBatchStatus) // Consultar estado de lote (GetStatusZip)

	// Certificates (FLAT with company_id filter)
	certificates := api.Group("/certificates")
	certificateHandler := NewCertificateHandler(db, cfg)
	certificates.Get("/", certificateHandler.GetByCompanyID)    // ?company_id=1 (active)
	certificates.Get("/all", certificateHandler.GetAllByCompanyID) // ?company_id=1 (history)
	certificates.Get("/:id", certificateHandler.GetByCompanyID)    // Get by ID (deprecated, use /all)
	certificates.Post("/", certificateHandler.Create)              // company_id in JSON body
	certificates.Delete("/:id", certificateHandler.Delete)

	// Resolutions (FLAT with company_id filter)
	resolutions := api.Group("/resolutions")
	resolutionRepo := repository.NewResolutionRepository(db)
	companyRepoForResolution := repository.NewCompanyRepository(db)
	resolutionService := service.NewResolutionService(resolutionRepo, companyRepoForResolution)
	companyServiceForResolution := service.NewCompanyService(companyRepoForResolution)
	resolutionHandler := NewResolutionHandler(resolutionService, companyServiceForResolution)
	resolutions.Get("/", resolutionHandler.GetAll)       // ?company_id=1
	resolutions.Get("/:id", resolutionHandler.GetByID)
	resolutions.Post("/", resolutionHandler.Create)      // company_id in JSON body
	resolutions.Delete("/:id", resolutionHandler.Delete)

	// Software (FLAT with company_id filter)
	software := api.Group("/software")
	softwareRepo := repository.NewSoftwareRepository(db)
	companyRepoForSoftware := repository.NewCompanyRepository(db)
	softwareService := service.NewSoftwareService(softwareRepo, companyRepoForSoftware)
	companyServiceForSoftware := service.NewCompanyService(companyRepoForSoftware)
	softwareHandler := NewSoftwareHandler(softwareService, companyServiceForSoftware)
	software.Get("/", softwareHandler.GetByCompanyID)    // ?company_id=1
	software.Get("/:id", softwareHandler.GetByID)
	software.Post("/", softwareHandler.Create)           // company_id in JSON body
	software.Put("/:id", softwareHandler.Update)
	software.Delete("/:id", softwareHandler.Delete)

	// Users (CRUD de usuarios)
	users := api.Group("/users")
	userHandler := NewUserHandler(db)
	users.Get("/", userHandler.GetAll)
	users.Get("/:id", userHandler.GetByID)
	users.Put("/:id", userHandler.Update)
	users.Delete("/:id", userHandler.Delete)

	// Auth protected routes
	authHandler := NewAuthHandler(db, cfg)
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/me", authHandler.Me)
	api.Post("/auth/change-password", userHandler.ChangePassword)
}
