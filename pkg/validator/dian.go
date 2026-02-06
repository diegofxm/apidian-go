package validator

import (
	"fmt"
	"strconv"
)

// ValidateNIT valida el NIT según reglas DIAN
func ValidateNIT(nit string, dv *string) error {
	// NIT requerido
	if nit == "" {
		return NewError("nit", "es requerido")
	}

	// Solo números
	if !IsNumeric(nit) {
		return NewError("nit", "debe contener solo números")
	}

	// Longitud entre 6 y 20 (Colombia tiene cédulas antiguas de 6-7 dígitos)
	if err := IsValidLength(nit, 6, 20, "nit"); err != nil {
		return err
	}

	// Validar dígito de verificación si se proporciona
	if dv != nil && *dv != "" {
		calculatedDV := CalculateDV(nit)
		if *dv != calculatedDV {
			return NewError("dv", fmt.Sprintf("dígito de verificación inválido. Esperado: %s, Recibido: %s", calculatedDV, *dv))
		}
	}

	return nil
}

// CalculateDV calcula el dígito de verificación del NIT según algoritmo DIAN
func CalculateDV(nit string) string {
	primes := []int{3, 7, 13, 17, 19, 23, 29, 37, 41, 43, 47, 53, 59, 67, 71}
	sum := 0

	nitLen := len(nit)
	for i := 0; i < nitLen; i++ {
		digit, _ := strconv.Atoi(string(nit[nitLen-1-i]))
		if i < len(primes) {
			sum += digit * primes[i]
		}
	}

	remainder := sum % 11
	if remainder == 0 || remainder == 1 {
		return strconv.Itoa(remainder)
	}

	return strconv.Itoa(11 - remainder)
}

// ValidateIdentification valida número de identificación según DIAN
// Acepta alpha_num, longitud entre 1 y 15
func ValidateIdentification(identification string, fieldName string) error {
	if identification == "" {
		return NewError(fieldName, "es requerido")
	}

	if !IsAlphaNumeric(identification) {
		return NewError(fieldName, "debe contener solo letras y números")
	}

	if err := IsValidLength(identification, 1, 15, fieldName); err != nil {
		return err
	}

	return nil
}

// ValidateIndustryCodes valida códigos CIIU según DIAN
// Máximo 4 códigos
func ValidateIndustryCodes(codes []string) error {
	if len(codes) > 4 {
		return NewError("industry_codes", "máximo 4 códigos CIIU permitidos según DIAN")
	}

	// Validar formato de cada código (4 dígitos)
	for i, code := range codes {
		if !IsNumeric(code) {
			return NewError("industry_codes", fmt.Sprintf("código CIIU #%d debe contener solo números", i+1))
		}
		if len(code) != 4 {
			return NewError("industry_codes", fmt.Sprintf("código CIIU #%d debe tener 4 dígitos", i+1))
		}
	}

	return nil
}

// ValidateEmail valida email según formato DIAN
func ValidateEmail(email string, fieldName string) error {
	if email == "" {
		return nil // Email es opcional
	}

	if !IsValidEmail(email) {
		return NewError(fieldName, "formato de email inválido")
	}

	if err := IsValidLength(email, 5, 100, fieldName); err != nil {
		return err
	}

	return nil
}

// ValidateAmount valida montos monetarios según DIAN
// Debe ser positivo o cero
func ValidateAmount(amount float64, fieldName string) error {
	if amount < 0 {
		return NewError(fieldName, "debe ser un valor positivo o cero")
	}

	return nil
}

// ValidatePercentage valida porcentajes según DIAN
// Debe estar entre 0 y 100
func ValidatePercentage(percentage float64, fieldName string) error {
	if percentage < 0 || percentage > 100 {
		return NewError(fieldName, "debe estar entre 0 y 100")
	}

	return nil
}

// ValidatePostalCode valida código postal colombiano
// 6 dígitos
func ValidatePostalCode(code string) error {
	if code == "" {
		return nil // Opcional
	}

	if !IsNumeric(code) {
		return NewError("postal_zone", "debe contener solo números")
	}

	if len(code) != 6 {
		return NewError("postal_zone", "debe tener 6 dígitos")
	}

	return nil
}

// ValidateCUFE valida el formato del CUFE (Código Único de Factura Electrónica)
// El CUFE es un hash SHA-384 de 96 caracteres hexadecimales
func ValidateCUFE(cufe string) error {
	if cufe == "" {
		return NewError("cufe", "es requerido")
	}

	// El CUFE debe tener exactamente 96 caracteres (SHA-384 en hexadecimal)
	if len(cufe) != 96 {
		return NewError("cufe", "debe tener exactamente 96 caracteres")
	}

	// Validar que solo contenga caracteres hexadecimales (0-9, a-f, A-F)
	for _, char := range cufe {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return NewError("cufe", "debe contener solo caracteres hexadecimales")
		}
	}

	return nil
}

// ValidateEnvironment valida el ambiente DIAN
func ValidateEnvironment(environment string) error {
	if environment != "1" && environment != "2" {
		return NewError("environment", "debe ser '1' (Producción) o '2' (Habilitación)")
	}
	return nil
}
