package service

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtConfig *config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtConfig *config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// Register registra un nuevo usuario
func (s *AuthService) Register(req *domain.RegisterRequest) (*domain.LoginResponse, error) {
	// Verificar si el email ya existe
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Crear usuario
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		// Si es error de duplicate key, retornar mensaje amigable
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Generar JWT token
	token, err := s.generateJWT(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return &domain.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login autentica un usuario
func (s *AuthService) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Buscar usuario por email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generar JWT token
	token, err := s.generateJWT(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return &domain.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// generateJWT genera un token JWT
func (s *AuthService) generateJWT(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.jwtConfig.Expiration)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}
