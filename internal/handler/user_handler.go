package handler

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"apidian-go/internal/repository"
	"apidian-go/internal/service"
	"apidian-go/pkg/errors"
	"apidian-go/pkg/response"
	"apidian-go/pkg/utils"
	"apidian-go/pkg/validator"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(db *database.Database) *UserHandler {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	return &UserHandler{userService: userService}
}

// GetAll gets all users with pagination
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	// Get pagination parameters
	page, pageSize := utils.ParsePaginationParams(c)
	
	result, err := h.userService.GetAll(page, pageSize)
	if err != nil {
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Users retrieved successfully", result)
}

// GetByID gets a user by ID
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			return response.NotFound(c, errors.ErrUserNotFound.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "User retrieved successfully", user)
}

// Update updates a user
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	var req domain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateUpdateUser(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Update user
	updatedUser, err := h.userService.Update(id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			return response.NotFound(c, errors.ErrUserNotFound.Message)
		}
		if err.Error() == "email already exists" {
			return response.BadRequest(c, errors.ErrEmailExists.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "User updated successfully", updatedUser)
}

// Delete deletes a user (soft delete)
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	// Verify not deleting themselves
	userID, err := utils.GetUserID(c)
	if err == nil && userID == id {
		return response.BadRequest(c, "Cannot delete your own user")
	}

	if err := h.userService.Delete(id); err != nil {
		if err.Error() == "user not found" {
			return response.NotFound(c, errors.ErrUserNotFound.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "User deleted successfully", nil)
}

// ChangePassword changes the authenticated user's password
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req domain.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateChangePassword(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Change password
	if err := h.userService.ChangePassword(userID, &req); err != nil {
		if err.Error() == "current password is incorrect" {
			return response.BadRequest(c, "Current password is incorrect")
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "Password changed successfully", nil)
}
