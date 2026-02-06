package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/utils"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetAll obtiene todos los usuarios con paginación
func (s *UserService) GetAll(page, pageSize int) (*domain.UserListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)
	
	users, total, err := s.userRepo.GetAll(page, pageSize)
	if err != nil {
		return nil, err
	}
	
	return &domain.UserListResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetByID obtiene un usuario por ID
func (s *UserService) GetByID(id int64) (*domain.User, error) {
	return s.userRepo.GetByID(id)
}

// Update actualiza un usuario
func (s *UserService) Update(id int64, req *domain.UpdateUserRequest) (*domain.User, error) {
	// Obtener usuario actual
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si se proporcionan
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Verificar si el nuevo email ya existe
		if *req.Email != user.Email {
			exists, err := s.userRepo.EmailExists(*req.Email)
			if err != nil {
				return nil, fmt.Errorf("error checking email: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("email already exists")
			}
			user.Email = *req.Email
		}
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hashing password: %w", err)
		}
		user.Password = string(hashedPassword)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Actualizar en BD
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete elimina un usuario (soft delete)
func (s *UserService) Delete(id int64) error {
	return s.userRepo.Delete(id)
}

// ChangePassword cambia la contraseña de un usuario
func (s *UserService) ChangePassword(userID int64, req *domain.ChangePasswordRequest) error {
	// Obtener usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verificar contraseña actual
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash de la nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Actualizar contraseña
	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

// GetProfile obtiene el perfil del usuario autenticado
func (s *UserService) GetProfile(userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}
