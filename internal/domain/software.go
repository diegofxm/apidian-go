package domain

import "time"

// Software representa la configuración del software DIAN por empresa
type Software struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	Identifier  string    `json:"identifier"`
	Pin         string    `json:"pin"`
	Environment string    `json:"environment"` // "1" = Producción, "2" = Habilitación
	TestSetID   *string   `json:"test_set_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateSoftwareRequest representa la solicitud para crear un software
type CreateSoftwareRequest struct {
	CompanyID   int64   `json:"company_id" validate:"required"`
	Identifier  string  `json:"identifier" validate:"required"`
	Pin         string  `json:"pin" validate:"required"`
	Environment string  `json:"environment" validate:"required"`
	TestSetID   *string `json:"test_set_id,omitempty"`
}

// UpdateSoftwareRequest representa la solicitud para actualizar un software
type UpdateSoftwareRequest struct {
	Identifier  *string `json:"identifier,omitempty"`
	Pin         *string `json:"pin,omitempty"`
	Environment *string `json:"environment,omitempty"`
	TestSetID   *string `json:"test_set_id,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
