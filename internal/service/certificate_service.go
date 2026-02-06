package service

import (
	"apidian-go/internal/config"
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/crypto"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CertificateService struct {
	certRepo    *repository.CertificateRepository
	companyRepo *repository.CompanyRepository
	storage     *config.StorageConfig
}

func NewCertificateService(certRepo *repository.CertificateRepository, companyRepo *repository.CompanyRepository, storage *config.StorageConfig) *CertificateService {
	return &CertificateService{
		certRepo:    certRepo,
		companyRepo: companyRepo,
		storage:     storage,
	}
}

// Create uploads and stores a new certificate
func (s *CertificateService) Create(req *domain.CreateCertificateRequest, userID int64) (*domain.CertificateResponse, error) {
	// Validate company exists and belongs to user
	company, err := s.companyRepo.GetByID(req.CompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	// Decode base64 certificate
	certificateData, err := crypto.DecodePKCS12(req.Certificate)
	if err != nil {
		return nil, err
	}

	// Validate certificate size (max 5MB)
	if len(certificateData) > 5*1024*1024 {
		return nil, errors.New("certificate size must not exceed 5MB")
	}

	// Note: We don't validate PKCS12 format here to support both BER and DER encodings
	// Validation will occur when the certificate is actually used for signing
	// This allows compatibility with certificates from various Certificate Authorities

	// Generate filename based on company NIT with timestamp for historical tracking
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d.p12", company.NIT, timestamp)

	certificatesDir := s.storage.CertificatesPath(company.NIT)
	if err := os.MkdirAll(certificatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Save certificate to filesystem
	certPath := filepath.Join(certificatesDir, filename)
	if err := os.WriteFile(certPath, certificateData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	// Encrypt password before storing
	encryptedPassword, err := crypto.EncryptPassword(req.Password)
	if err != nil {
		// Clean up file if password encryption fails
		os.Remove(certPath)
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Delete previous certificates (hard delete from DB and filesystem)
	// This ensures only one certificate exists per company at any time
	if err := s.deletePreviousCertificates(req.CompanyID, company.NIT); err != nil {
		// Clean up newly uploaded file if deletion of previous certs fails
		os.Remove(certPath)
		return nil, fmt.Errorf("failed to delete previous certificates: %w", err)
	}

	// Create certificate record (now safe because previous ones are deleted)
	cert := &domain.Certificate{
		CompanyID: req.CompanyID,
		Name:      filename,
		Password:  encryptedPassword,
		IsActive:  true,
	}

	cert, err = s.certRepo.Create(cert)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(certPath)
		return nil, fmt.Errorf("failed to create certificate record: %w", err)
	}

	return &domain.CertificateResponse{
		ID:        cert.ID,
		CompanyID: cert.CompanyID,
		Name:      cert.Name,
		Path:      certPath,
		IsActive:  cert.IsActive,
		CreatedAt: cert.CreatedAt,
		UpdatedAt: cert.UpdatedAt,
	}, nil
}

// GetByCompanyID gets the active certificate for a company
func (s *CertificateService) GetByCompanyID(companyID int64, userID int64) (*domain.CertificateResponse, error) {
	// Validate company belongs to user
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	cert, err := s.certRepo.GetByCompanyID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("certificate not found")
		}
		return nil, err
	}

	certPath := s.GetCertificatePath(cert.Name, company.NIT)

	return &domain.CertificateResponse{
		ID:        cert.ID,
		CompanyID: cert.CompanyID,
		Name:      cert.Name,
		Path:      certPath,
		IsActive:  cert.IsActive,
		CreatedAt: cert.CreatedAt,
		UpdatedAt: cert.UpdatedAt,
	}, nil
}

// GetAllByCompanyID gets all certificates for a company
func (s *CertificateService) GetAllByCompanyID(companyID int64, userID int64) ([]domain.CertificateResponse, error) {
	// Validate company belongs to user
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	if company.UserID != userID {
		return nil, errors.New("unauthorized access to company")
	}

	certs, err := s.certRepo.GetAllByCompanyID(companyID)
	if err != nil {
		return nil, err
	}

	var responses []domain.CertificateResponse
	for _, cert := range certs {
		certPath := s.GetCertificatePath(cert.Name, company.NIT)
		responses = append(responses, domain.CertificateResponse{
			ID:        cert.ID,
			CompanyID: cert.CompanyID,
			Name:      cert.Name,
			Path:      certPath,
			IsActive:  cert.IsActive,
			CreatedAt: cert.CreatedAt,
			UpdatedAt: cert.UpdatedAt,
		})
	}

	return responses, nil
}

// Delete deletes a certificate
func (s *CertificateService) Delete(id int64, userID int64) error {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("certificate not found")
		}
		return err
	}

	// Validate company belongs to user
	company, err := s.companyRepo.GetByID(cert.CompanyID)
	if err != nil {
		return errors.New("company not found")
	}

	if company.UserID != userID {
		return errors.New("unauthorized access to certificate")
	}

	// Soft delete in database
	if err := s.certRepo.Delete(id); err != nil {
		return err
	}

	// Optionally delete file from filesystem (company already retrieved above)
	certPath := s.GetCertificatePath(cert.Name, company.NIT)
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to delete certificate file: %v\n", err)
	}

	return nil
}

// deletePreviousCertificates deletes all previous certificates for a company
// This includes both database records and physical files
func (s *CertificateService) deletePreviousCertificates(companyID int64, companyNIT string) error {
	// Get all existing certificates for this company
	certs, err := s.certRepo.GetAllByCompanyIDIncludingInactive(companyID)
	if err != nil {
		return fmt.Errorf("failed to get existing certificates: %w", err)
	}

	// Delete physical files first
	for _, cert := range certs {
		certPath := s.GetCertificatePath(cert.Name, companyNIT)
		if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
			// Log warning but continue - file might already be deleted
			fmt.Printf("Warning: failed to delete certificate file %s: %v\n", certPath, err)
		} else if err == nil {
			fmt.Printf("Deleted certificate file: %s\n", certPath)
		}
	}

	// Delete all database records for this company
	if err := s.certRepo.DeleteAllByCompanyID(companyID); err != nil {
		return fmt.Errorf("failed to delete certificate records: %w", err)
	}

	return nil
}

// GetCertificatePath returns the full path to a certificate file
// Path structure: /storage/{NIT}/certificates/{filename}
func (s *CertificateService) GetCertificatePath(filename string, companyNIT string) string {
	return s.storage.CertificatePath(companyNIT, filename)
}

// GetCertificateForSigning gets certificate data and password for signing operations
func (s *CertificateService) GetCertificateForSigning(companyID int64) (path string, password string, err error) {
	cert, err := s.certRepo.GetByCompanyID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("no active certificate found for company")
		}
		return "", "", err
	}

	// Decrypt password
	decryptedPassword, err := crypto.DecryptPassword(cert.Password)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt certificate password: %w", err)
	}

	// Get company to retrieve NIT for path
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get company: %w", err)
	}

	certPath := s.GetCertificatePath(cert.Name, company.NIT)

	// Verify file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", errors.New("certificate file not found on filesystem")
	}

	// Read certificate file
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Validate PKCS12 format and password when actually using the certificate
	// This validation happens here (not at upload) to support both BER and DER encodings
	if err := crypto.ValidatePKCS12(certData, decryptedPassword); err != nil {
		return "", "", fmt.Errorf("certificate validation failed: %w. The certificate may be corrupted or the password is incorrect. Please upload a new certificate", err)
	}

	return certPath, decryptedPassword, nil
}
