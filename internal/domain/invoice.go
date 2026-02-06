package domain

import "time"

// Invoice representa una factura electrónica (tabla documents con type_document_id = 1)
type Invoice struct {
	// Campos base (tabla documents)
	ID                     int64          `json:"id"`
	CompanyID              int64          `json:"company_id"`
	CustomerID             int64          `json:"customer_id"`
	ResolutionID           int64          `json:"resolution_id"`
	Number                 string         `json:"number"`
	Consecutive            int64          `json:"consecutive"`
	UUID                   *string        `json:"uuid,omitempty"`
	IssueDate              time.Time      `json:"issue_date"`
	IssueTime              time.Time      `json:"issue_time"`
	DueDate                *time.Time     `json:"due_date,omitempty"`
	TypeDocumentID         int            `json:"type_document_id"`
	CurrencyCodeID         int            `json:"currency_code_id"`
	Notes                  *string        `json:"notes,omitempty"`
	PaymentMethodID        *int           `json:"payment_method_id,omitempty"`
	PaymentFormID          *int           `json:"payment_form_id,omitempty"`
	Subtotal               float64        `json:"subtotal"`
	TaxTotal               float64        `json:"tax_total"`
	Total                  float64        `json:"total"`
	XMLPath                *string        `json:"xml_path,omitempty"`
	PDFPath                *string        `json:"pdf_path,omitempty"`
	ZipPath                *string        `json:"zip_path,omitempty"`
	QRCodeURL              *string        `json:"qr_code_url,omitempty"`
	TrackID                *string        `json:"track_id,omitempty"`
	Status                 string         `json:"status"`
	DIANStatus             *string        `json:"dian_status,omitempty"`
	DIANResponse           *string        `json:"dian_response,omitempty"`
	DIANStatusCode         *string        `json:"dian_status_code,omitempty"`
	DIANStatusDescription  *string        `json:"dian_status_description,omitempty"`
	SentToDIANAt           *time.Time     `json:"sent_to_dian_at,omitempty"`
	AcceptedByDIANAt       *time.Time     `json:"accepted_by_dian_at,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	
	// Códigos DIAN (de catálogos)
	InvoiceTypeCode    string  `json:"invoice_type_code,omitempty"`
	CurrencyCode       string  `json:"currency_code,omitempty"`
	PaymentMethodCode  *string `json:"payment_method_code,omitempty"`
	PaymentMethodName  *string `json:"payment_method_name,omitempty"`
	PaymentFormCode    *string `json:"payment_form_code,omitempty"`
	PaymentFormName    *string `json:"payment_form_name,omitempty"`
	
	// Datos anidados (de JOINs) - Necesarios para generación XML DIAN
	Company    *CompanyDetail       `json:"company,omitempty"`
	Customer   *CustomerDetail      `json:"customer,omitempty"`
	Resolution *ResolutionDetail    `json:"resolution,omitempty"`
	Software   *SoftwareDetail      `json:"software,omitempty"`
	Lines      []InvoiceLineDetail  `json:"lines,omitempty"`
}

// InvoiceLine representa una línea de detalle de una factura (tabla document_lines)
// DEPRECATED: Usar InvoiceLineDetail para datos completos con JOINs
type InvoiceLine struct {
	ID                 int64     `json:"id"`
	DocumentID         int64     `json:"document_id"`
	ProductID          int64     `json:"product_id"`
	LineNumber         int64     `json:"line_number"`
	Description        string    `json:"description"`
	Quantity           float64   `json:"quantity"`
	UnitPrice          float64   `json:"unit_price"`
	LineTotal          float64   `json:"line_total"`
	TaxRate            float64   `json:"tax_rate"`
	TaxAmount          float64   `json:"tax_amount"`
	BrandName          *string   `json:"brand_name,omitempty"`
	ModelName          *string   `json:"model_name,omitempty"`
	StandardItemCode   *string   `json:"standard_item_code,omitempty"`
	ClassificationCode *string   `json:"classification_code,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// CompanyDetail contiene datos completos del emisor (AccountingSupplierParty)
type CompanyDetail struct {
	ID                   int64   `json:"id"`
	NIT                  string  `json:"nit"`
	DV                   *string `json:"dv,omitempty"`
	Name                 string  `json:"name"`
	TradeName            *string `json:"trade_name,omitempty"`
	RegistrationName     string  `json:"registration_name"`
	DocumentTypeCode     string  `json:"document_type_code"`
	DocumentTypeName     string  `json:"document_type_name"`
	TaxLevelCode         string  `json:"tax_level_code"`
	TaxLevelName         string  `json:"tax_level_name"`
	TypeOrganizationCode string  `json:"type_organization_code"`
	TypeRegimeCode       string  `json:"type_regime_code"`
	TypeRegimeName       string  `json:"type_regime_name"`
	IndustryCodes        *string `json:"industry_codes,omitempty"`
	AddressLine          string  `json:"address_line"`
	PostalZone           *string `json:"postal_zone,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	Email                *string `json:"email,omitempty"`
	Website              *string `json:"website,omitempty"`
	Municipality         string  `json:"municipality"`
	MunicipalityCode     string  `json:"municipality_code"`
	Department           string  `json:"department"`
	DepartmentCode       string  `json:"department_code"`
	CountryCode          string  `json:"country_code"`
	CountryName          string  `json:"country_name"`
	TaxSchemeID          string  `json:"tax_scheme_id"`
	TaxSchemeName        string  `json:"tax_scheme_name"`
	LogoPath             *string `json:"logo_path,omitempty"`
}

// CustomerDetail contiene datos completos del adquiriente (AccountingCustomerParty)
type CustomerDetail struct {
	ID                   int64   `json:"id"`
	IdentificationNumber string  `json:"identification_number"`
	DV                   *string `json:"dv,omitempty"`
	Name                 string  `json:"name"`
	TradeName            *string `json:"trade_name,omitempty"`
	DocumentTypeCode     string  `json:"document_type_code"`
	DocumentTypeName     string  `json:"document_type_name"`
	TaxLevelCode         string  `json:"tax_level_code"`
	TaxLevelName         string  `json:"tax_level_name"`
	TypeOrganizationCode string  `json:"type_organization_code"`
	TypeRegimeCode       string  `json:"type_regime_code"`
	TypeRegimeName       string  `json:"type_regime_name"`
	AddressLine          string  `json:"address_line"`
	PostalZone           *string `json:"postal_zone,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	Email                *string `json:"email,omitempty"`
	Municipality         string  `json:"municipality"`
	MunicipalityCode     string  `json:"municipality_code"`
	Department           string  `json:"department"`
	DepartmentCode       string  `json:"department_code"`
	CountryCode          string  `json:"country_code"`
	CountryName          string  `json:"country_name"`
	TaxSchemeID          string  `json:"tax_scheme_id"`
	TaxSchemeName        string  `json:"tax_scheme_name"`
}

// ResolutionDetail contiene datos de la resolución DIAN
type ResolutionDetail struct {
	ID           int64      `json:"id"`
	Prefix       string     `json:"prefix"`
	Resolution   string     `json:"resolution"`
	TechnicalKey *string    `json:"technical_key,omitempty"`
	FromNumber   int64      `json:"from_number"`
	ToNumber     int64      `json:"to_number"`
	DateFrom     time.Time  `json:"date_from"`
	DateTo       time.Time  `json:"date_to"`
}

// SoftwareDetail contiene datos del software DIAN
type SoftwareDetail struct {
	ID          int64   `json:"id"`
	Identifier  string  `json:"identifier"`
	PIN         string  `json:"pin"`
	Environment string  `json:"environment"` // "1" = Producción, "2" = Habilitación
	TestSetID   *string `json:"test_set_id,omitempty"`
}

// InvoiceLineDetail contiene datos completos de una línea con JOINs
type InvoiceLineDetail struct {
	// Campos base (tabla document_lines)
	ID                 int64     `json:"id"`
	DocumentID         int64     `json:"document_id"`
	ProductID          int64     `json:"product_id"`
	LineNumber         int64     `json:"line_number"`
	Description        string    `json:"description"`
	Quantity           float64   `json:"quantity"`
	UnitPrice          float64   `json:"unit_price"`
	LineTotal          float64   `json:"line_total"`
	TaxRate            float64   `json:"tax_rate"`
	TaxAmount          float64   `json:"tax_amount"`
	BrandName          *string   `json:"brand_name,omitempty"`
	ModelName          *string   `json:"model_name,omitempty"`
	StandardItemCode   *string   `json:"standard_item_code,omitempty"`
	ClassificationCode *string   `json:"classification_code,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	
	// Datos del producto (de JOINs)
	ProductCode         string  `json:"product_code"`
	ProductName         string  `json:"product_name"`
	ProductStandardCode *string `json:"product_standard_code,omitempty"`
	UNSPSCCode          *string `json:"unspsc_code,omitempty"`
	
	// Códigos DIAN (de catálogos)
	UnitCode     string `json:"unit_code"`
	UnitName     string `json:"unit_name"`
	TaxTypeCode  string `json:"tax_type_code"`
	TaxTypeName  string `json:"tax_type_name"`
}

// CreateInvoiceRequest representa la solicitud para crear una factura
type CreateInvoiceRequest struct {
	CompanyID       int64                     `json:"company_id" validate:"required"`
	CustomerID      int64                     `json:"customer_id" validate:"required"`
	ResolutionID    int64                     `json:"resolution_id" validate:"required"`
	IssueDate       string                    `json:"issue_date" validate:"required"` // YYYY-MM-DD
	DueDate         *string                   `json:"due_date,omitempty"`             // YYYY-MM-DD
	CurrencyCodeID  int                       `json:"currency_code_id" validate:"required"`
	Notes           *string                   `json:"notes,omitempty"`
	PaymentMethodID *int                      `json:"payment_method_id,omitempty"`
	PaymentFormID   *int                      `json:"payment_form_id,omitempty"` // Opcional: se calcula automáticamente si no se envía
	Lines           []CreateInvoiceLineRequest `json:"lines" validate:"required,min=1,dive"`
}

// CreateInvoiceLineRequest representa la solicitud para crear una línea de factura
type CreateInvoiceLineRequest struct {
	ProductID          int64    `json:"product_id" validate:"required"`
	Description        *string  `json:"description,omitempty"`         // Opcional, se toma de products si no se envía
	Quantity           float64  `json:"quantity" validate:"required,gt=0"`
	UnitPrice          *float64 `json:"unit_price,omitempty"`          // Opcional, se toma de products si no se envía
	TaxRate            *float64 `json:"tax_rate,omitempty"`            // Opcional, se toma de products si no se envía
	BrandName          *string  `json:"brand_name,omitempty"`
	ModelName          *string  `json:"model_name,omitempty"`
	StandardItemCode   *string  `json:"standard_item_code,omitempty"`
	ClassificationCode *string  `json:"classification_code,omitempty"`
}

// UpdateInvoiceRequest representa la solicitud para actualizar una factura
type UpdateInvoiceRequest struct {
	DueDate         *string `json:"due_date,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	PaymentMethodID *int    `json:"payment_method_id,omitempty"`
	PaymentDueDate  *string `json:"payment_due_date,omitempty"`
}

// InvoiceListResponse represents the paginated invoice list response
type InvoiceListResponse struct {
	Invoices []Invoice `json:"invoices"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
