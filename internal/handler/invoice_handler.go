package handler

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/repository"
	"apidian-go/internal/service"
	"apidian-go/internal/service/invoice"
	"apidian-go/pkg/errors"
	"apidian-go/pkg/response"
	"apidian-go/pkg/utils"
	"apidian-go/pkg/validator"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type InvoiceHandler struct {
	service        *invoice.InvoiceService
	companyService *service.CompanyService
}

// formatMoney formatea un valor float64 a string con 2 decimales
func formatMoney(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func NewInvoiceHandler(db *database.Database, cfg *config.Config) *InvoiceHandler {
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
	companyService := service.NewCompanyService(companyRepo)

	return &InvoiceHandler{
		service:        invoiceService,
		companyService: companyService,
	}
}

// Create creates a new invoice
func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req domain.CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateCreateInvoice(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create invoice
	invoice, err := h.service.Create(&req, userID)
	if err != nil {
		if err.Error() == "company not found" || err.Error() == "customer not found" || 
		   err.Error() == "resolution not found" || err.Error() == "product not found in line 1" {
			return response.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized access to company" || 
		   err.Error() == "customer does not belong to company" ||
		   err.Error() == "resolution does not belong to company" {
			return response.Unauthorized(c, err.Error())
		}
		// TEMPORAL: Mostrar error completo para debugging
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "Invoice created successfully", invoice)
}

// GetByID gets an invoice by ID
func (h *InvoiceHandler) GetByID(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Get invoice
	invoice, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, errors.ErrInvoiceNotFound.Message)
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Invoice retrieved successfully", invoice)
}

// GetAll gets all invoices for a company using NIT/DV
func (h *InvoiceHandler) GetAll(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company_id from query params
	companyIDStr := c.Query("company_id")
	if companyIDStr == "" {
		return response.BadRequest(c, "company_id query parameter is required")
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid company_id")
	}

	// Verify company belongs to user
	_, err = h.companyService.GetByID(companyID, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.BadRequest(c, err.Error())
	}

	// Get pagination parameters
	page, pageSize := utils.ParsePaginationParams(c)
	limit := pageSize
	offset := utils.CalculateOffset(page, pageSize)

	// Get invoices
	invoices, err := h.service.GetByCompanyID(companyID, userID, limit, offset)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Invoices retrieved successfully", invoices)
}

// Update updates an invoice
func (h *InvoiceHandler) Update(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	var req domain.UpdateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateUpdateInvoice(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Update invoice
	if err := h.service.Update(id, &req, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, errors.ErrInvoiceNotFound.Message)
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		if err.Error() == "only draft invoices can be updated" {
			return response.BadRequest(c, "Only draft invoices can be updated")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Invoice updated successfully", nil)
}

// Delete deletes an invoice
func (h *InvoiceHandler) Delete(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Delete invoice
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, errors.ErrInvoiceNotFound.Message)
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		if err.Error() == "only draft invoices can be deleted" {
			return response.BadRequest(c, "Only draft invoices can be deleted")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Invoice deleted successfully", nil)
}

// Sign signs an invoice
func (h *InvoiceHandler) Sign(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Sign invoice
	if err := h.service.Sign(id, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		if err.Error() == "only draft invoices can be signed" {
			return response.BadRequest(c, "Only draft invoices can be signed")
		}
		return response.InternalServerError(c, err.Error())
	}

	// Obtener factura firmada con toda la información
	invoice, err := h.service.GetByID(id, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve signed invoice")
	}

	// Construir respuesta completa
	data := &domain.DocumentData{
		InvoiceID:     invoice.ID,
		Number:        invoice.Number,
		URLInvoiceXML: "FES-" + invoice.Number + ".xml",
		URLInvoicePDF: "FES-" + invoice.Number + ".pdf",
	}

	// Agregar CUFE si existe
	if invoice.UUID != nil && *invoice.UUID != "" {
		data.CUFE = *invoice.UUID
		
		// Generar QR string si tenemos CUFE
		// Formato: NumFac: X\nFecFac: Y\nNitFac: Z\n...
		qrStr := "NumFac: " + invoice.Number + "\n"
		qrStr += "FecFac: " + invoice.IssueDate.Format("2006-01-02") + "\n"
		qrStr += "NitFac: " + invoice.Company.NIT + "\n"
		qrStr += "DocAdq: " + invoice.Customer.IdentificationNumber + "\n"
		qrStr += "ValFac: " + formatMoney(invoice.Subtotal) + "\n"
		qrStr += "ValIva: " + formatMoney(invoice.TaxTotal) + "\n"
		qrStr += "ValOtroIm: 0.00\n"
		qrStr += "ValTotal: " + formatMoney(invoice.Total) + "\n"
		qrStr += "CUFE: " + *invoice.UUID + "\n"
		qrStr += "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=" + *invoice.UUID
		
		data.QRStr = qrStr
	}

	// Construir respuesta usando DocumentResponse
	resp := domain.NewSuccessResponse("Factura #"+invoice.Number+" firmada con éxito", data)
	
	return c.Status(fiber.StatusOK).JSON(resp)
}

// SendToDIAN sends an invoice to DIAN
func (h *InvoiceHandler) SendToDIAN(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Send to DIAN
	if err := h.service.SendToDIAN(id, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		if err.Error() == "only signed invoices can be sent to DIAN" {
			return response.BadRequest(c, "Only signed invoices can be sent to DIAN")
		}
		// Verificar si es un rechazo de DIAN (error de negocio)
		if strings.HasPrefix(err.Error(), "DIAN_REJECTION:") {
			// Extraer el mensaje después de "DIAN_REJECTION: "
			message := strings.TrimPrefix(err.Error(), "DIAN_REJECTION: ")
			// HTTP 422 Unprocessable Entity para errores de negocio de DIAN
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"error":   message,
			})
		}
		// Otros errores son errores técnicos (500)
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "Invoice sent to DIAN successfully", nil)
}

// GeneratePDF generates the PDF for a signed invoice
func (h *InvoiceHandler) GeneratePDF(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Generate PDF
	if err := h.service.GeneratePDF(id, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "PDF generated successfully", nil)
}

// GenerateAttachedDocument generates the AttachedDocument for client
func (h *InvoiceHandler) GenerateAttachedDocument(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Generate AttachedDocument
	if err := h.service.GenerateAttachedDocument(id, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "AttachedDocument generated successfully", nil)
}

// DownloadZIP downloads the final ZIP file for client
func (h *InvoiceHandler) DownloadZIP(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Get ZIP path
	zipPath, err := h.service.DownloadInvoiceZip(id, userID)
	if err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, err.Error())
	}

	// Send file
	return c.SendFile(zipPath)
}

