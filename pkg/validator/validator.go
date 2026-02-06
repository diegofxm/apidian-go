package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode"
)

// ValidationError representa un error de validación con mensaje amigable
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// NewError crea un nuevo error de validación
func NewError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsRequired valida que un campo no esté vacío
func IsRequired(value string, fieldName string) error {
	if value == "" {
		return NewError(fieldName, "es requerido")
	}
	return nil
}

// IsNumeric valida que un string contenga solo números
func IsNumeric(value string) bool {
	for _, char := range value {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return len(value) > 0
}

// IsAlphaNumeric valida que un string contenga solo letras y números
func IsAlphaNumeric(value string) bool {
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return len(value) > 0
}

// IsValidEmail valida el formato de un email
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidLength valida que un string esté dentro de un rango de longitud
func IsValidLength(value string, min, max int, fieldName string) error {
	length := len(value)
	if length < min {
		return NewError(fieldName, fmt.Sprintf("debe tener al menos %d caracteres", min))
	}
	if length > max {
		return NewError(fieldName, fmt.Sprintf("debe tener máximo %d caracteres", max))
	}
	return nil
}

// IsValidRange valida que un número esté dentro de un rango
func IsValidRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return NewError(fieldName, fmt.Sprintf("debe estar entre %d y %d", min, max))
	}
	return nil
}

// IsValidDecimal valida que un string sea un decimal válido
func IsValidDecimal(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// IsValidArrayLength valida que un array tenga una longitud específica
func IsValidArrayLength(length, max int, fieldName string) error {
	if length > max {
		return NewError(fieldName, fmt.Sprintf("debe tener máximo %d elementos", max))
	}
	return nil
}

// IsValidPhone valida formato de teléfono colombiano
func IsValidPhone(phone string) bool {
	// Acepta formatos: 3001234567, 6011234567, +573001234567
	phoneRegex := regexp.MustCompile(`^(\+57)?[36]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidURL valida formato de URL
func IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)
	return urlRegex.MatchString(url)
}
