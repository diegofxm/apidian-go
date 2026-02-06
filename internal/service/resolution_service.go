package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/utils"
	"errors"
)

type ResolutionService struct {
	repo        *repository.ResolutionRepository
	companyRepo *repository.CompanyRepository
}

func NewResolutionService(repo *repository.ResolutionRepository, companyRepo *repository.CompanyRepository) *ResolutionService {
	return &ResolutionService{
		repo:        repo,
		companyRepo: companyRepo,
	}
}

// Create crea una nueva resolución
func (s *ResolutionService) Create(userID int64, req *domain.CreateResolutionRequest) (*domain.Resolution, error) {
	// Verificar que la empresa existe y pertenece al usuario
	company, err := s.companyRepo.GetByID(req.CompanyID)
	if err != nil {
		if err.Error() == "company not found" {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	// Verificar que la empresa pertenece al usuario
	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	// Verificar que no exista ya una resolución con el mismo prefijo para esta empresa
	existing, _ := s.repo.GetByCompanyAndPrefix(req.CompanyID, req.Prefix)
	if existing != nil {
		return nil, errors.New("resolution with this prefix already exists for this company")
	}

	return s.repo.Create(req)
}

// GetByID obtiene una resolución por ID
func (s *ResolutionService) GetByID(id int64, userID int64) (*domain.Resolution, error) {
	resolution, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verificar que la empresa pertenece al usuario
	company, err := s.companyRepo.GetByID(resolution.CompanyID)
	if err != nil || company.UserID != userID {
		return nil, errors.New("unauthorized access to resolution")
	}

	return resolution, nil
}

// GetByCompanyID obtiene todas las resoluciones de una empresa
func (s *ResolutionService) GetByCompanyID(companyID, userID int64, page, pageSize int) (*domain.ResolutionListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, err
	}

	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	return s.repo.GetByCompanyID(companyID, page, pageSize)
}

// GetByUserID obtiene todas las resoluciones de todas las empresas del usuario
func (s *ResolutionService) GetByUserID(userID int64, page, pageSize int) (*domain.ResolutionListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)
	return s.repo.GetByUserID(userID, page, pageSize)
}

// Delete elimina (soft delete) una resolución
func (s *ResolutionService) Delete(id int64, userID int64) error {
	// Verificar que la resolución existe y pertenece al usuario
	_, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
