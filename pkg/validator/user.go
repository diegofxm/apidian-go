package validator

import (
	"apidian-go/internal/domain"
	"fmt"
)

// ValidateRegister valida la solicitud de registro
func ValidateRegister(req *domain.RegisterRequest) error {
	if req.Name == "" {
		return fmt.Errorf("el nombre es requerido")
	}
	if len(req.Name) < 3 {
		return fmt.Errorf("el nombre debe tener al menos 3 caracteres")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("el nombre no debe superar 255 caracteres")
	}

	if req.Email == "" {
		return fmt.Errorf("el email es requerido")
	}
	if !IsValidEmail(req.Email) {
		return fmt.Errorf("el email no es válido")
	}

	if req.Password == "" {
		return fmt.Errorf("la contraseña es requerida")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("la contraseña debe tener al menos 8 caracteres")
	}

	return nil
}

// ValidateLogin valida la solicitud de login
func ValidateLogin(req *domain.LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("el email es requerido")
	}
	if !IsValidEmail(req.Email) {
		return fmt.Errorf("el email no es válido")
	}

	if req.Password == "" {
		return fmt.Errorf("la contraseña es requerida")
	}

	return nil
}

// ValidateUpdateUser valida la solicitud de actualización de usuario
func ValidateUpdateUser(req *domain.UpdateUserRequest) error {
	if req.Name != nil {
		if len(*req.Name) < 3 {
			return fmt.Errorf("el nombre debe tener al menos 3 caracteres")
		}
		if len(*req.Name) > 255 {
			return fmt.Errorf("el nombre no debe superar 255 caracteres")
		}
	}

	if req.Email != nil {
		if !IsValidEmail(*req.Email) {
			return fmt.Errorf("el email no es válido")
		}
	}

	if req.Password != nil {
		if len(*req.Password) < 8 {
			return fmt.Errorf("la contraseña debe tener al menos 8 caracteres")
		}
	}

	return nil
}

// ValidateChangePassword valida la solicitud de cambio de contraseña
func ValidateChangePassword(req *domain.ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return fmt.Errorf("la contraseña actual es requerida")
	}

	if req.NewPassword == "" {
		return fmt.Errorf("la nueva contraseña es requerida")
	}
	if len(req.NewPassword) < 8 {
		return fmt.Errorf("la nueva contraseña debe tener al menos 8 caracteres")
	}

	if req.CurrentPassword == req.NewPassword {
		return fmt.Errorf("la nueva contraseña debe ser diferente a la actual")
	}

	return nil
}
