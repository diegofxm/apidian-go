package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrNotFound          = New("NOT_FOUND", "Recurso no encontrado")
	ErrInvalidInput      = New("INVALID_INPUT", "Datos de entrada inválidos")
	ErrUnauthorized      = New("UNAUTHORIZED", "Acceso no autorizado")
	ErrForbidden         = New("FORBIDDEN", "Acceso prohibido")
	ErrInternalServer    = New("INTERNAL_SERVER", "Error interno del servidor")
	ErrDatabaseOperation = New("DATABASE_ERROR", "Error en operación de base de datos")
	ErrDuplicateEntry    = New("DUPLICATE_ENTRY", "El registro ya existe")
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Credenciales inválidas")
	ErrEmailExists       = New("EMAIL_EXISTS", "El email ya está registrado")
	ErrCompanyNotFound   = New("COMPANY_NOT_FOUND", "Empresa no encontrada")
	ErrCustomerNotFound  = New("CUSTOMER_NOT_FOUND", "Cliente no encontrado")
	ErrProductNotFound   = New("PRODUCT_NOT_FOUND", "Producto no encontrado")
	ErrInvoiceNotFound   = New("INVOICE_NOT_FOUND", "Factura no encontrada")
	ErrResolutionNotFound = New("RESOLUTION_NOT_FOUND", "Resolución no encontrada")
	ErrUserNotFound      = New("USER_NOT_FOUND", "Usuario no encontrado")
	ErrSoftwareNotFound  = New("SOFTWARE_NOT_FOUND", "Software no encontrado")
	ErrInvalidNIT        = New("INVALID_NIT", "NIT inválido")
	ErrInvalidDV         = New("INVALID_DV", "Dígito de verificación inválido")
	ErrInvalidCUFE       = New("INVALID_CUFE", "CUFE inválido")
	ErrDIANError         = New("DIAN_ERROR", "Error al comunicarse con DIAN")
)
