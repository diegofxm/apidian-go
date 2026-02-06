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

type ResolutionHandler struct {
	service        *service.ResolutionService
	companyService *service.CompanyService
}

func NewResolutionHandler(service *service.ResolutionService, companyService *service.CompanyService) *ResolutionHandler {
	return &ResolutionHandler{
		service:        service,
		companyService: companyService,
	}
}

// GetAll gets all resolutions for a company using NIT/DV
func (h *ResolutionHandler) GetAll(c *fiber.Ctx) error {
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
	company, err := h.companyService.GetByID(companyID, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.BadRequest(c, err.Error())
	}

	// Pagination
	page, pageSize := utils.ParsePaginationParams(c)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get company resolutions
	result, err := h.service.GetByCompanyID(company.ID, userID, page, pageSize)
	if err != nil {
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Resolutions retrieved successfully", result)
}

// GetByID gets a resolution by ID
func (h *ResolutionHandler) GetByID(c *fiber.Ctx) error {
	// Get resolution ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid resolution ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get resolution
	resolution, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "resolution not found" {
			return response.NotFound(c, errors.ErrResolutionNotFound.Message)
		}
		if err.Error() == "unauthorized access to resolution" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Resolution retrieved successfully", resolution)
}

// Create creates a new resolution
func (h *ResolutionHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateResolutionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateCreateResolution(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create resolution
	resolution, err := h.service.Create(userID, &req)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if errMsg == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		if errMsg == "resolution with this prefix already exists for this company" {
			return response.Conflict(c, "A resolution with prefix "+req.Prefix+" already exists for this company")
		}

		// Foreign key errors
		if strings.Contains(errMsg, "fk_resolutions_company") {
			return response.BadRequest(c, "The specified company does not exist")
		}
		if strings.Contains(errMsg, "fk_resolutions_type_document") {
			return response.BadRequest(c, "The specified document type does not exist")
		}

		// CHECK constraint errors
		if strings.Contains(errMsg, "chk_resolutions_range") {
			return response.BadRequest(c, "Invalid numbering range (from_number must be less than or equal to to_number)")
		}
		if strings.Contains(errMsg, "chk_resolutions_dates") {
			return response.BadRequest(c, "Invalid date range (date_from must be less than or equal to date_to)")
		}

		// Unique constraint errors
		if strings.Contains(errMsg, "uq_resolutions_company_prefix") {
			return response.Conflict(c, "A resolution with prefix "+req.Prefix+" already exists for this company")
		}

		return response.InternalServerError(c, "Error creating resolution: "+errMsg)
	}

	return response.Success(c, "Resolution created successfully", resolution)
}

// Delete deletes (soft delete) a resolution
func (h *ResolutionHandler) Delete(c *fiber.Ctx) error {
	// Get resolution ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid resolution ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Delete resolution
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "resolution not found" {
			return response.NotFound(c, errors.ErrResolutionNotFound.Message)
		}
		if err.Error() == "unauthorized access to resolution" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Resolution deleted successfully", nil)
}
