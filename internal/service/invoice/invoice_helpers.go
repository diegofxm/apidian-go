package invoice

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"apidian-go/internal/domain"
	"apidian-go/pkg/crypto"
	attachedpkg "github.com/diegofxm/ubl21-dian/documents/attached"
	"github.com/diegofxm/ubl21-dian/signature"
)

// updateInvoiceMetadata actualiza UUID y XML path en la BD
func (s *InvoiceService) updateInvoiceMetadata(id int64, uuid, xmlPath string) error {
	// Actualizar UUID (CUFE)
	if err := s.invoiceRepo.UpdateUUID(id, uuid); err != nil {
		return err
	}
	
	// Actualizar XML path
	if err := s.invoiceRepo.UpdateXMLPath(id, xmlPath); err != nil {
		return err
	}
	
	return nil
}

// createZipFile crea un archivo ZIP con un solo archivo XML
func (s *InvoiceService) createZipFile(zipPath, xmlFileName string, xmlContent []byte) error {
	// Crear archivo ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error creating zip file: %w", err)
	}
	defer zipFile.Close()

	// Crear writer ZIP
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Agregar XML al ZIP
	xmlWriter, err := zipWriter.Create(xmlFileName)
	if err != nil {
		return fmt.Errorf("error creating entry in zip: %w", err)
	}

	if _, err := xmlWriter.Write(xmlContent); err != nil {
		return fmt.Errorf("error writing to zip: %w", err)
	}

	return nil
}

