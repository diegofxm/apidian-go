package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type CompanyRepository struct {
	db *database.Database
}

func NewCompanyRepository(db *database.Database) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// Create crea una nueva empresa
func (r *CompanyRepository) Create(userID int64, req *domain.CreateCompanyRequest) (*domain.Company, error) {
	query := `
		INSERT INTO companies (
			user_id, document_type_id, nit, dv, name, trade_name, registration_name,
			tax_level_code_id, tax_type_id, type_organization_id, type_regime_id, industry_codes,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, website, logo_path
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		) RETURNING id, dv, trade_name, postal_zone, phone, email, website, logo_path,
		            is_active, created_at, updated_at
	`

	company := &domain.Company{
		UserID:             userID,
		DocumentTypeID:     req.DocumentTypeID,
		NIT:                req.NIT,
		DV:                 req.DV,
		Name:               req.Name,
		TradeName:          req.TradeName,
		RegistrationName:   req.RegistrationName,
		TaxLevelCodeID:     req.TaxLevelCodeID,
		TaxTypeID:          req.TaxTypeID,
		TypeOrganizationID: req.TypeOrganizationID,
		TypeRegimeID:       req.TypeRegimeID,
		IndustryCodes:      req.IndustryCodes,
		CountryID:          req.CountryID,
		DepartmentID:       req.DepartmentID,
		MunicipalityID:     req.MunicipalityID,
		AddressLine:        req.AddressLine,
		PostalZone:         req.PostalZone,
		Phone:              req.Phone,
		Email:              req.Email,
		Website:            req.Website,
		LogoPath:           req.LogoPath,
	}

	err := r.db.DB.QueryRow(
		query,
		userID,
		req.DocumentTypeID,
		req.NIT,
		req.DV,
		req.Name,
		req.TradeName,
		req.RegistrationName,
		req.TaxLevelCodeID,
		req.TaxTypeID,
		req.TypeOrganizationID,
		req.TypeRegimeID,
		pq.Array(&req.IndustryCodes),
		req.CountryID,
		req.DepartmentID,
		req.MunicipalityID,
		req.AddressLine,
		req.PostalZone,
		req.Phone,
		req.Email,
		req.Website,
		req.LogoPath,
	).Scan(
		&company.ID,
		&company.DV,
		&company.TradeName,
		&company.PostalZone,
		&company.Phone,
		&company.Email,
		&company.Website,
		&company.LogoPath,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("company with NIT %s already exists", req.NIT)
		}
		return nil, fmt.Errorf("error creating company: %w", err)
	}

	return company, nil
}

// GetByID obtiene una empresa por ID
func (r *CompanyRepository) GetByID(id int64) (*domain.Company, error) {
	query := `
		SELECT 
			id, user_id, document_type_id, nit, dv, name, trade_name, registration_name,
			tax_level_code_id, type_organization_id, type_regime_id, industry_codes,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, website, logo_path, is_active,
			created_at, updated_at
		FROM companies
		WHERE id = $1 AND is_active = true
	`

	company := &domain.Company{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&company.ID,
		&company.UserID,
		&company.DocumentTypeID,
		&company.NIT,
		&company.DV,
		&company.Name,
		&company.TradeName,
		&company.RegistrationName,
		&company.TaxLevelCodeID,
		&company.TypeOrganizationID,
		&company.TypeRegimeID,
		pq.Array(&company.IndustryCodes),
		&company.CountryID,
		&company.DepartmentID,
		&company.MunicipalityID,
		&company.AddressLine,
		&company.PostalZone,
		&company.Phone,
		&company.Email,
		&company.Website,
		&company.LogoPath,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("company not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting company: %w", err)
	}

	return company, nil
}

