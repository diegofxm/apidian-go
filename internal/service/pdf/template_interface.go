package pdf

import (
	"apidian-go/internal/domain"

	"github.com/johnfercher/maroto/v2/pkg/core"
)

// InvoiceTemplate define la interfaz para templates de facturas
type InvoiceTemplate interface {
	// BuildPDF construye el documento PDF completo para una factura
	BuildPDF(invoice *domain.Invoice, logoPath string) core.Maroto
}