// GeneratePDF genera el PDF de una factura firmada
func (s *InvoiceService) GeneratePDF(id int64, userID int64) error {
	// 1. Obtener factura completa
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 2. Validar que esté firmada
	if invoice.Status != "signed" && invoice.Status != "sent" {
		return fmt.Errorf("invoice must be signed to generate PDF")
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

	// 5. Generar QR Code (con CUFE)
	// TODO: Generar QR Code cuando sea necesario
	_ = invoice.UUID // Placeholder

	// 6. Generar PDF usando ubl21-dian/pdf
	// TODO: Implementar cuando el módulo pdf esté disponible en ubl21-dian
	// Por ahora, solo guardamos la ruta del QR
	
	// PDF se genera on-demand, no se guarda en disco
	// Solo guardamos una referencia para futuro uso
	pdfPath := fmt.Sprintf("/api/v1/invoices/%d/pdf", id)

	// 7. Actualizar BD con pdf_path
	if err := s.invoiceRepo.UpdatePDFPath(id, pdfPath); err != nil {
		return fmt.Errorf("error updating PDF path: %w", err)
	}
	
	// TODO: Actualizar qr_code_url cuando se tenga el método en repository

	// Placeholder: En producción, aquí se generaría el PDF real
	// usando una librería como github.com/jung-kurt/gofpdf o similar
	_ = xmlSigned // Evitar warning de variable no usada

	return nil
}

// GenerateAttachedDocument genera el AttachedDocument para el cliente
func (s *InvoiceService) GenerateAttachedDocument(id int64, userID int64) error {
	// 1. Obtener factura completa
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 2. Validar que esté enviada a DIAN
	if invoice.Status != "sent" {
		return fmt.Errorf("invoice must be sent to DIAN to generate AttachedDocument")
	}

	// 3. Validar que tenga XML firmado
	if invoice.XMLPath == nil || *invoice.XMLPath == "" {
		return fmt.Errorf("invoice does not have signed XML")
	}

	// Crear directorio de factura si no existe
	invoiceDir := s.storage.InvoicePath(invoice.Company.NIT, invoice.Number)
	if err := os.MkdirAll(invoiceDir, 0755); err != nil {
		return fmt.Errorf("error creating invoice directory: %w", err)
	}
	
	// 4. Leer Invoice firmado
	invoiceXML, err := os.ReadFile(*invoice.XMLPath)
	if err != nil {
		return fmt.Errorf("error reading signed invoice XML: %w", err)
	}

	// 5. Leer ApplicationResponse
	appResponsePath := s.storage.InvoiceApplicationResponsePath(invoice.Company.NIT, invoice.Number)
	appResponseXML, err := os.ReadFile(appResponsePath)
	if err != nil {
		return fmt.Errorf("error reading ApplicationResponse XML: %w", err)
	}

	// 6. Construir AttachedDocument usando el nuevo builder
	sender := attachedpkg.PartyData{
		RegistrationName: invoice.Company.RegistrationName,
		CompanyID:        invoice.Company.NIT,
		SchemeID:         getStringValue(invoice.Company.DV),
		SchemeName:       invoice.Company.DocumentTypeCode,
		TaxLevelCode:     invoice.Company.TaxLevelCode,
		TaxSchemeID:      invoice.Company.TaxSchemeID,
		TaxSchemeName:    invoice.Company.TaxSchemeName,
	}

	receiver := attachedpkg.PartyData{
		RegistrationName: invoice.Customer.Name,
		CompanyID:        invoice.Customer.IdentificationNumber,
		SchemeID:         getStringValue(invoice.Customer.DV),
		SchemeName:       invoice.Customer.DocumentTypeCode,
		TaxLevelCode:     invoice.Customer.TaxLevelCode,
		TaxSchemeID:      invoice.Customer.TaxSchemeID,
		TaxSchemeName:    invoice.Customer.TaxSchemeName,
	}

	// Extraer fecha y hora de validación del ApplicationResponse (simplificado)
	validationTime := time.Now()

	appResponse := attachedpkg.ApplicationResponseData{
		InvoiceID:            invoice.Number,
		CUFE:                 *invoice.UUID,
		IssueDate:            invoice.IssueDate,
		ResponseXML:          string(appResponseXML),
		ValidationResultCode: "02", // 02 = Validado por DIAN
		ValidationDate:       validationTime.Format("2006-01-02"),
		ValidationTime:       validationTime.Format("15:04:05-07:00"),
	}

	// Construir AttachedDocument
	builder := attachedpkg.NewBuilder().
		SetProfileExecutionID(invoice.Software.Environment).
		SetID(*invoice.UUID).
		SetIssueDate(time.Now()).
		SetParentDocumentID(invoice.Number).
		SetSender(sender).
		SetReceiver(receiver).
		SetSignedInvoiceXML(string(invoiceXML)).
		SetApplicationResponse(appResponse)

	// 7. Generar XML del AttachedDocument (sin firma)
	attachedXMLBytes, err := builder.ToXML()
	if err != nil {
		return fmt.Errorf("error generating AttachedDocument XML: %w", err)
	}

	// 8. Guardar AttachedDocument sin firma (temporal, se eliminará después de firmar)
	attachedPath := filepath.Join(invoiceDir, fmt.Sprintf("AttachedDocument-%s.xml", invoice.Number))
	if err := os.WriteFile(attachedPath, attachedXMLBytes, 0644); err != nil {
		return fmt.Errorf("error saving AttachedDocument XML: %w", err)
	}

	// 9. Obtener certificado y firmar AttachedDocument
	cert, err := s.certificateRepo.GetByCompanyID(invoice.CompanyID)
	if err != nil {
		return fmt.Errorf("no certificate found for company: %w", err)
	}
	
	// 9.1. Desencriptar contraseña del certificado
	decryptedPassword, err := crypto.DecryptPassword(cert.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt certificate password: %w", err)
	}
	
	certPath := s.storage.CertificatePath(invoice.Company.NIT, cert.Name)
	
	// Usar fallback automático a PEM si P12 falla
	signer, err := signature.NewSignerFromP12WithFallback(certPath, decryptedPassword)
	if err != nil {
		return fmt.Errorf("error loading certificate: %w", err)
	}

	// NO envolver - firmar el XML tal como se genera
	// WrapInvoiceWithFixedNamespaces destruye el formato y cambia el hash
	attachedXMLWrapped := attachedXMLBytes

	attachedXMLSigned, err := signer.SignXML(attachedXMLWrapped)
	if err != nil {
		return fmt.Errorf("error signing AttachedDocument: %w", err)
	}

	// 10. Guardar AttachedDocument firmado
	attachedSignedPath := filepath.Join(invoiceDir, fmt.Sprintf("ad%s.xml", invoice.Number))
	if err := os.WriteFile(attachedSignedPath, attachedXMLSigned, 0644); err != nil {
		return fmt.Errorf("error saving signed AttachedDocument: %w", err)
	}

	// 11. Crear ZIP final con todos los documentos
	zipPath := filepath.Join(invoiceDir, fmt.Sprintf("ad%s.zip", invoice.Number))
	if err := s.createAttachedDocumentZip(zipPath, invoice.Number, invoiceXML, appResponseXML, attachedXMLSigned); err != nil {
		return fmt.Errorf("error creating final ZIP: %w", err)
	}

	// 12. Actualizar BD con zip_path
	if err := s.invoiceRepo.UpdateZIPPath(id, zipPath); err != nil {
		return fmt.Errorf("error updating ZIP path: %w", err)
	}

	return nil
}

// createAttachedDocumentZip crea el ZIP final para el cliente con todos los documentos
func (s *InvoiceService) createAttachedDocumentZip(zipPath, invoiceNumber string, invoiceXML, appResponseXML, attachedXML []byte) error {
	// Crear archivo ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error creating zip file: %w", err)
	}
	defer zipFile.Close()

	// Crear writer ZIP
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Agregar Invoice firmado
	if err := addFileToZip(zipWriter, fmt.Sprintf("FES-%s.xml", invoiceNumber), invoiceXML); err != nil {
		return err
	}

	// Agregar ApplicationResponse
	if err := addFileToZip(zipWriter, fmt.Sprintf("ApplicationResponse-%s.xml", invoiceNumber), appResponseXML); err != nil {
		return err
	}

	// Agregar AttachedDocument firmado
	if err := addFileToZip(zipWriter, fmt.Sprintf("ad%s.xml", invoiceNumber), attachedXML); err != nil {
		return err
	}

	return nil
}

