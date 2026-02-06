package domain

import (
	"time"
)

// Company representa una empresa emisora de facturas electrónicas
type Company struct {
	ID                  int64    `json:"id"`
	UserID              int64    `json:"user_id"`
	DocumentTypeID      int      `json:"document_type_id"`
	NIT                 string   `json:"nit"`
	DV                  *string  `json:"dv,omitempty"`
	Name                string   `json:"name"`
	TradeName           *string  `json:"trade_name,omitempty"`
	RegistrationName    string   `json:"registration_name"`
	TaxLevelCodeID      int      `json:"tax_level_code_id"`
	TaxTypeID           *int     `json:"tax_type_id,omitempty"`
	TypeOrganizationID  int      `json:"type_organization_id"`
	TypeRegimeID        int      `json:"type_regime_id"`
	IndustryCodes       []string `json:"industry_codes,omitempty"`
	CountryID           int      `json:"country_id"`
	DepartmentID        int      `json:"department_id"`
	MunicipalityID      int      `json:"municipality_id"`
	AddressLine         string   `json:"address_line"`
	PostalZone          *string  `json:"postal_zone,omitempty"`
	Phone               *string  `json:"phone,omitempty"`
	Email               *string  `json:"email,omitempty"`
	Website             *string  `json:"website,omitempty"`
	LogoPath            *string  `json:"logo_path"`
	IsActive            bool     `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateCompanyRequest representa la solicitud para crear una empresa
type CreateCompanyRequest struct {
	DocumentTypeID     int      `json:"document_type_id" validate:"required"`
	NIT                string   `json:"nit" validate:"required,min=6,max=20"`
	DV                 *string  `json:"dv,omitempty"`
	Name               string   `json:"name" validate:"required,min=3,max=255"`
	TradeName          *string  `json:"trade_name,omitempty"`
	RegistrationName   string   `json:"registration_name" validate:"required,min=3,max=255"`
	TaxLevelCodeID     int      `json:"tax_level_code_id" validate:"required"`
	TaxTypeID          *int     `json:"tax_type_id,omitempty"`
	TypeOrganizationID int      `json:"type_organization_id" validate:"required"`
	TypeRegimeID       int      `json:"type_regime_id" validate:"required"`
	IndustryCodes      []string `json:"industry_codes,omitempty" validate:"max=4"`
	CountryID          int      `json:"country_id" validate:"required"`
	DepartmentID       int      `json:"department_id" validate:"required"`
	MunicipalityID     int      `json:"municipality_id" validate:"required"`
	AddressLine        string   `json:"address_line" validate:"required,min=5,max=255"`
	PostalZone         *string  `json:"postal_zone,omitempty"`
	Phone              *string  `json:"phone,omitempty"`
	Email              *string  `json:"email,omitempty" validate:"omitempty,email"`
	Website            *string  `json:"website,omitempty"`
	LogoPath           *string  `json:"logo_path,omitempty"`
}

// UpdateCompanyRequest representa la solicitud para actualizar una empresa
type UpdateCompanyRequest struct {
	Name               *string  `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	TradeName          *string  `json:"trade_name,omitempty"`
	RegistrationName   *string  `json:"registration_name,omitempty" validate:"omitempty,min=3,max=255"`
	TaxLevelCodeID     *int     `json:"tax_level_code_id,omitempty"`
	TaxTypeID          *int     `json:"tax_type_id,omitempty"`
	TypeOrganizationID *int     `json:"type_organization_id,omitempty"`
	TypeRegimeID       *int     `json:"type_regime_id,omitempty"`
	IndustryCodes      []string `json:"industry_codes,omitempty" validate:"max=4"`
	DepartmentID       *int     `json:"department_id,omitempty"`
	MunicipalityID     *int     `json:"municipality_id,omitempty"`
	AddressLine        *string  `json:"address_line,omitempty" validate:"omitempty,min=5,max=255"`
	PostalZone         *string  `json:"postal_zone,omitempty"`
	Phone              *string  `json:"phone,omitempty"`
	Email              *string  `json:"email,omitempty" validate:"omitempty,email"`
	Website            *string  `json:"website,omitempty"`
	LogoPath           *string  `json:"logo_path,omitempty"`
	IsActive           *bool    `json:"is_active,omitempty"`
}

// CompanyListResponse representa la respuesta con lista de empresas
type CompanyListResponse struct {
	Companies []Company `json:"companies"`
	Total     int       `json:"total"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}
