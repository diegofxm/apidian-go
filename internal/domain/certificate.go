package domain

import "time"

// Certificate represents a digital certificate (PKCS12) for signing electronic documents
type Certificate struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name"`
	Password  string    `json:"-"` // Never expose in JSON
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCertificateRequest represents the request to upload a new certificate
type CreateCertificateRequest struct {
	CompanyID   int64  `json:"company_id" validate:"required"`
	Certificate string `json:"certificate" validate:"required"` // Base64 encoded .p12 file
	Password    string `json:"password" validate:"required"`
}

// CertificateResponse represents the public certificate information
type CertificateResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