// addFileToZip agrega un archivo al ZIP
func addFileToZip(zipWriter *zip.Writer, fileName string, content []byte) error {
	writer, err := zipWriter.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating entry in zip: %w", err)
	}

	if _, err := writer.Write(content); err != nil {
		return fmt.Errorf("error writing to zip: %w", err)
	}

	return nil
}

// DownloadInvoiceZip retorna el path del ZIP final para descarga
func (s *InvoiceService) DownloadInvoiceZip(id int64, userID int64) (string, error) {
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return "", err
	}

	if invoice.ZipPath == nil || *invoice.ZipPath == "" {
		return "", fmt.Errorf("invoice does not have ZIP file, generate AttachedDocument first")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(*invoice.ZipPath); os.IsNotExist(err) {
		return "", fmt.Errorf("ZIP file not found on disk")
	}

	return *invoice.ZipPath, nil
}

// GetInvoiceXML retorna el XML firmado de una factura
func (s *InvoiceService) GetInvoiceXML(id int64, userID int64) ([]byte, error) {
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	if invoice.XMLPath == nil || *invoice.XMLPath == "" {
		return nil, fmt.Errorf("invoice does not have signed XML")
	}

	xmlContent, err := os.ReadFile(*invoice.XMLPath)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %w", err)
	}

	return xmlContent, nil
}

// GetInvoicePDF retorna el PDF de una factura
func (s *InvoiceService) GetInvoicePDF(id int64, userID int64) (string, error) {
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return "", err
	}

	if invoice.PDFPath == nil || *invoice.PDFPath == "" {
		return "", fmt.Errorf("invoice does not have PDF, generate it first")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(*invoice.PDFPath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF file not found on disk")
	}

	return *invoice.PDFPath, nil
}

// CopyFile copia un archivo de src a dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// getStringValue retorna el valor de un puntero a string o vacío
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getEnvironmentStr(software *domain.SoftwareDetail) string {
	if software.Environment == "2" || software.Environment == "Habilitacion" {
		return "2"
	}
	return "1"
}

func getInvoiceNote(inv *domain.Invoice) string {
	if inv.Notes != nil {
		return *inv.Notes
	}
	return ""
}

func getProviderSchemeID(typeOrgCode string, dv *string) string {
	// Si es Persona Natural (cédula), schemeID = "1"
	if typeOrgCode == "2" {
		return "1"
	}
	// Si es Persona Jurídica (NIT), schemeID = DV del NIT
	if dv != nil && *dv != "" {
		return *dv
	}
	return "0" // Fallback si no hay DV
}

func getProviderSchemeName(typeOrgCode string) string {
	if typeOrgCode == "2" {
		return "13" // Persona Natural
	}
	return "31" // Persona Jurídica (NIT)
}

func getDocumentTypeSchemeID(typeDocCode string) string {
	// Para AccountingSupplierParty y AccountingCustomerParty
	// schemeID según tipo de documento
	switch typeDocCode {
	case "13": // Cédula de Ciudadanía
		return "1"
	case "31": // NIT
		return "4"
	case "22": // Cédula de Extranjería
		return "2"
	case "41": // Pasaporte
		return "3"
	default:
		return "1"
	}
}

func getDocumentTypeSchemeName(typeDocCode string) string {
	// schemeName es el código del tipo de documento
	return typeDocCode
}

func getPaymentMethodCode(paymentMethodID *int64) string {
	if paymentMethodID == nil || *paymentMethodID == 0 {
		return "10"
	}
	switch *paymentMethodID {
	case 1:
		return "10"
	case 2:
		return "48"
	case 3:
		return "49"
	case 4:
		return "30"
	default:
		return "10"
	}
}

func formatIndustryCodes(codes string) string {
	if codes == "" {
		return ""
	}
	// Remover llaves { }
	codes = strings.TrimPrefix(codes, "{")
	codes = strings.TrimSuffix(codes, "}")
	// Reemplazar comas por punto y coma
	codes = strings.ReplaceAll(codes, ",", ";")
	return codes
}
