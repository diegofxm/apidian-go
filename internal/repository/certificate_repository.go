package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"fmt"
)

type CertificateRepository struct {
	db *database.Database
}

func NewCertificateRepository(db *database.Database) *CertificateRepository {
	return &CertificateRepository{db: db}
}

// Create creates a new certificate
func (r *CertificateRepository) Create(cert *domain.Certificate) (*domain.Certificate, error) {
	query := `
		INSERT INTO certificates (company_id, name, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, company_id, name, password, is_active, created_at, updated_at
	`

	err := r.db.DB.QueryRow(
		query,
		cert.CompanyID,
		cert.Name,
		cert.Password,
		cert.IsActive,
	).Scan(
		&cert.ID,
		&cert.CompanyID,
		&cert.Name,
		&cert.Password,
		&cert.IsActive,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)

	return cert, err
}

// GetByCompanyID gets the active certificate for a company
func (r *CertificateRepository) GetByCompanyID(companyID int64) (*domain.Certificate, error) {
	query := `
		SELECT id, company_id, name, password, is_active, created_at, updated_at
		FROM certificates
		WHERE company_id = $1 AND is_active = true
		LIMIT 1
	`

	cert := &domain.Certificate{}
	err := r.db.DB.QueryRow(query, companyID).Scan(
		&cert.ID,
		&cert.CompanyID,
		&cert.Name,
		&cert.Password,
		&cert.IsActive,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)

	return cert, err
}

// GetByID gets a certificate by ID
func (r *CertificateRepository) GetByID(id int64) (*domain.Certificate, error) {
	query := `
		SELECT id, company_id, name, password, is_active, created_at, updated_at
		FROM certificates
		WHERE id = $1
	`

	cert := &domain.Certificate{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&cert.ID,
		&cert.CompanyID,
		&cert.Name,
		&cert.Password,
		&cert.IsActive,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)

	return cert, err
}

// GetAllByCompanyID gets all certificates for a company (including inactive)
func (r *CertificateRepository) GetAllByCompanyIDIncludingInactive(companyID int64) ([]*domain.Certificate, error) {
	query := `
		SELECT id, company_id, name, password, is_active, created_at, updated_at
		FROM certificates
		WHERE company_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certificates []*domain.Certificate
	for rows.Next() {
		cert := &domain.Certificate{}
		err := rows.Scan(
			&cert.ID,
			&cert.CompanyID,
			&cert.Name,
			&cert.Password,
			&cert.IsActive,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}

	return certificates, rows.Err()
}

// DeleteAllByCompanyID permanently deletes all certificates for a company
// This is used before uploading a new certificate to keep only one active certificate
func (r *CertificateRepository) DeleteAllByCompanyID(companyID int64) error {
	query := `DELETE FROM certificates WHERE company_id = $1`

	result, err := r.db.DB.Exec(query, companyID)
	if err != nil {
		return err
	}

	// Log how many rows were deleted (for debugging)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d certificate(s) for company %d\n", rowsAffected, companyID)
	}

	return nil
}

// Delete soft deletes a certificate (sets is_active to false)
func (r *CertificateRepository) Delete(id int64) error {
	query := `
		UPDATE certificates
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.DB.Exec(query, id)
	return err
}

// GetAllByCompanyID gets all certificates for a company (including inactive)
func (r *CertificateRepository) GetAllByCompanyID(companyID int64) ([]domain.Certificate, error) {
	query := `
		SELECT id, company_id, name, password, is_active, created_at, updated_at
		FROM certificates
		WHERE company_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certificates []domain.Certificate
	for rows.Next() {
		var cert domain.Certificate
		err := rows.Scan(
			&cert.ID,
			&cert.CompanyID,
			&cert.Name,
			&cert.Password,
			&cert.IsActive,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}

	return certificates, rows.Err()
}
