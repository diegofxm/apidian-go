package domain

import "time"

// Resolution representa una resolución de numeración DIAN
type Resolution struct {
	ID             int64     `json:"id"`
	CompanyID      int64     `json:"company_id"`
	TypeDocumentID int       `json:"type_document_id"`
	Prefix         string    `json:"prefix"`
	Resolution     string    `json:"resolution"`
	TechnicalKey   *string   `json:"technical_key,omitempty"`
	FromNumber     int64     `json:"from_number"`
	ToNumber       int64     `json:"to_number"`
	CurrentNumber  int64     `json:"current_number"`
	DateFrom       time.Time `json:"date_from"`
	DateTo         time.Time `json:"date_to"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateResolutionRequest representa la solicitud para crear una resolución
type CreateResolutionRequest struct {
	CompanyID      int64   `json:"company_id" validate:"required"`
	TypeDocumentID int     `json:"type_document_id" validate:"required"`
	Prefix         string  `json:"prefix" validate:"required"`
	Resolution     string  `json:"resolution" validate:"required"`
	TechnicalKey   *string `json:"technical_key,omitempty"`
	FromNumber     int64   `json:"from_number" validate:"required"`
	ToNumber       int64   `json:"to_number" validate:"required"`
	DateFrom       string  `json:"date_from" validate:"required"` // Format: YYYY-MM-DD
	DateTo         string  `json:"date_to" validate:"required"`   // Format: YYYY-MM-DD
}

// ResolutionListResponse representa la respuesta paginada de resoluciones
type ResolutionListResponse struct {
	Resolutions []Resolution `json:"resolutions"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
}
