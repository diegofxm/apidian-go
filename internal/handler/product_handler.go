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
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	service        *service.ProductService
	companyService *service.CompanyService
}

func NewProductHandler(db *database.Database) *ProductHandler {
	productRepo := repository.NewProductRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	productService := service.NewProductService(productRepo)
	companyService := service.NewCompanyService(companyRepo)
	return &ProductHandler{
		service:        productService,
		companyService: companyService,
	}
}

// GetAll gets all products for a company using NIT/DV
func (h *ProductHandler) GetAll(c *fiber.Ctx) error {
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

	// Get pagination parameters
	page, pageSize := utils.ParsePaginationParams(c)

	// Get company products
	products, err := h.service.GetByCompanyID(company.ID, page, pageSize)
	if err != nil {
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Products retrieved successfully", products)
}

// GetByID gets a product by ID
func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	// Get product ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid product ID")
	}

	// Get product
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}
	product, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "product not found" {
			return response.NotFound(c, errors.ErrProductNotFound.Message)
		}
		if err.Error() == "unauthorized access to product" {
			return response.Unauthorized(c, "Unauthorized access to product")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Product retrieved successfully", product)
}

// Create creates a new product
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateCreateProduct(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create product
	product, err := h.service.Create(userID, &req)
	if err != nil {
		errMsg := err.Error()

		// Duplicate code error
		if strings.Contains(errMsg, "already exists for this company") {
			return response.Conflict(c, "A product with code "+req.Code+" already exists for this company")
		}

		// Foreign key errors
		if strings.Contains(errMsg, "fk_products_company") {
			return response.BadRequest(c, "The specified company does not exist")
		}
		if strings.Contains(errMsg, "fk_products_unit_code") {
			return response.BadRequest(c, "The specified unit of measure does not exist")
		}
		if strings.Contains(errMsg, "fk_products_tax_type") {
			return response.BadRequest(c, "El tipo de impuesto especificado no existe")
		}

		// Errores de CHECK constraints
		if strings.Contains(errMsg, "chk_products_price") {
			return response.BadRequest(c, "El precio debe ser mayor o igual a 0")
		}
		if strings.Contains(errMsg, "chk_products_tax_rate") {
			return response.BadRequest(c, "La tasa de impuesto debe estar entre 0 y 100")
		}

		// Return the real error for debugging
		return response.InternalServerError(c, "Error creating product: "+errMsg)
	}

	return response.Success(c, "Product created successfully", product)
}

// Update updates a product
func (h *ProductHandler) Update(c *fiber.Ctx) error {
	// Get product ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid product ID")
	}

	// Parse request body
	var req domain.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request
	if err := validator.ValidateUpdateProduct(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Update product
	if err := h.service.Update(id, 0, &req); err != nil {
		errMsg := err.Error()

		if errMsg == "product not found" {
			return response.NotFound(c, "Product not found")
		}
		if errMsg == "unauthorized access to product" {
			return response.Unauthorized(c, "Unauthorized access to product")
		}

		// Foreign key errors
		if strings.Contains(errMsg, "fk_products_unit_code") {
			return response.BadRequest(c, "The specified unit of measure does not exist")
		}
		if strings.Contains(errMsg, "fk_products_tax_type") {
			return response.BadRequest(c, "The specified tax type does not exist")
		}

		// CHECK constraint errors
		if strings.Contains(errMsg, "chk_products_price") {
			return response.BadRequest(c, "Price must be greater than or equal to 0")
		}
		if strings.Contains(errMsg, "chk_products_tax_rate") {
			return response.BadRequest(c, "Tax rate must be between 0 and 100")
		}

		return response.InternalServerError(c, "Error updating product")
	}

	return response.Success(c, "Product updated successfully", nil)
}

// Delete deletes (soft delete) a product
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	// Get product ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid product ID")
	}

	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Delete product
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "product not found" {
			return response.NotFound(c, errors.ErrProductNotFound.Message)
		}
		if err.Error() == "unauthorized access to product" {
			return response.Unauthorized(c, "Unauthorized access to product")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Product deleted successfully", nil)
}
