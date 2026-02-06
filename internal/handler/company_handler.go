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
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CompanyHandler struct {
	service *service.CompanyService
	storage *config.StorageConfig
}

func NewCompanyHandler(db *database.Database, cfg *config.Config) *CompanyHandler {
	repo := repository.NewCompanyRepository(db)
	svc := service.NewCompanyService(repo)
	return &CompanyHandler{service: svc, storage: &cfg.Storage}
}

func (h *CompanyHandler) parseBody(c *fiber.Ctx, out any) error {
	contentType := strings.ToLower(c.Get("Content-Type"))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		data := c.FormValue("data")
		// Si no hay campo "data", es porque solo viene el logo (actualización parcial)
		// Devolvemos nil para permitir que el request continúe con campos vacíos
		if strings.TrimSpace(data) == "" {
			return nil
		}
		return json.Unmarshal([]byte(data), out)
	}
	return c.BodyParser(out)
}

func (h *CompanyHandler) saveCompanyLogo(c *fiber.Ctx, nit string) (*string, error) {
	file, err := c.FormFile("logo")
	if err != nil {
		return nil, nil
	}
	if file == nil {
		return nil, nil
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Unsupported logo file type. Use .png, .jpg, .jpeg")
	}

	// Guardar logo en storage/app/companies/{nit}/profile/logo.png
	logoAbs := h.storage.CompanyLogoPath(nit)
	// Cambiar extensión si no es .png
	if ext != ".png" {
		logoAbs = strings.TrimSuffix(logoAbs, ".png") + ext
	}
	logoRel := filepath.ToSlash(filepath.Join("companies", nit, "profile", "logo"+ext))
	
	if err := os.MkdirAll(filepath.Dir(logoAbs), 0755); err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create logo directory")
	}
	if err := c.SaveFile(file, logoAbs); err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to save logo file")
	}
	return &logoRel, nil
}

// GetAll gets all companies for the authenticated user
func (h *CompanyHandler) GetAll(c *fiber.Ctx) error {
	// Get user_id from context (set by auth middleware)
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get pagination parameters
	page, pageSize := utils.ParsePaginationParams(c)

	// Get companies
	companies, err := h.service.GetByUserID(userID, page, pageSize)
	if err != nil {
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Companies retrieved successfully", companies)
}

// GetByID gets a company by ID
func (h *CompanyHandler) GetByID(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid company ID")
	}

	// Get company
	company, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, errors.ErrCompanyNotFound.Message)
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, errors.ErrUnauthorized.Message)
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Company retrieved successfully", company)
}

// Create creates a new company
func (h *CompanyHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateCompanyRequest
	if err := h.parseBody(c, &req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	logoPath, err := h.saveCompanyLogo(c, req.NIT)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"success": false, "error": fe.Message})
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}
	if logoPath != nil {
		req.LogoPath = logoPath
	}

	// Validate request with DIAN rules
	if err := validator.ValidateCreateCompany(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create company
	company, err := h.service.Create(userID, &req)
	if err != nil {
		errMsg := err.Error()
		
		// Duplicate NIT error
		if strings.Contains(errMsg, "uq_companies_nit") || strings.Contains(errMsg, "duplicate key") {
			return response.Conflict(c, "A company with NIT "+req.NIT+" already exists")
		}
		
		// Foreign key errors
		if strings.Contains(errMsg, "fk_companies_document_type") {
			return response.BadRequest(c, "The specified document type does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_tax_level_code") {
			return response.BadRequest(c, "The specified tax level does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_type_organization") {
			return response.BadRequest(c, "The specified organization type does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_type_regime") {
			return response.BadRequest(c, "The specified regime type does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_country") {
			return response.BadRequest(c, "The specified country does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_department") {
			return response.BadRequest(c, "The specified department does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_municipality") {
			return response.BadRequest(c, "The specified municipality does not exist")
		}
		
		// CHECK constraint error
		if strings.Contains(errMsg, "chk_companies_industry_codes") {
			return response.BadRequest(c, "Must specify between 1 and 3 CIIU codes")
		}
		
		return response.InternalServerError(c, "Error creating company. Please verify the data sent")
	}

	return response.Created(c, "Company created successfully", company)
}

// Update updates a company
func (h *CompanyHandler) Update(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid company ID")
	}

	// Get company first to validate permissions and get NIT for logo path
	company, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	// Parse request body (puede estar vacío si solo viene logo)
	var req domain.UpdateCompanyRequest
	if err := h.parseBody(c, &req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Ignorar logo_path si viene en el JSON (solo se acepta vía archivo)
	req.LogoPath = nil

	// Validate request
	if err := validator.ValidateUpdateCompany(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Save logo if present
	logoPath, err := h.saveCompanyLogo(c, company.NIT)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"success": false, "error": fe.Message})
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}
	if logoPath != nil {
		req.LogoPath = logoPath
	}

	// Update company (solo si hay cambios)
	if err := h.service.Update(id, userID, &req); err != nil {
		errMsg := err.Error()
		
		if errMsg == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if errMsg == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		
		// Foreign key errors
		if strings.Contains(errMsg, "fk_companies_tax_level_code") {
			return response.BadRequest(c, "The specified tax level does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_type_organization") {
			return response.BadRequest(c, "The specified organization type does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_type_regime") {
			return response.BadRequest(c, "The specified regime type does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_department") {
			return response.BadRequest(c, "The specified department does not exist")
		}
		if strings.Contains(errMsg, "fk_companies_municipality") {
			return response.BadRequest(c, "The specified municipality does not exist")
		}
		
		// CHECK constraint error
		if strings.Contains(errMsg, "chk_companies_industry_codes") {
			return response.BadRequest(c, "Must specify between 1 and 3 CIIU codes")
		}
		
		return response.InternalServerError(c, "Error updating company")
	}

	return response.Success(c, "Company updated successfully", nil)
}

// Delete deletes a company (soft delete)
func (h *CompanyHandler) Delete(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid company ID")
	}

	// Delete company
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "company not found" {
			return response.NotFound(c, "Company not found")
		}
		if err.Error() == "unauthorized access to company" {
			return response.Unauthorized(c, "Unauthorized access to company")
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, "Company deleted successfully", nil)
}

// UploadCertificate is deprecated - use POST /api/v1/certificates instead
// This endpoint is kept for backward compatibility
func (h *CompanyHandler) UploadCertificate(c *fiber.Ctx) error {
	return response.BadRequest(c, "This endpoint is deprecated. Please use POST /api/v1/certificates with company_id in the request body")
}
