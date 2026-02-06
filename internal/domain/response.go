package domain

// DocumentResponse estructura genérica de respuesta para documentos electrónicos
// Compatible con la respuesta de apidian PHP
type DocumentResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Error   string        `json:"error,omitempty"`
	Data    *DocumentData `json:"data,omitempty"`
}

// DocumentData contiene todos los datos del documento procesado
type DocumentData struct {
	// Información básica
	DocumentID int64  `json:"document_id,omitempty"`
	InvoiceID  int64  `json:"invoice_id,omitempty"` // Para compatibilidad
	Number     string `json:"number"`
	CUFE       string `json:"cufe,omitempty"`
	CUDE       string `json:"cude,omitempty"`
	QRStr      string `json:"qr_str,omitempty"`

	// Respuesta de DIAN
	ResponseDian interface{} `json:"ResponseDian,omitempty"`

	// URLs de archivos
	URLInvoiceXML      string `json:"urlinvoicexml,omitempty"`
	URLInvoicePDF      string `json:"urlinvoicepdf,omitempty"`
	URLInvoiceAttached string `json:"urlinvoiceattached,omitempty"`

	// Archivos en base64 (opcional)
	InvoiceXML         string `json:"invoicexml,omitempty"`
	ZipInvoiceXML      string `json:"zipinvoicexml,omitempty"`
	UnsignedInvoiceXML string `json:"unsignedinvoicexml,omitempty"`
	ReqFE              string `json:"reqfe,omitempty"`
	RptaFE             string `json:"rptafe,omitempty"`
	AttachedDocument   string `json:"attacheddocument,omitempty"`
}

// NewSuccessResponse crea una respuesta exitosa
func NewSuccessResponse(message string, data *DocumentData) *DocumentResponse {
	return &DocumentResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse crea una respuesta de error
func NewErrorResponse(err error) *DocumentResponse {
	return &DocumentResponse{
		Success: false,
		Error:   err.Error(),
	}
}
