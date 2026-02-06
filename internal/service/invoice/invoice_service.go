package invoice

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/crypto"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/diegofxm/ubl21-dian/signature"
	"github.com/diegofxm/ubl21-dian/soap"
	"github.com/diegofxm/ubl21-dian/soap/types"
)

type InvoiceService struct {
	invoiceRepo     *repository.InvoiceRepository
	companyRepo     *repository.CompanyRepository
	customerRepo    *repository.CustomerRepository
	resolutionRepo  *repository.ResolutionRepository
	productRepo     *repository.ProductRepository
	certificateRepo *repository.CertificateRepository
	storage         *config.StorageConfig
	keepUnsignedXML bool
}

func NewInvoiceService(
	invoiceRepo *repository.InvoiceRepository,
	companyRepo *repository.CompanyRepository,
	customerRepo *repository.CustomerRepository,
	resolutionRepo *repository.ResolutionRepository,
	productRepo *repository.ProductRepository,
	certificateRepo *repository.CertificateRepository,
	storage *config.StorageConfig,
	keepUnsignedXML bool,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:     invoiceRepo,
		companyRepo:     companyRepo,
		customerRepo:    customerRepo,
		resolutionRepo:  resolutionRepo,
		productRepo:     productRepo,
		certificateRepo: certificateRepo,
		storage:         storage,
		keepUnsignedXML: keepUnsignedXML,
	}
}

// Create crea una nueva factura con validaciones de negocio
func (s *InvoiceService) Create(req *domain.CreateInvoiceRequest, userID int64) (*domain.Invoice, error) {
	// Validar que la empresa pertenezca al usuario
	company, err := s.companyRepo.GetByID(req.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("company not found")
	}
	if company.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to company")
	}

	// Validar que el cliente pertenezca a la empresa
	customer, err := s.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found")
	}
	if customer.CompanyID != req.CompanyID {
		return nil, fmt.Errorf("customer does not belong to company")
	}

	// Validar que la resolución pertenezca a la empresa
	resolution, err := s.resolutionRepo.GetByID(req.ResolutionID)
	if err != nil {
		return nil, fmt.Errorf("resolution not found")
	}
	if resolution.CompanyID != req.CompanyID {
		return nil, fmt.Errorf("resolution does not belong to company")
	}

	// Validar que la resolución esté activa
	if !resolution.IsActive {
		return nil, fmt.Errorf("resolution is not active")
	}

	// Obtener y actualizar consecutivo de forma atómica desde resolutions.current_number
	nextConsecutive, err := s.resolutionRepo.GetAndIncrementConsecutive(req.ResolutionID)
	if err != nil {
		return nil, fmt.Errorf("error getting consecutive: %w", err)
	}

	// Parsear fechas en la zona horaria local (Colombia)
	issueDate, err := time.ParseInLocation("2006-01-02", req.IssueDate, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid issue_date format, use YYYY-MM-DD")
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.ParseInLocation("2006-01-02", *req.DueDate, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &parsed
	}

	// Calcular payment_form_id automáticamente si no viene en el request
	paymentFormID := req.PaymentFormID
	if paymentFormID == nil {
		// Determinar automáticamente basándose en las fechas
		if dueDate == nil || dueDate.Equal(issueDate) {
			// Contado: sin due_date o mismo día
			contado := 1
			paymentFormID = &contado
		} else {
			// Crédito: due_date posterior a issue_date
			credito := 2
			paymentFormID = &credito
		}
	}

	// Calcular totales de las líneas
	var subtotal, taxTotal float64
	var lines []domain.InvoiceLine

	for i, lineReq := range req.Lines {
		// Validar que el producto pertenezca a la empresa
		product, err := s.productRepo.GetByID(lineReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found in line %d", i+1)
		}
		if product.CompanyID != req.CompanyID {
			return nil, fmt.Errorf("product in line %d does not belong to company", i+1)
		}

		// Usar description del producto si no se proporciona
		description := ""
		if lineReq.Description != nil && *lineReq.Description != "" {
			description = *lineReq.Description
		} else if product.Description != nil {
			description = *product.Description
		}

		// Usar unit_price del producto si no se proporciona
		unitPrice := product.Price
		if lineReq.UnitPrice != nil {
			unitPrice = *lineReq.UnitPrice
		}

		// Usar tax_rate del producto si no se proporciona
		taxRate := product.TaxRate
		if lineReq.TaxRate != nil {
			taxRate = *lineReq.TaxRate
		}

		// Calcular totales de la línea
		lineTotal := lineReq.Quantity * unitPrice
		taxAmount := lineTotal * (taxRate / 100)

		subtotal += lineTotal
		taxTotal += taxAmount

		line := domain.InvoiceLine{
			ProductID:          lineReq.ProductID,
			Description:        description,
			Quantity:           lineReq.Quantity,
			UnitPrice:          unitPrice,
			LineTotal:          lineTotal,
			TaxRate:            taxRate,
			TaxAmount:          taxAmount,
			BrandName:          lineReq.BrandName,
			ModelName:          lineReq.ModelName,
			StandardItemCode:   lineReq.StandardItemCode,
			ClassificationCode: lineReq.ClassificationCode,
		}
		lines = append(lines, line)
	}

	total := subtotal + taxTotal

	// Generar número de factura (formato: PREFIX + consecutivo)
	number := fmt.Sprintf("%s%d", resolution.Prefix, nextConsecutive)

	// Crear factura
	invoice := &domain.Invoice{
		CompanyID:       req.CompanyID,
		CustomerID:      req.CustomerID,
		ResolutionID:    req.ResolutionID,
		Number:          number,
		Consecutive:     nextConsecutive,
		IssueDate:       issueDate,
		IssueTime:       time.Now(),
		DueDate:         dueDate,
		TypeDocumentID:  1, // 1 = Factura
		CurrencyCodeID:  req.CurrencyCodeID,
		Notes:           req.Notes,
		PaymentMethodID: req.PaymentMethodID,
		PaymentFormID:   paymentFormID,
		Subtotal:        subtotal,
		TaxTotal:        taxTotal,
		Total:           total,
		Status:          "draft",
	}

	// Guardar en base de datos
	if err := s.invoiceRepo.Create(invoice, lines); err != nil {
		return nil, fmt.Errorf("error creating invoice: %w", err)
	}

	return invoice, nil
}

