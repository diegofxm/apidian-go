package validator

import (
	"apidian-go/internal/domain"
	"time"
)

// ValidateCreateResolution valida la solicitud de creación de resolución
func ValidateCreateResolution(req *domain.CreateResolutionRequest) error {
	// CompanyID requerido
	if req.CompanyID == 0 {
		return NewError("company_id", "es requerido")
	}

	// TypeDocumentID requerido
	if req.TypeDocumentID == 0 {
		return NewError("type_document_id", "es requerido")
	}

	// Prefix requerido (máximo 10 caracteres)
	if req.Prefix == "" {
		return NewError("prefix", "es requerido")
	}
	if len(req.Prefix) > 10 {
		return NewError("prefix", "debe tener máximo 10 caracteres")
	}

	// Resolution requerido (número de resolución DIAN)
	if req.Resolution == "" {
		return NewError("resolution", "es requerido")
	}
	if len(req.Resolution) > 50 {
		return NewError("resolution", "debe tener máximo 50 caracteres")
	}

	// FromNumber y ToNumber requeridos
	if req.FromNumber <= 0 {
		return NewError("from_number", "debe ser mayor a 0")
	}
	if req.ToNumber <= 0 {
		return NewError("to_number", "debe ser mayor a 0")
	}
	if req.FromNumber > req.ToNumber {
		return NewError("from_number", "debe ser menor o igual a to_number")
	}

	// Validar fechas
	if req.DateFrom == "" {
		return NewError("date_from", "es requerido")
	}
	if req.DateTo == "" {
		return NewError("date_to", "es requerido")
	}

	// Validar formato de fechas (YYYY-MM-DD)
	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		return NewError("date_from", "formato inválido, debe ser YYYY-MM-DD")
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		return NewError("date_to", "formato inválido, debe ser YYYY-MM-DD")
	}

	// Validar que date_from sea menor o igual a date_to
	if dateFrom.After(dateTo) {
		return NewError("date_from", "debe ser menor o igual a date_to")
	}

	return nil
}
