package handler

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/service"
	"apidian-go/pkg/errors"
	"apidian-go/pkg/response"
	"apidian-go/pkg/utils"
	"apidian-go/pkg/validator"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type SoftwareHandler struct {
	service        *service.SoftwareService
	companyService *service.CompanyService
}

func NewSoftwareHandler(service *service.SoftwareService, companyService *service.CompanyService) *SoftwareHandler {
	return &SoftwareHandler{
		service:        service,
		companyService: companyService,
	}
}

// GetByCompanyID gets the software for a company using NIT/DV
func (h *SoftwareHandler) GetByCompanyID(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company_id from query params
	companyIDStr := c.Query("company_id")
	if companyIDStr == "" {
		return response.BadRequest(c, "company_id query parameter is required")
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid company_id")
	}

	// Verify company belongs to user
	_, err = h.companyService.GetByID(companyID, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.BadRequest(c, err.Error())
	}

	// Get software
	software, err := h.service.GetByCompanyID(companyID, userID)
	if err != nil {
		if err.Error() == "software not found" {
			return response.NotFound(c, errors.ErrSoftwareNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Software retrieved successfully", software)
}

// GetByID gets a software by ID
func (h *SoftwareHandler) GetByID(c *fiber.Ctx) error {
	// Get software ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid software ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get software
	software, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "software not found" {
			return response.NotFound(c, errors.ErrSoftwareNotFound.Message)
		}
		if err.Error() == "unauthorized access to software" {
			return response.Unauthorized(c, "Unauthorized access to software")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Software retrieved successfully", software)
}

// Create creates a new software configuration
func (h *SoftwareHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateSoftwareRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateCreateSoftware(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create software
	software, err := h.service.Create(userID, &req)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if errMsg == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		if errMsg == "software already exists for this company" {
			return response.Conflict(c, "A software configuration already exists for this company")
		}

		// Foreign key errors
		if strings.Contains(errMsg, "fk_software_company") {
			return response.BadRequest(c, "The specified company does not exist")
		}

		// CHECK constraint errors
		if strings.Contains(errMsg, "chk_software_environment") {
			return response.BadRequest(c, "Environment must be '1' (Production) or '2' (Enablement)")
		}

		// Unique constraint errors
		if strings.Contains(errMsg, "software_company_id_key") {
			return response.Conflict(c, "A software configuration already exists for this company")
		}

		return response.InternalServerError(c, "Error creating software: "+errMsg)
	}

	return response.Success(c, "Software created successfully", software)
}

// Update updates a software
func (h *SoftwareHandler) Update(c *fiber.Ctx) error {
	// Get software ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid software ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.UpdateSoftwareRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateUpdateSoftware(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Update software
	if err := h.service.Update(id, userID, &req); err != nil {
		errMsg := err.Error()

		if errMsg == "software not found" {
			return response.NotFound(c, errors.ErrSoftwareNotFound.Message)
		}
		if errMsg == "unauthorized access to software" {
			return response.Unauthorized(c, "Unauthorized access to software")
		}

		// CHECK constraint errors
		if strings.Contains(errMsg, "chk_software_environment") {
			return response.BadRequest(c, "Environment must be '1' (Production) or '2' (Enablement)")
		}

		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Software updated successfully", nil)
}

// Delete deletes (soft delete) a software
func (h *SoftwareHandler) Delete(c *fiber.Ctx) error {
	// Get software ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid software ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Delete software
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "software not found" {
			return response.NotFound(c, errors.ErrSoftwareNotFound.Message)
		}
		if err.Error() == "unauthorized access to software" {
			return response.Unauthorized(c, "Unauthorized access to software")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Software deleted successfully", nil)
}