// GetByID obtiene una factura por ID validando permisos
func (s *InvoiceService) GetByID(id int64, userID int64) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validar que la empresa de la factura pertenezca al usuario
	company, err := s.companyRepo.GetByID(invoice.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("company not found")
	}
	if company.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to invoice")
	}

	return invoice, nil
}

// GetByCompanyID gets all invoices for a company
func (s *InvoiceService) GetByCompanyID(companyID int64, userID int64, limit, offset int) (*domain.InvoiceListResponse, error) {
	// Validate that the company belongs to the user
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("company not found")
	}
	if company.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to company")
	}

	invoices, total, err := s.invoiceRepo.GetByCompanyID(companyID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Calculate page from offset and limit
	page := (offset / limit) + 1
	if limit == 0 {
		page = 1
	}

	return &domain.InvoiceListResponse{
		Invoices: invoices,
		Total:    int(total),
		Page:     page,
		PageSize: limit,
	}, nil
}

// Update actualiza una factura (solo campos editables)
func (s *InvoiceService) Update(id int64, req *domain.UpdateInvoiceRequest, userID int64) error {
	// Obtener factura
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Solo se puede editar si está en draft
	if invoice.Status != "draft" {
		return fmt.Errorf("only draft invoices can be updated")
	}

	// Parsear fechas si se proporcionan (en zona horaria local)
	if req.DueDate != nil {
		parsed, err := time.ParseInLocation("2006-01-02", *req.DueDate, time.Local)
		if err != nil {
			return fmt.Errorf("invalid due_date format, use YYYY-MM-DD")
		}
		invoice.DueDate = &parsed
	}

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if req.PaymentMethodID != nil {
		invoice.PaymentMethodID = req.PaymentMethodID
	}

	return s.invoiceRepo.Update(invoice)
}

// Delete elimina una factura (solo si está en draft)
func (s *InvoiceService) Delete(id int64, userID int64) error {
	// Validar permisos
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Solo se puede eliminar si está en draft
	if invoice.Status != "draft" {
		return fmt.Errorf("only draft invoices can be deleted")
	}

	return s.invoiceRepo.Delete(id)
}

