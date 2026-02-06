package validator

import "apidian-go/internal/domain"

// ValidateCreateProduct valida la creación de un producto según reglas DIAN
func ValidateCreateProduct(req *domain.CreateProductRequest) error {
	// Código requerido (alfanumérico, 1-50 caracteres)
	if err := IsRequired(req.Code, "code"); err != nil {
		return err
	}
	if !IsAlphaNumeric(req.Code) {
		return NewError("code", "debe ser alfanumérico (letras y números)")
	}
	if err := IsValidLength(req.Code, 1, 50, "code"); err != nil {
		return err
	}

	// Nombre requerido
	if err := IsRequired(req.Name, "name"); err != nil {
		return err
	}
	if err := IsValidLength(req.Name, 3, 255, "name"); err != nil {
		return err
	}

	// Descripción (opcional)
	if req.Description != nil && *req.Description != "" {
		if err := IsValidLength(*req.Description, 3, 500, "description"); err != nil {
			return err
		}
	}

	// Precio (debe ser mayor a 0)
	if err := ValidateAmount(req.Price, "price"); err != nil {
		return err
	}

	// Tasa de impuesto (0-100%)
	if err := ValidatePercentage(req.TaxRate, "tax_rate"); err != nil {
		return err
	}

	// Validar IDs requeridos
	if req.CompanyID == 0 {
		return NewError("company_id", "es requerido")
	}
	if req.UnitCodeID == 0 {
		return NewError("unit_code_id", "es requerido")
	}
	if req.TaxTypeID == 0 {
		return NewError("tax_type_id", "es requerido")
	}

	return nil
}

// ValidateUpdateProduct valida la actualización de un producto
func ValidateUpdateProduct(req *domain.UpdateProductRequest) error {
	// Nombre (opcional en update, pero si se envía debe ser válido)
	if req.Name != nil && *req.Name != "" {
		if err := IsValidLength(*req.Name, 3, 255, "name"); err != nil {
			return err
		}
	}

	// Descripción
	if req.Description != nil && *req.Description != "" {
		if err := IsValidLength(*req.Description, 3, 500, "description"); err != nil {
			return err
		}
	}

	// Precio
	if req.Price != nil {
		if err := ValidateAmount(*req.Price, "price"); err != nil {
			return err
		}
	}

	// Tasa de impuesto
	if req.TaxRate != nil {
		if err := ValidatePercentage(*req.TaxRate, "tax_rate"); err != nil {
			return err
		}
	}

	return nil
}
