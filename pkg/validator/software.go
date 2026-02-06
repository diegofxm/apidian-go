package validator

import (
	"apidian-go/internal/domain"
	"strings"
)

// ValidateCreateSoftware valida la solicitud de creación de software
func ValidateCreateSoftware(req *domain.CreateSoftwareRequest) error {
	// CompanyID requerido
	if req.CompanyID == 0 {
		return NewError("company_id", "es requerido")
	}

	// Identifier requerido (UUID)
	if req.Identifier == "" {
		return NewError("identifier", "es requerido")
	}
	if len(req.Identifier) < 10 || len(req.Identifier) > 255 {
		return NewError("identifier", "debe tener entre 10 y 255 caracteres")
	}

	// Pin requerido (5 dígitos)
	if req.Pin == "" {
		return NewError("pin", "es requerido")
	}
	if !IsNumeric(req.Pin) {
		return NewError("pin", "debe contener solo números")
	}
	if len(req.Pin) != 5 {
		return NewError("pin", "debe tener exactamente 5 dígitos")
	}

	// Environment requerido ("1" o "2")
	if req.Environment == "" {
		return NewError("environment", "es requerido")
	}
	if req.Environment != "1" && req.Environment != "2" {
		return NewError("environment", "debe ser '1' (Producción) o '2' (Habilitación)")
	}

	// TestSetID opcional pero si existe debe ser UUID
	if req.TestSetID != nil && *req.TestSetID != "" {
		if len(*req.TestSetID) < 10 {
			return NewError("test_set_id", "formato inválido")
		}
	}

	return nil
}

// ValidateUpdateSoftware valida la solicitud de actualización de software
func ValidateUpdateSoftware(req *domain.UpdateSoftwareRequest) error {
	// Al menos un campo debe estar presente
	if req.Identifier == nil && req.Pin == nil && req.Environment == nil && 
	   req.TestSetID == nil && req.IsActive == nil {
		return NewError("request", "debe proporcionar al menos un campo para actualizar")
	}

	// Validar Identifier si está presente
	if req.Identifier != nil {
		if *req.Identifier == "" {
			return NewError("identifier", "no puede estar vacío")
		}
		if len(*req.Identifier) < 10 || len(*req.Identifier) > 255 {
			return NewError("identifier", "debe tener entre 10 y 255 caracteres")
		}
	}

	// Validar Pin si está presente
	if req.Pin != nil {
		if *req.Pin == "" {
			return NewError("pin", "no puede estar vacío")
		}
		if !IsNumeric(*req.Pin) {
			return NewError("pin", "debe contener solo números")
		}
		if len(*req.Pin) != 5 {
			return NewError("pin", "debe tener exactamente 5 dígitos")
		}
	}

	// Validar Environment si está presente
	if req.Environment != nil {
		env := strings.TrimSpace(*req.Environment)
		if env != "1" && env != "2" {
			return NewError("environment", "debe ser '1' (Producción) o '2' (Habilitación)")
		}
	}

	// Validar TestSetID si está presente
	if req.TestSetID != nil && *req.TestSetID != "" {
		if len(*req.TestSetID) < 10 {
			return NewError("test_set_id", "formato inválido")
		}
	}

	return nil
}
