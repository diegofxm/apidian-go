package handler

import (
	"apidian-go/internal/config"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/repository"
	"apidian-go/internal/service/invoice"
	"apidian-go/internal/service/pdf"
	"apidian-go/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type PDFHandler struct {
	invoiceRepo    *repository.InvoiceRepository
	invoiceService *invoice.InvoiceService
	pdfService     *pdf.PDFInvoiceService
}

func NewPDFHandler(db *database.Database, cfg *config.Config) *PDFHandler {
	invoiceRepo := repository.NewInvoiceRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	resolutionRepo := repository.NewResolutionRepository(db)
	productRepo := repository.NewProductRepository(db)
	certificateRepo := repository.NewCertificateRepository(db)

	invoiceService := invoice.NewInvoiceService(
		invoiceRepo,
		companyRepo,
		customerRepo,
		resolutionRepo,
		productRepo,
		certificateRepo,
		&cfg.Storage,
		cfg.Invoice.KeepUnsignedXML,
	)
	
	// Usar Maroto v2 para generar PDFs con diseño profesional
	pdfService := pdf.NewPDFInvoiceService(&cfg.Storage)

	return &PDFHandler{
		invoiceRepo:    invoiceRepo,
		invoiceService: invoiceService,
		pdfService:     pdfService,
	}
}

// GenerateInvoicePDFByNumber genera el PDF usando el número de factura
// URL: /api/v1/invoices/pdf/SETP-99000000
func (h *PDFHandler) GenerateInvoicePDFByNumber(c *fiber.Ctx) error {
	// Get number from URL
	number := c.Params("number")
	if number == "" {
		return response.BadRequest(c, "Invoice number is required")
	}

	// Buscar factura por número
	invoice, err := h.invoiceRepo.GetByNumber(number)
	if err != nil {
		return response.NotFound(c, "Invoice not found")
	}

	// Generar PDF dinámicamente con Maroto
	pdfBytes, err := h.pdfService.GenerateInvoicePDF(invoice)
	if err != nil {
		return response.InternalServerError(c, "Failed to generate PDF: "+err.Error())
	}

	// Configurar headers para visualizar en navegador (inline)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=\""+number+".pdf\"")
	
	return c.Send(pdfBytes)
}
