package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"errors"
)

type SoftwareService struct {
	repo        *repository.SoftwareRepository
	companyRepo *repository.CompanyRepository
}

func NewSoftwareService(repo *repository.SoftwareRepository, companyRepo *repository.CompanyRepository) *SoftwareService {
	return &SoftwareService{
		repo:        repo,
		companyRepo: companyRepo,
	}
}

// Create crea una nueva configuración de software
func (s *SoftwareService) Create(userID int64, req *domain.CreateSoftwareRequest) (*domain.Software, error) {
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

	// Verificar que no exista ya un software para esta empresa
	existing, _ := s.repo.GetByCompanyID(req.CompanyID)
	if existing != nil {
		return nil, errors.New("software already exists for this company")
	}

	return s.repo.Create(req)
}

// GetByID obtiene un software por ID
func (s *SoftwareService) GetByID(id int64, userID int64) (*domain.Software, error) {
	software, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verificar que la empresa pertenece al usuario
	company, err := s.companyRepo.GetByID(software.CompanyID)
	if err != nil || company.UserID != userID {
		return nil, errors.New("unauthorized access to software")
	}

	return software, nil
}

// GetByCompanyID obtiene el software de una empresa
func (s *SoftwareService) GetByCompanyID(companyID int64, userID int64) (*domain.Software, error) {
	// Verificar que la empresa pertenece al usuario
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, err
	}

	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	return s.repo.GetByCompanyID(companyID)
}

// Update actualiza un software
func (s *SoftwareService) Update(id int64, userID int64, req *domain.UpdateSoftwareRequest) error {
	// Verificar que el software existe y pertenece al usuario
	software, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Verificar que la empresa pertenece al usuario
	company, err := s.companyRepo.GetByID(software.CompanyID)
	if err != nil || company.UserID != userID {
		return errors.New("unauthorized access to software")
	}

	return s.repo.Update(id, req)
}

// Delete elimina (soft delete) un software
func (s *SoftwareService) Delete(id int64, userID int64) error {
	// Verificar que el software existe y pertenece al usuario
	_, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
