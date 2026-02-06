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
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CertificateHandler struct {
	service *service.CertificateService
}

func NewCertificateHandler(db *database.Database, cfg *config.Config) *CertificateHandler {
	certRepo := repository.NewCertificateRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	certService := service.NewCertificateService(certRepo, companyRepo, &cfg.Storage)

	return &CertificateHandler{
		service: certService,
	}
}

// Create uploads a new certificate for a company
func (h *CertificateHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateCertificateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate required fields
	if req.Certificate == "" {
		return response.BadRequest(c, "The 'certificate' field is required")
	}
	if req.Password == "" {
		return response.BadRequest(c, "The 'password' field is required")
	}
	if req.CompanyID == 0 {
		return response.BadRequest(c, "The 'company_id' field is required")
	}

	// Create certificate
	cert, err := h.service.Create(&req, userID)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if errMsg == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		if errMsg == "certificate must be valid base64 encoded data" {
			return response.BadRequest(c, "Certificate must be valid base64 encoded data")
		}
		if errMsg == "certificate size must not exceed 5MB" {
			return response.BadRequest(c, "Certificate size must not exceed 5MB")
		}

		return response.InternalServerError(c, "Error uploading certificate: "+errMsg)
	}

	return response.Success(c, "Certificate uploaded successfully", cert)
}

// GetByCompanyID gets the active certificate for a company
func (h *CertificateHandler) GetByCompanyID(c *fiber.Ctx) error {
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

	// Get certificate
	cert, err := h.service.GetByCompanyID(companyID, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		if err.Error() == "certificate not found" {
			return response.NotFound(c, "Certificate not found")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Certificate retrieved successfully", cert)
}

// GetAllByCompanyID gets all certificates for a company (history)
func (h *CertificateHandler) GetAllByCompanyID(c *fiber.Ctx) error {
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

	// Get certificates
	certs, err := h.service.GetAllByCompanyID(companyID, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Certificates retrieved successfully", fiber.Map{
		"certificates": certs,
		"total":        len(certs),
	})
}

// Delete deletes a certificate
func (h *CertificateHandler) Delete(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get certificate ID from params
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid certificate ID")
	}

	// Delete certificate
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "certificate not found" {
			return response.NotFound(c, "Certificate not found")
		}
		if err.Error() == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if err.Error() == "unauthorized access to certificate" {
			return response.Unauthorized(c, "Unauthorized access to certificate")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Certificate deleted successfully", nil)
}
