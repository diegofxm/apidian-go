package handler

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/repository"
	"apidian-go/internal/service"
	"apidian-go/pkg/errors"
	"apidian-go/pkg/response"
	"apidian-go/pkg/utils"
	"apidian-go/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    *repository.UserRepository
}

func NewAuthHandler(db *database.Database, cfg *config.Config) *AuthHandler {
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, &cfg.JWT)
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

// Register registers a new user
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateRegister(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Register user
	loginResp, err := h.authService.Register(&req)
	if err != nil {
		if err.Error() == "email already exists" {
			return response.BadRequest(c, errors.ErrEmailExists.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "User registered successfully", loginResp)
}

// Login authenticates a user
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateLogin(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Login
	loginResp, err := h.authService.Login(&req)
	if err != nil {
		return response.Unauthorized(c, errors.ErrInvalidCredentials.Message)
	}

	return response.Success(c, "Login successful", loginResp)
}

// Logout logs out the user
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// With JWT we don't need to invalidate the token on the server
	// The client simply needs to delete the token
	return response.Success(c, "Logout successful", nil)
}

// Me gets the authenticated user's profile
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, errors.ErrUnauthorized.Message)
	}

	// Get complete user from database
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, errors.ErrUserNotFound.Message)
	}

	return response.Success(c, "Profile retrieved successfully", user)
}
