package validator

import "apidian-go/internal/domain"

// ValidateCreateCompany valida la creación de una empresa según reglas DIAN
func ValidateCreateCompany(req *domain.CreateCompanyRequest) error {
	// NIT con dígito verificador
	if err := ValidateNIT(req.NIT, req.DV); err != nil {
		return err
	}

	// Nombre comercial
	if err := IsRequired(req.Name, "name"); err != nil {
		return err
	}
	if err := IsValidLength(req.Name, 3, 255, "name"); err != nil {
		return err
	}

	// Razón social
	if err := IsRequired(req.RegistrationName, "registration_name"); err != nil {
		return err
	}
	if err := IsValidLength(req.RegistrationName, 3, 255, "registration_name"); err != nil {
		return err
	}

	// Códigos CIIU
	if err := ValidateIndustryCodes(req.IndustryCodes); err != nil {
		return err
	}

	// Dirección
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

	// Website (opcional)
	if req.Website != nil && *req.Website != "" {
		if !IsValidURL(*req.Website) {
			return NewError("website", "formato de URL inválido. Debe iniciar con http:// o https://")
		}
	}

	// Validar IDs requeridos (existen en DB, pero validamos que no sean 0)
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

// ValidateUpdateCompany valida la actualización de una empresa
func ValidateUpdateCompany(req *domain.UpdateCompanyRequest) error {
	// Nombre (opcional en update, pero si se envía debe ser válido)
	if req.Name != nil && *req.Name != "" {
		if err := IsValidLength(*req.Name, 3, 255, "name"); err != nil {
			return err
		}
	}

	// Razón social (opcional en update)
	if req.RegistrationName != nil && *req.RegistrationName != "" {
		if err := IsValidLength(*req.RegistrationName, 3, 255, "registration_name"); err != nil {
			return err
		}
	}

	// Códigos CIIU
	if req.IndustryCodes != nil {
		if err := ValidateIndustryCodes(req.IndustryCodes); err != nil {
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

	// Website
	if req.Website != nil && *req.Website != "" {
		if !IsValidURL(*req.Website) {
			return NewError("website", "formato de URL inválido. Debe iniciar con http:// o https://")
		}
	}

	return nil
}
