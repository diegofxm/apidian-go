package validator

import (
	"apidian-go/internal/domain"
	"fmt"
)

// ValidateCreateInvoice valida la solicitud de creación de factura
func ValidateCreateInvoice(req *domain.CreateInvoiceRequest) error {
	if req.CompanyID <= 0 {
		return fmt.Errorf("company_id es requerido")
	}

	if req.CustomerID <= 0 {
		return fmt.Errorf("customer_id es requerido")
	}

	if req.ResolutionID <= 0 {
		return fmt.Errorf("resolution_id es requerido")
	}

	if req.IssueDate == "" {
		return fmt.Errorf("issue_date es requerido")
	}

	if req.CurrencyCodeID <= 0 {
		return fmt.Errorf("currency_code_id es requerido")
	}

	if len(req.Lines) == 0 {
		return fmt.Errorf("debe incluir al menos una línea de factura")
	}

	// Validar cada línea
	for i, line := range req.Lines {
		if err := ValidateCreateInvoiceLine(&line, i+1); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCreateInvoiceLine valida una línea de factura
func ValidateCreateInvoiceLine(line *domain.CreateInvoiceLineRequest, lineNumber int) error {
	if line.ProductID <= 0 {
		return fmt.Errorf("product_id es requerido en la línea %d", lineNumber)
	}

	if line.Quantity <= 0 {
		return fmt.Errorf("quantity debe ser mayor a 0 en la línea %d", lineNumber)
	}

	// Validar unit_price solo si se proporciona
	if line.UnitPrice != nil && *line.UnitPrice < 0 {
		return fmt.Errorf("unit_price no puede ser negativo en la línea %d", lineNumber)
	}

	// Validar tax_rate solo si se proporciona
	if line.TaxRate != nil && (*line.TaxRate < 0 || *line.TaxRate > 100) {
		return fmt.Errorf("tax_rate debe estar entre 0 y 100 en la línea %d", lineNumber)
	}

	return nil
}

// ValidateUpdateInvoice valida la solicitud de actualización de factura
func ValidateUpdateInvoice(req *domain.UpdateInvoiceRequest) error {
	// No hay validaciones estrictas para update ya que todos los campos son opcionales
	return nil
}
