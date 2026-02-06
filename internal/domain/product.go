package domain

import "time"

// Product representa un producto o servicio
type Product struct {
	ID                       int64     `json:"id"`
	CompanyID                int64     `json:"company_id"`
	Code                     string    `json:"code"`
	Name                     string    `json:"name"`
	Description              *string   `json:"description,omitempty"`
	TypeItemIdentificationID *int      `json:"type_item_identification_id,omitempty"`
	StandardItemCode         *string   `json:"standard_item_code,omitempty"`
	UNSPSCCode               *string   `json:"unspsc_code,omitempty"`
	UnitCodeID               int       `json:"unit_code_id"`
	Price                    float64   `json:"price"`
	TaxTypeID                int       `json:"tax_type_id"`
	TaxRate                  float64   `json:"tax_rate"`
	BrandName                *string   `json:"brand_name,omitempty"`
	ModelName                *string   `json:"model_name,omitempty"`
	IsActive                 bool      `json:"is_active"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// CreateProductRequest representa la solicitud para crear un producto
type CreateProductRequest struct {
	CompanyID                int64   `json:"company_id" validate:"required"`
	Code                     string  `json:"code" validate:"required"`
	Name                     string  `json:"name" validate:"required"`
	Description              *string `json:"description,omitempty"`
	TypeItemIdentificationID *int    `json:"type_item_identification_id,omitempty"`
	StandardItemCode         *string `json:"standard_item_code,omitempty"`
	UNSPSCCode               *string `json:"unspsc_code,omitempty"`
	UnitCodeID               int     `json:"unit_code_id" validate:"required"`
	Price                    float64 `json:"price" validate:"required"`
	TaxTypeID                int     `json:"tax_type_id" validate:"required"`
	TaxRate                  float64 `json:"tax_rate"`
	BrandName                *string `json:"brand_name,omitempty"`
	ModelName                *string `json:"model_name,omitempty"`
}

// UpdateProductRequest representa la solicitud para actualizar un producto
type UpdateProductRequest struct {
	Name                     *string  `json:"name,omitempty"`
	Description              *string  `json:"description,omitempty"`
	TypeItemIdentificationID *int     `json:"type_item_identification_id,omitempty"`
	StandardItemCode         *string  `json:"standard_item_code,omitempty"`
	UNSPSCCode               *string  `json:"unspsc_code,omitempty"`
	UnitCodeID               *int     `json:"unit_code_id,omitempty"`
	Price                    *float64 `json:"price,omitempty"`
	TaxTypeID                *int     `json:"tax_type_id,omitempty"`
	TaxRate                  *float64 `json:"tax_rate,omitempty"`
	BrandName                *string  `json:"brand_name,omitempty"`
	ModelName                *string  `json:"model_name,omitempty"`
	IsActive                 *bool    `json:"is_active,omitempty"`
}

// ProductListResponse representa la respuesta paginada de productos
type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