// GetByUserID obtiene todas las empresas de un usuario
func (r *CompanyRepository) GetByUserID(userID int64, page, pageSize int) ([]domain.Company, int, error) {
	offset := (page - 1) * pageSize

	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM companies WHERE user_id = $1 AND is_active = true`
	err := r.db.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting companies: %w", err)
	}

	// Obtener empresas
	query := `
		SELECT 
			id, user_id, document_type_id, nit, dv, name, trade_name, registration_name,
			tax_level_code_id, type_organization_id, type_regime_id, industry_codes,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, website, logo_path, is_active,
			created_at, updated_at
		FROM companies
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying companies: %w", err)
	}
	defer rows.Close()

	companies := []domain.Company{}
	for rows.Next() {
		var company domain.Company
		err := rows.Scan(
			&company.ID,
			&company.UserID,
			&company.DocumentTypeID,
			&company.NIT,
			&company.DV,
			&company.Name,
			&company.TradeName,
			&company.RegistrationName,
			&company.TaxLevelCodeID,
			&company.TypeOrganizationID,
			&company.TypeRegimeID,
			pq.Array(&company.IndustryCodes),
			&company.CountryID,
			&company.DepartmentID,
			&company.MunicipalityID,
			&company.AddressLine,
			&company.PostalZone,
			&company.Phone,
			&company.Email,
			&company.Website,
			&company.LogoPath,
			&company.IsActive,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning company: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, total, nil
}

// Update actualiza una empresa
func (r *CompanyRepository) Update(id int64, req *domain.UpdateCompanyRequest) error {
	query := `
		UPDATE companies SET
			name = COALESCE($1, name),
			trade_name = COALESCE($2, trade_name),
			registration_name = COALESCE($3, registration_name),
			tax_level_code_id = COALESCE($4, tax_level_code_id),
			tax_type_id = COALESCE($5, tax_type_id),
			type_organization_id = COALESCE($6, type_organization_id),
			type_regime_id = COALESCE($7, type_regime_id),
			industry_codes = COALESCE($8, industry_codes),
			department_id = COALESCE($9, department_id),
			municipality_id = COALESCE($10, municipality_id),
			address_line = COALESCE($11, address_line),
			postal_zone = COALESCE($12, postal_zone),
			phone = COALESCE($13, phone),
			email = COALESCE($14, email),
			website = COALESCE($15, website),
			logo_path = COALESCE($16, logo_path),
			is_active = COALESCE($17, is_active),
			updated_at = NOW()
		WHERE id = $18
	`

	result, err := r.db.DB.Exec(
		query,
		req.Name,
		req.TradeName,
		req.RegistrationName,
		req.TaxLevelCodeID,
		req.TaxTypeID,
		req.TypeOrganizationID,
		req.TypeRegimeID,
		pq.Array(&req.IndustryCodes),
		req.DepartmentID,
		req.MunicipalityID,
		req.AddressLine,
		req.PostalZone,
		req.Phone,
		req.Email,
		req.Website,
		req.LogoPath,
		req.IsActive,
		id,
	)

	if err != nil {
		return fmt.Errorf("error updating company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// Delete elimina (soft delete) una empresa
func (r *CompanyRepository) Delete(id int64) error {
	query := `UPDATE companies SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// GetByNIT obtiene una empresa por NIT y DV
func (r *CompanyRepository) GetByNIT(nit, dv string) (*domain.Company, error) {
	query := `
		SELECT 
			id, user_id, document_type_id, nit, dv, name, trade_name, registration_name,
			tax_level_code_id, type_organization_id, type_regime_id, industry_codes,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, website, logo_path, is_active,
			created_at, updated_at
		FROM companies
		WHERE nit = $1 AND dv = $2 AND is_active = true
	`

	company := &domain.Company{}
	err := r.db.DB.QueryRow(query, nit, dv).Scan(
		&company.ID,
		&company.UserID,
		&company.DocumentTypeID,
		&company.NIT,
		&company.DV,
		&company.Name,
		&company.TradeName,
		&company.RegistrationName,
		&company.TaxLevelCodeID,
		&company.TypeOrganizationID,
		&company.TypeRegimeID,
		pq.Array(&company.IndustryCodes),
		&company.CountryID,
		&company.DepartmentID,
		&company.MunicipalityID,
		&company.AddressLine,
		&company.PostalZone,
		&company.Phone,
		&company.Email,
		&company.Website,
		&company.LogoPath,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("company not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting company by NIT: %w", err)
	}

	return company, nil
}
