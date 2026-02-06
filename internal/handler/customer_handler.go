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

type CustomerHandler struct {
	service        *service.CustomerService
	companyService *service.CompanyService
}

func NewCustomerHandler(db *database.Database) *CustomerHandler {
	customerRepo := repository.NewCustomerRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	customerService := service.NewCustomerService(customerRepo)
	companyService := service.NewCompanyService(companyRepo)
	return &CustomerHandler{
		service:        customerService,
		companyService: companyService,
	}
}

// GetAll gets all customers with optional company_id filter
func (h *CustomerHandler) GetAll(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Get company_id from query params (optional filter)
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

	// Get company customers
	result, err := h.service.GetByCompanyID(company.ID, page, pageSize)
	if err != nil {
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Customers retrieved successfully", result)
}

// GetByID gets a customer by ID
func (h *CustomerHandler) GetByID(c *fiber.Ctx) error {
	// Get customer ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid customer ID")
	}

	// Get customer directly (service verifies ownership internally)
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}
	customer, err := h.service.GetByID(id, userID)
	if err != nil {
		if err.Error() == "customer not found" {
			return response.NotFound(c, errors.ErrCustomerNotFound.Message)
		}
		if err.Error() == "unauthorized access to customer" {
			return response.Unauthorized(c, "Unauthorized access to customer")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Customer retrieved successfully", customer)
}

// Create creates a new customer
func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	// Get user_id from context
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Parse request body
	var req domain.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request with DIAN rules
	if err := validator.ValidateCreateCustomer(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Create customer
	customer, err := h.service.Create(userID, &req)
	if err != nil {
		errMsg := err.Error()

		// Duplicate customer error
		if errMsg == "customer with identification "+req.IdentificationNumber+" already exists for this company" {
			return response.Conflict(c, "A customer with identification "+req.IdentificationNumber+" already exists for this company")
		}

		// Company does not exist error
		if strings.Contains(errMsg, "fk_customers_company") {
			return response.BadRequest(c, "The specified company does not exist. Verify the company_id")
		}

		// Document type does not exist error
		if strings.Contains(errMsg, "fk_customers_document_type") {
			return response.BadRequest(c, "The specified document type does not exist")
		}

		// Country/department/municipality errors
		if strings.Contains(errMsg, "fk_customers_country") {
			return response.BadRequest(c, "The specified country does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_department") {
			return response.BadRequest(c, "The specified department does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_municipality") {
			return response.BadRequest(c, "The specified municipality does not exist")
		}

		// Catalog errors
		if strings.Contains(errMsg, "fk_customers_tax_level_code") {
			return response.BadRequest(c, "The specified tax level does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_type_organization") {
			return response.BadRequest(c, "The specified organization type does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_type_regime") {
			return response.BadRequest(c, "The specified regime type does not exist")
		}

		return response.InternalServerError(c, "Error creating customer. Please verify the data sent")
	}

	return response.Success(c, "Customer created successfully", customer)
}

// Update updates a customer
func (h *CustomerHandler) Update(c *fiber.Ctx) error {
	// Get customer ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid customer ID")
	}

	// Parse request body
	var req domain.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate request with DIAN rules
	if err := validator.ValidateUpdateCustomer(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	// Update customer (service verifica pertenencia internamente)
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		errMsg := err.Error()

		if errMsg == "customer not found" {
			return response.NotFound(c, errors.ErrCustomerNotFound.Message)
		}
		if errMsg == "unauthorized access to customer" {
			return response.Unauthorized(c, "Unauthorized access to customer")
		}

		// Foreign key errors
		if strings.Contains(errMsg, "fk_customers_department") {
			return response.BadRequest(c, "The specified department does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_municipality") {
			return response.BadRequest(c, "The specified municipality does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_tax_level_code") {
			return response.BadRequest(c, "The specified tax level does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_type_organization") {
			return response.BadRequest(c, "The specified organization type does not exist")
		}
		if strings.Contains(errMsg, "fk_customers_type_regime") {
			return response.BadRequest(c, "The specified regime type does not exist")
		}

		return response.InternalServerError(c, "Error updating customer")
	}

	return response.Success(c, "Customer updated successfully", nil)
}

// Delete deletes (soft delete) a customer
func (h *CustomerHandler) Delete(c *fiber.Ctx) error {
	// Get customer ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid customer ID")
	}

	// Delete customer (service verifies ownership internally)
	userID, err := utils.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "customer not found" {
			return response.NotFound(c, errors.ErrCustomerNotFound.Message)
		}
		if err.Error() == "unauthorized access to customer" {
			return response.Unauthorized(c, "Unauthorized access to customer")
		}
		return response.InternalServerError(c, errors.ErrInternalServer.Message)
	}

	return response.Success(c, "Customer deleted successfully", nil)
}