// Sign firma una factura electrónicamente con certificado digital
func (s *InvoiceService) Sign(id int64, userID int64) error {
	// 1. Obtener factura completa con JOINs
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 2. Validar estado
	if invoice.Status != "draft" {
		return fmt.Errorf("only draft invoices can be signed (current status: '%s')", invoice.Status)
	}

	// 3. Validar datos para DIAN
	if err := ValidateInvoiceForDIAN(invoice); err != nil {
		return fmt.Errorf("invoice validation failed: %w", err)
	}

	// 3.1. Actualizar IssueDate e IssueTime al momento de firma
	// IMPORTANTE: DIAN valida que IssueDate/IssueTime = SigningTime (regla FAD09e)
	now := time.Now()
	invoice.IssueDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	invoice.IssueTime = now
	
	// Actualizar en BD
	if err := s.invoiceRepo.UpdateIssueDateAndTime(id, invoice.IssueDate, invoice.IssueTime); err != nil {
		return fmt.Errorf("failed to update issue date/time: %w", err)
	}

	// 4. Generar XML sin firma usando TEMPLATES (nuevo sistema)
	xmlUnsignedBytes, cufe, err := s.BuildInvoiceWithTemplates(invoice)
	if err != nil {
		return fmt.Errorf("error generating XML with templates: %w", err)
	}

	// 6. Crear directorio de storage para la factura
	invoiceDir := s.storage.InvoicePath(invoice.Company.NIT, invoice.Number)
	if err := os.MkdirAll(invoiceDir, 0755); err != nil {
		return fmt.Errorf("error creating invoice directory: %w", err)
	}

	// 7. Guardar XML sin firma (opcional, solo si keepUnsignedXML está activado)
	unsignedPath := s.storage.InvoiceXMLPath(invoice.Company.NIT, invoice.Number)
	if err := os.WriteFile(unsignedPath, xmlUnsignedBytes, 0644); err != nil {
		return fmt.Errorf("error saving unsigned XML: %w", err)
	}

	// 8. Obtener certificado activo de la empresa
	cert, err := s.certificateRepo.GetByCompanyID(invoice.CompanyID)
	if err != nil {
		return fmt.Errorf("no certificate found for company: %w", err)
	}
	
	// 8.1. Desencriptar contraseña del certificado
	decryptedPassword, err := crypto.DecryptPassword(cert.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt certificate password: %w", err)
	}
	
	// 8.2. Construir ruta al certificado P12
	certPath := s.storage.CertificatePath(invoice.Company.NIT, cert.Name)
	
	// 9. Crear signer desde certificado P12 con fallback automático a PEM
	// Si el P12 está en formato BER (no DER), automáticamente lo convierte a PEM usando OpenSSL
	signer, err := signature.NewSignerFromP12WithFallback(certPath, decryptedPassword)
	if err != nil {
		return fmt.Errorf("error loading certificate: %w\n\nMake sure OpenSSL is installed: apt-get install openssl", err)
	}

	// 10. NO envolver - firmar el XML tal como se genera
	// WrapInvoiceWithFixedNamespaces destruye el formato y cambia el hash
	// La canonicalización C14N debe manejar el formato automáticamente
	xmlWrapped := xmlUnsignedBytes

	// 11. Firmar XML con namespaces fijos
	xmlSignedBytes, err := signer.SignXML(xmlWrapped)
	if err != nil {
		return fmt.Errorf("error signing XML: %w", err)
	}

	// 12. Guardar XML firmado
	signedPath := s.storage.InvoiceSignedXMLPath(invoice.Company.NIT, invoice.Number)
	if err := os.WriteFile(signedPath, xmlSignedBytes, 0644); err != nil {
		return fmt.Errorf("error saving signed XML: %w", err)
	}

	// 13. Eliminar XML sin firmar si keepUnsignedXML es false
	if !s.keepUnsignedXML {
		if err := os.Remove(unsignedPath); err != nil {
			// Log el error pero no fallar el proceso de firma
			fmt.Printf("Warning: could not delete unsigned XML: %v\n", err)
		}
	}

	// 14. Actualizar BD con UUID (CUFE), xml_path y status
	if err := s.invoiceRepo.UpdateStatus(id, "signed"); err != nil {
		return err
	}

	// Actualizar UUID y XML path
	if err := s.updateInvoiceMetadata(id, cufe, signedPath); err != nil {
		return fmt.Errorf("error updating invoice metadata: %w", err)
	}

	return nil
}

