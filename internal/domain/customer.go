package domain

import "time"

// Customer representa un cliente (adquiriente) de facturas electrónicas
type Customer struct {
	ID                   int64     `json:"id"`
	CompanyID            int64     `json:"company_id"`
	DocumentTypeID       int       `json:"document_type_id"`
	IdentificationNumber string    `json:"identification_number"`
	DV                   *string   `json:"dv,omitempty"`
	Name                 string    `json:"name"`
	TradeName            *string   `json:"trade_name,omitempty"`
	TaxLevelCodeID       int       `json:"tax_level_code_id"`
	TaxTypeID            *int      `json:"tax_type_id,omitempty"`
	TypeOrganizationID   int       `json:"type_organization_id"`
	TypeRegimeID         int       `json:"type_regime_id"`
	CountryID            int       `json:"country_id"`
	DepartmentID         int       `json:"department_id"`
	MunicipalityID       int       `json:"municipality_id"`
	AddressLine          string    `json:"address_line"`
	PostalZone           *string   `json:"postal_zone,omitempty"`
	Phone                *string   `json:"phone,omitempty"`
	Email                *string   `json:"email,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateCustomerRequest representa la solicitud para crear un cliente
type CreateCustomerRequest struct {
	CompanyID            int64   `json:"company_id" validate:"required"`
	DocumentTypeID       int     `json:"document_type_id" validate:"required"`
	IdentificationNumber string  `json:"identification_number" validate:"required"`
	DV                   *string `json:"dv,omitempty"`
	Name                 string  `json:"name" validate:"required"`
	TradeName            *string `json:"trade_name,omitempty"`
	TaxLevelCodeID       int     `json:"tax_level_code_id" validate:"required"`
	TaxTypeID            *int    `json:"tax_type_id,omitempty"`
	TypeOrganizationID   int     `json:"type_organization_id" validate:"required"`
	TypeRegimeID         int     `json:"type_regime_id" validate:"required"`
	CountryID            int     `json:"country_id" validate:"required"`
	DepartmentID         int     `json:"department_id" validate:"required"`
	MunicipalityID       int     `json:"municipality_id" validate:"required"`
	AddressLine          string  `json:"address_line" validate:"required"`
	PostalZone           *string `json:"postal_zone,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	Email                *string `json:"email,omitempty"`
}

// UpdateCustomerRequest representa la solicitud para actualizar un cliente
type UpdateCustomerRequest struct {
	Name               *string `json:"name,omitempty"`
	TradeName          *string `json:"trade_name,omitempty"`
	TaxLevelCodeID     *int    `json:"tax_level_code_id,omitempty"`
	TaxTypeID          *int    `json:"tax_type_id,omitempty"`
	TypeOrganizationID *int    `json:"type_organization_id,omitempty"`
	TypeRegimeID       *int    `json:"type_regime_id,omitempty"`
	DepartmentID       *int    `json:"department_id,omitempty"`
	MunicipalityID     *int    `json:"municipality_id,omitempty"`
	AddressLine        *string `json:"address_line,omitempty"`
	PostalZone         *string `json:"postal_zone,omitempty"`
	Phone              *string `json:"phone,omitempty"`
	Email              *string `json:"email,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
}

// CustomerListResponse representa la respuesta paginada de clientes
type CustomerListResponse struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}
