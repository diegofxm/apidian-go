package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/utils"
	"apidian-go/pkg/validator"
	"fmt"
)

type CompanyService struct {
	repo *repository.CompanyRepository
}

func NewCompanyService(repo *repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

// Create crea una nueva empresa
func (s *CompanyService) Create(userID int64, req *domain.CreateCompanyRequest) (*domain.Company, error) {
	// Validar NIT con dígito verificador
	if err := validator.ValidateNIT(req.NIT, req.DV); err != nil {
		return nil, err
	}

	// Crear empresa (PostgreSQL maneja validaciones de unique, check constraints, etc.)
	company, err := s.repo.Create(userID, req)
	if err != nil {
		return nil, err
	}

	return company, nil
}

// GetByID obtiene una empresa por ID
func (s *CompanyService) GetByID(id int64, userID int64) (*domain.Company, error) {
	company, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verificar que la empresa pertenece al usuario
	if company.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to company")
	}

	return company, nil
}

// GetByUserID obtiene todas las empresas de un usuario con paginación
func (s *CompanyService) GetByUserID(userID int64, page, pageSize int) (*domain.CompanyListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)

	companies, total, err := s.repo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &domain.CompanyListResponse{
		Companies: companies,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// Update actualiza una empresa
func (s *CompanyService) Update(id int64, userID int64, req *domain.UpdateCompanyRequest) error {
	// Verificar que la empresa existe y pertenece al usuario
	company, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if company.UserID != userID {
		return fmt.Errorf("unauthorized access to company")
	}

	// Actualizar empresa (PostgreSQL maneja validaciones)
	if err := s.repo.Update(id, req); err != nil {
		return err
	}

	return nil
}

// Delete elimina (soft delete) una empresa
func (s *CompanyService) Delete(id int64, userID int64) error {
	// Verificar que la empresa existe y pertenece al usuario
	company, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if company.UserID != userID {
		return fmt.Errorf("unauthorized access to company")
	}

	// Eliminar empresa
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	return nil
}

// GetByNIT obtiene una empresa por NIT y DV validando pertenencia al usuario
func (s *CompanyService) GetByNIT(nit, dv string, userID int64) (*domain.Company, error) {
	// Validar NIT/DV
	if err := validator.ValidateNIT(nit, &dv); err != nil {
		return nil, err
	}
	
	// Obtener empresa por NIT
	company, err := s.repo.GetByNIT(nit, dv)
	if err != nil {
		return nil, err
	}
	
	// Verificar que el usuario tenga acceso a esta empresa
	if company.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to company")
	}
	
	return company, nil
}