// SendToDIAN envía una factura firmada a la DIAN vía SOAP
func (s *InvoiceService) SendToDIAN(id int64, userID int64) error {
	// 1. Obtener factura completa
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 2. Validar estado
	if invoice.Status != "signed" {
		return fmt.Errorf("only signed invoices can be sent to DIAN")
	}

	// 3. Validar que tenga XML firmado
	if invoice.XMLPath == nil || *invoice.XMLPath == "" {
		return fmt.Errorf("invoice does not have signed XML")
	}

	// 4. Leer XML firmado
	xmlSigned, err := os.ReadFile(*invoice.XMLPath)
	if err != nil {
		return fmt.Errorf("error reading signed XML: %w", err)
	}

	// 5. Crear ZIP con el XML firmado
	zipPath := s.storage.InvoiceZIPPath(invoice.Company.NIT, invoice.Number)
	if err := s.createZipFile(zipPath, fmt.Sprintf("FES-%s.xml", invoice.Number), xmlSigned); err != nil {
		return fmt.Errorf("error creating ZIP: %w", err)
	}

	// 6. Leer ZIP y convertir a Base64
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return fmt.Errorf("error reading ZIP: %w", err)
	}
	zipBase64 := base64.StdEncoding.EncodeToString(zipData)

	// 7. Determinar ambiente DIAN
	var environment types.Environment
	if invoice.Software.Environment == "1" {
		environment = types.Produccion
	} else {
		environment = types.Habilitacion
	}

	// 8. Obtener certificado para SOAP security header
	cert, err := s.certificateRepo.GetByCompanyID(invoice.CompanyID)
	if err != nil {
		return fmt.Errorf("no certificate found for company: %w", err)
	}
	
	// 8.1. Desencriptar contraseña del certificado
	decryptedPassword, err := crypto.DecryptPassword(cert.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt certificate password: %w", err)
	}
	
	// 8.2. Construir ruta al certificado P12
	certPath := s.storage.CertificatePath(invoice.Company.NIT, cert.Name)
	
	// 8.3. Convertir P12 a PEM de cliente (solo cert de usuario, sin CA certs)
	clientPemPath, err := signature.ConvertP12ToClientPEM(certPath, decryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to convert certificate to client PEM: %w", err)
	}

	// 9. Crear cliente SOAP
	config := &types.Config{
		Environment: environment,
		Certificate: clientPemPath,
		PrivateKey:  clientPemPath,
	}
	client, err := soap.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating SOAP client: %w", err)
	}
	
	// 10. Preparar request para DIAN según ambiente
	var response *types.Response
	
	if environment == types.Habilitacion && invoice.Software.TestSetID != nil && *invoice.Software.TestSetID != "" {
		testSetRequest := &types.SendTestSetAsyncRequest{
			FileName:    fmt.Sprintf("FES-%s.zip", invoice.Number),
			ContentFile: zipBase64,
			TestSetId:   *invoice.Software.TestSetID,
		}
		testSetResponse, err := client.SendTestSetAsync(testSetRequest)
		if err != nil {
			return fmt.Errorf("error sending to DIAN (TestSet): %w", err)
		}
		response = &testSetResponse.Response
	} else {
		syncRequest := &types.SendBillSyncRequest{
			FileName:    fmt.Sprintf("FES-%s.zip", invoice.Number),
			ContentFile: zipBase64,
		}
		syncResponse, err := client.SendBillSync(syncRequest)
		if err != nil {
			return fmt.Errorf("error sending to DIAN: %w", err)
		}
		response = &syncResponse.Response
	}

	// 11. Guardar XmlDocumentKey (TrackId) si existe (para consultas posteriores con GetStatus)
	if response.XmlDocumentKey != "" {
		if err := s.invoiceRepo.UpdateTrackId(id, response.XmlDocumentKey); err != nil {
			fmt.Printf("Warning: Failed to save TrackId: %v\n", err)
		}
	}

	// 12. Guardar ApplicationResponse si existe (antes de validar)
	if response.XmlBase64Bytes != "" {
		appResponseXML, err := base64.StdEncoding.DecodeString(response.XmlBase64Bytes)
		if err == nil {
			appResponsePath := s.storage.InvoiceApplicationResponsePath(invoice.Company.NIT, invoice.Number)
			if err := os.WriteFile(appResponsePath, appResponseXML, 0644); err != nil {
				fmt.Printf("Warning: Failed to save ApplicationResponse: %v\n", err)
			}
		}
	}

	// 13. Validar respuesta
	if !response.IsValid {
		s.invoiceRepo.UpdateDIANStatus(id, "rejected", response.StatusMessage, response.StatusCode, response.StatusDescription)
		message := response.StatusDescription
		if message == "" {
			message = response.StatusMessage
		}
		return fmt.Errorf("DIAN_REJECTION: StatusCode=%s, Message=%s", response.StatusCode, message)
	}

	// 14. Actualizar BD con éxito
	if err := s.invoiceRepo.UpdateStatus(id, "sent"); err != nil {
		return err
	}

	if err := s.invoiceRepo.UpdateDIANStatus(id, "accepted", response.StatusMessage, response.StatusCode, response.StatusDescription); err != nil {
		return err
	}

	return nil
}
