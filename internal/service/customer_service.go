package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/utils"
	"fmt"
)

type CustomerService struct {
	repo *repository.CustomerRepository
}

func NewCustomerService(repo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// Create crea un nuevo cliente
func (s *CustomerService) Create(userID int64, req *domain.CreateCustomerRequest) (*domain.Customer, error) {
	// Validar que no exista cliente con la misma identificación en la empresa
	existing, err := s.repo.GetByIdentification(req.CompanyID, req.IdentificationNumber)
	if err != nil {
		return nil, fmt.Errorf("error checking existing customer: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("customer with identification %s already exists for this company", req.IdentificationNumber)
	}

	// Crear cliente (PostgreSQL maneja validaciones)
	customer, err := s.repo.Create(userID, req)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetByID obtiene un cliente por ID
func (s *CustomerService) GetByID(id int64, companyID int64) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && customer.CompanyID != companyID {
		return nil, fmt.Errorf("unauthorized access to customer")
	}

	return customer, nil
}

// GetByCompanyID obtiene todos los clientes de una empresa con paginación
func (s *CustomerService) GetByCompanyID(companyID int64, page, pageSize int) (*domain.CustomerListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)

	customers, total, err := s.repo.GetByCompanyID(companyID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &domain.CustomerListResponse{
		Customers: customers,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// Update actualiza un cliente
func (s *CustomerService) Update(id int64, companyID int64, req *domain.UpdateCustomerRequest) error {
	// Verificar que el cliente existe
	customer, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && customer.CompanyID != companyID {
		return fmt.Errorf("unauthorized access to customer")
	}

	// Actualizar cliente (PostgreSQL maneja validaciones)
	if err := s.repo.Update(id, req); err != nil {
		return err
	}

	return nil
}

// Delete elimina (soft delete) un cliente
func (s *CustomerService) Delete(id int64, companyID int64) error {
	// Verificar que el cliente existe
	customer, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && customer.CompanyID != companyID {
		return fmt.Errorf("unauthorized access to customer")
	}

	// Eliminar cliente
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	return nil
}