// GetXML returns the signed XML of an invoice
func (h *InvoiceHandler) GetXML(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Get XML content
	xmlContent, err := h.service.GetInvoiceXML(id, userID)
	if err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, err.Error())
	}

	// Set content type and return XML
	c.Set("Content-Type", "application/xml")
	return c.Send(xmlContent)
}

// GetPDF returns the PDF of an invoice
func (h *InvoiceHandler) GetPDF(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Get PDF path
	pdfPath, err := h.service.GetInvoicePDF(id, userID)
	if err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		return response.InternalServerError(c, err.Error())
	}

	// Send file
	return c.SendFile(pdfPath)
}

// GetInvoiceStatus consulta el estado de una factura en DIAN
func (h *InvoiceHandler) GetInvoiceStatus(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get invoice ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID")
	}

	// Parse request body
	var req struct {
		TrackId string `json:"track_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.TrackId == "" {
		return response.BadRequest(c, "track_id is required")
	}

	// Call service
	if err := h.service.GetInvoiceStatus(id, req.TrackId, userID); err != nil {
		if err.Error() == "invoice not found" {
			return response.NotFound(c, "Invoice not found")
		}
		if err.Error() == "unauthorized access to invoice" {
			return response.Unauthorized(c, "Unauthorized access to invoice")
		}
		if strings.Contains(err.Error(), "DIAN rejected") {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "Invoice status updated successfully", nil)
}
