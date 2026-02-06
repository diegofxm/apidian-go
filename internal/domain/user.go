package domain

import "time"

// User representa un usuario del sistema
type User struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	Password        string     `json:"-"` // No exponer en JSON
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RegisterRequest representa la solicitud de registro
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest representa la solicitud de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// UpdateUserRequest representa la solicitud para actualizar un usuario
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// ChangePasswordRequest representa la solicitud para cambiar contraseña
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UserListResponse representa la respuesta paginada de usuarios
type UserListResponse struct {
	Users    []*User `json:"users"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
