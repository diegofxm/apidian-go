package validator

import "apidian-go/internal/domain"

// ValidateCreateCustomer valida la creación de un cliente según reglas DIAN
func ValidateCreateCustomer(req *domain.CreateCustomerRequest) error {
	// Número de identificación (alpha_num, 1-15 caracteres)
	if err := ValidateIdentification(req.IdentificationNumber, "identification_number"); err != nil {
		return err
	}

	// Validar DV si es NIT (tipo documento 31)
	if req.DocumentTypeID == 31 {
		if err := ValidateNIT(req.IdentificationNumber, req.DV); err != nil {
			return err
		}
	}

	// Nombre requerido
	if err := IsRequired(req.Name, "name"); err != nil {
		return err
	}
	if err := IsValidLength(req.Name, 3, 255, "name"); err != nil {
		return err
	}

	// Dirección requerida
	if err := IsRequired(req.AddressLine, "address_line"); err != nil {
		return err
	}
	if err := IsValidLength(req.AddressLine, 5, 255, "address_line"); err != nil {
		return err
	}

	// Código postal (opcional)
	if req.PostalZone != nil && *req.PostalZone != "" {
		if err := ValidatePostalCode(*req.PostalZone); err != nil {
			return err
		}
	}

	// Teléfono (opcional)
	if req.Phone != nil && *req.Phone != "" {
		if !IsValidPhone(*req.Phone) {
			return NewError("phone", "formato de teléfono inválido. Use formato: 3001234567 o 6011234567")
		}
	}

	// Email (opcional)
	if req.Email != nil {
		if err := ValidateEmail(*req.Email, "email"); err != nil {
			return err
		}
	}

	// Validar IDs requeridos
	if req.CompanyID == 0 {
		return NewError("company_id", "es requerido")
	}
	if req.DocumentTypeID == 0 {
		return NewError("document_type_id", "es requerido")
	}
	if req.TaxLevelCodeID == 0 {
		return NewError("tax_level_code_id", "es requerido")
	}
	if req.TypeOrganizationID == 0 {
		return NewError("type_organization_id", "es requerido")
	}
	if req.TypeRegimeID == 0 {
		return NewError("type_regime_id", "es requerido")
	}
	if req.CountryID == 0 {
		return NewError("country_id", "es requerido")
	}
	if req.DepartmentID == 0 {
		return NewError("department_id", "es requerido")
	}
	if req.MunicipalityID == 0 {
		return NewError("municipality_id", "es requerido")
	}

	return nil
}

// ValidateUpdateCustomer valida la actualización de un cliente
func ValidateUpdateCustomer(req *domain.UpdateCustomerRequest) error {
	// Nombre (opcional en update, pero si se envía debe ser válido)
	if req.Name != nil && *req.Name != "" {
		if err := IsValidLength(*req.Name, 3, 255, "name"); err != nil {
			return err
		}
	}

	// Dirección
	if req.AddressLine != nil && *req.AddressLine != "" {
		if err := IsValidLength(*req.AddressLine, 5, 255, "address_line"); err != nil {
			return err
		}
	}

	// Código postal
	if req.PostalZone != nil && *req.PostalZone != "" {
		if err := ValidatePostalCode(*req.PostalZone); err != nil {
			return err
		}
	}

	// Teléfono
	if req.Phone != nil && *req.Phone != "" {
		if !IsValidPhone(*req.Phone) {
			return NewError("phone", "formato de teléfono inválido. Use formato: 3001234567 o 6011234567")
		}
	}

	// Email
	if req.Email != nil && *req.Email != "" {
		if err := ValidateEmail(*req.Email, "email"); err != nil {
			return err
		}
	}

	return nil
}
