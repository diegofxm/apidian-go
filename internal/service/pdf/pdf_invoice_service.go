package pdf

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"fmt"
	"os"
)

// PDFInvoiceService maneja la generación de PDFs de facturas
type PDFInvoiceService struct {
	storage  *config.StorageConfig
	template InvoiceTemplate
}

// NewPDFInvoiceService crea una nueva instancia del servicio de PDFs
func NewPDFInvoiceService(storage *config.StorageConfig) *PDFInvoiceService {
	return &PDFInvoiceService{
		storage:  storage,
		template: NewDefaultTemplate(),
	}
}

// GenerateInvoicePDF genera el PDF de una factura y retorna los bytes
func (s *PDFInvoiceService) GenerateInvoicePDF(invoice *domain.Invoice) ([]byte, error) {
	logoPath := s.getLogoPath(invoice.Company)
	
	m := s.template.BuildPDF(invoice, logoPath)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("error al generar documento: %w", err)
	}

	pdfBytes := document.GetBytes()
	return pdfBytes, nil
}

// getLogoPath obtiene la ruta del logo de la empresa o el logo por defecto
func (s *PDFInvoiceService) getLogoPath(company *domain.CompanyDetail) string {
	if company != nil && company.LogoPath != nil && *company.LogoPath != "" {
		profilePath := s.storage.CompanyProfilePath(company.NIT)
		
		// Buscar logo con diferentes extensiones
		extensions := []string{".png", ".jpg", ".jpeg"}
		for _, ext := range extensions {
			logoPath := fmt.Sprintf("%s/logo%s", profilePath, ext)
			if _, err := os.Stat(logoPath); err == nil {
				return logoPath
			}
		}
	}
	
	// Si no existe logo personalizado, usar el logo por defecto
	return s.storage.DefaultLogoPath()
}
