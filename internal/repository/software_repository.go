package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"errors"
)

type SoftwareRepository struct {
	db *database.Database
}

func NewSoftwareRepository(db *database.Database) *SoftwareRepository {
	return &SoftwareRepository{db: db}
}

// Create crea una nueva configuración de software
func (r *SoftwareRepository) Create(req *domain.CreateSoftwareRequest) (*domain.Software, error) {
	query := `
		INSERT INTO software (
			company_id, identifier, pin, environment, test_set_id
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id, is_active, created_at, updated_at
	`

	software := &domain.Software{
		CompanyID:   req.CompanyID,
		Identifier:  req.Identifier,
		Pin:         req.Pin,
		Environment: req.Environment,
		TestSetID:   req.TestSetID,
	}

	err := r.db.DB.QueryRow(
		query,
		req.CompanyID,
		req.Identifier,
		req.Pin,
		req.Environment,
		req.TestSetID,
	).Scan(
		&software.ID,
		&software.IsActive,
		&software.CreatedAt,
		&software.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return software, nil
}

// GetByID obtiene un software por ID
func (r *SoftwareRepository) GetByID(id int64) (*domain.Software, error) {
	query := `
		SELECT 
			id, company_id, identifier, pin, environment, test_set_id,
			is_active, created_at, updated_at
		FROM software
		WHERE id = $1 AND is_active = true
	`

	software := &domain.Software{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&software.ID,
		&software.CompanyID,
		&software.Identifier,
		&software.Pin,
		&software.Environment,
		&software.TestSetID,
		&software.IsActive,
		&software.CreatedAt,
		&software.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("software not found")
		}
		return nil, err
	}

	return software, nil
}

// GetByCompanyID obtiene el software de una empresa
func (r *SoftwareRepository) GetByCompanyID(companyID int64) (*domain.Software, error) {
	query := `
		SELECT 
			id, company_id, identifier, pin, environment, test_set_id,
			is_active, created_at, updated_at
		FROM software
		WHERE company_id = $1 AND is_active = true
	`

	software := &domain.Software{}
	err := r.db.DB.QueryRow(query, companyID).Scan(
		&software.ID,
		&software.CompanyID,
		&software.Identifier,
		&software.Pin,
		&software.Environment,
		&software.TestSetID,
		&software.IsActive,
		&software.CreatedAt,
		&software.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("software not found")
		}
		return nil, err
	}

	return software, nil
}

// Update actualiza un software
func (r *SoftwareRepository) Update(id int64, req *domain.UpdateSoftwareRequest) error {
	query := `
		UPDATE software SET
			identifier = COALESCE($1, identifier),
			pin = COALESCE($2, pin),
			environment = COALESCE($3, environment),
			test_set_id = COALESCE($4, test_set_id),
			is_active = COALESCE($5, is_active),
			updated_at = NOW()
		WHERE id = $6 AND is_active = true
	`

	result, err := r.db.DB.Exec(
		query,
		req.Identifier,
		req.Pin,
		req.Environment,
		req.TestSetID,
		req.IsActive,
		id,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("software not found")
	}

	return nil
}

// Delete elimina (soft delete) un software
func (r *SoftwareRepository) Delete(id int64) error {
	query := `
		UPDATE software SET
			is_active = false,
			updated_at = NOW()
		WHERE id = $1 AND is_active = true
	`

	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("software not found")
	}

	return nil
}
