package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type CustomerRepository struct {
	db *database.Database
}

func NewCustomerRepository(db *database.Database) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create crea un nuevo cliente
func (r *CustomerRepository) Create(userID int64, req *domain.CreateCustomerRequest) (*domain.Customer, error) {
	query := `
		INSERT INTO customers (
			company_id, document_type_id, identification_number, dv, name, trade_name,
			tax_level_code_id, tax_type_id, type_organization_id, type_regime_id,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id, dv, trade_name, postal_zone, phone, email, is_active, created_at, updated_at
	`

	customer := &domain.Customer{
		CompanyID:            req.CompanyID,
		DocumentTypeID:       req.DocumentTypeID,
		IdentificationNumber: req.IdentificationNumber,
		DV:                   req.DV,
		Name:                 req.Name,
		TradeName:            req.TradeName,
		TaxLevelCodeID:       req.TaxLevelCodeID,
		TaxTypeID:            req.TaxTypeID,
		TypeOrganizationID:   req.TypeOrganizationID,
		TypeRegimeID:         req.TypeRegimeID,
		CountryID:            req.CountryID,
		DepartmentID:         req.DepartmentID,
		MunicipalityID:       req.MunicipalityID,
		AddressLine:          req.AddressLine,
		PostalZone:           req.PostalZone,
		Phone:                req.Phone,
		Email:                req.Email,
	}

	err := r.db.DB.QueryRow(
		query,
		req.CompanyID,
		req.DocumentTypeID,
		req.IdentificationNumber,
		req.DV,
		req.Name,
		req.TradeName,
		req.TaxLevelCodeID,
		req.TaxTypeID,
		req.TypeOrganizationID,
		req.TypeRegimeID,
		req.CountryID,
		req.DepartmentID,
		req.MunicipalityID,
		req.AddressLine,
		req.PostalZone,
		req.Phone,
		req.Email,
	).Scan(
		&customer.ID,
		&customer.DV,
		&customer.TradeName,
		&customer.PostalZone,
		&customer.Phone,
		&customer.Email,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("customer with identification %s already exists for this company", req.IdentificationNumber)
		}
		return nil, fmt.Errorf("error creating customer: %w", err)
	}

	return customer, nil
}

// GetByID obtiene un cliente por ID
func (r *CustomerRepository) GetByID(id int64) (*domain.Customer, error) {
	query := `
		SELECT 
			id, company_id, document_type_id, identification_number, dv, name, trade_name,
			tax_level_code_id, type_organization_id, type_regime_id,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, is_active, created_at, updated_at
		FROM customers
		WHERE id = $1 AND is_active = true
	`

	customer := &domain.Customer{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&customer.ID,
		&customer.CompanyID,
		&customer.DocumentTypeID,
		&customer.IdentificationNumber,
		&customer.DV,
		&customer.Name,
		&customer.TradeName,
		&customer.TaxLevelCodeID,
		&customer.TypeOrganizationID,
		&customer.TypeRegimeID,
		&customer.CountryID,
		&customer.DepartmentID,
		&customer.MunicipalityID,
		&customer.AddressLine,
		&customer.PostalZone,
		&customer.Phone,
		&customer.Email,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting customer: %w", err)
	}

	return customer, nil
}

// GetByCompanyID obtiene todos los clientes de una empresa
func (r *CustomerRepository) GetByCompanyID(companyID int64, page, pageSize int) ([]domain.Customer, int, error) {
	offset := (page - 1) * pageSize

	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM customers WHERE company_id = $1 AND is_active = true`
	err := r.db.DB.QueryRow(countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting customers: %w", err)
	}

	// Obtener clientes
	query := `
		SELECT 
			id, company_id, document_type_id, identification_number, dv, name, trade_name,
			tax_level_code_id, type_organization_id, type_regime_id,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, is_active, created_at, updated_at
		FROM customers
		WHERE company_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, companyID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying customers: %w", err)
	}
	defer rows.Close()

	customers := []domain.Customer{}
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.CompanyID,
			&customer.DocumentTypeID,
			&customer.IdentificationNumber,
			&customer.DV,
			&customer.Name,
			&customer.TradeName,
			&customer.TaxLevelCodeID,
			&customer.TypeOrganizationID,
			&customer.TypeRegimeID,
			&customer.CountryID,
			&customer.DepartmentID,
			&customer.MunicipalityID,
			&customer.AddressLine,
			&customer.PostalZone,
			&customer.Phone,
			&customer.Email,
			&customer.IsActive,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning customer: %w", err)
		}
		customers = append(customers, customer)
	}

	return customers, total, nil
}

// GetByUserID obtiene todos los clientes de todas las empresas del usuario
func (r *CustomerRepository) GetByUserID(userID int64, page, pageSize int) ([]domain.Customer, int, error) {
	offset := (page - 1) * pageSize

	// Contar total
	var total int
	countQuery := `
		SELECT COUNT(*) 
		FROM customers c
		INNER JOIN companies co ON c.company_id = co.id
		WHERE co.user_id = $1 AND c.is_active = true AND co.is_active = true
	`
	err := r.db.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting customers: %w", err)
	}

	// Obtener clientes
	query := `
		SELECT 
			c.id, c.company_id, c.document_type_id, c.identification_number, c.dv, c.name, c.trade_name,
			c.tax_level_code_id, c.type_organization_id, c.type_regime_id,
			c.country_id, c.department_id, c.municipality_id, c.address_line, c.postal_zone,
			c.phone, c.email, c.is_active, c.created_at, c.updated_at
		FROM customers c
		INNER JOIN companies co ON c.company_id = co.id
		WHERE co.user_id = $1 AND c.is_active = true AND co.is_active = true
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying customers: %w", err)
	}
	defer rows.Close()

	customers := []domain.Customer{}
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.CompanyID,
			&customer.DocumentTypeID,
			&customer.IdentificationNumber,
			&customer.DV,
			&customer.Name,
			&customer.TradeName,
			&customer.TaxLevelCodeID,
			&customer.TypeOrganizationID,
			&customer.TypeRegimeID,
			&customer.CountryID,
			&customer.DepartmentID,
			&customer.MunicipalityID,
			&customer.AddressLine,
			&customer.PostalZone,
			&customer.Phone,
			&customer.Email,
			&customer.IsActive,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning customer: %w", err)
		}
		customers = append(customers, customer)
	}

	return customers, total, nil
}

// Update actualiza un cliente
func (r *CustomerRepository) Update(id int64, req *domain.UpdateCustomerRequest) error {
	query := `
		UPDATE customers SET
			name = COALESCE($1, name),
			trade_name = COALESCE($2, trade_name),
			tax_level_code_id = COALESCE($3, tax_level_code_id),
			tax_type_id = COALESCE($4, tax_type_id),
			type_organization_id = COALESCE($5, type_organization_id),
			type_regime_id = COALESCE($6, type_regime_id),
			department_id = COALESCE($7, department_id),
			municipality_id = COALESCE($8, municipality_id),
			address_line = COALESCE($9, address_line),
			postal_zone = COALESCE($10, postal_zone),
			phone = COALESCE($11, phone),
			email = COALESCE($12, email),
			is_active = COALESCE($13, is_active),
			updated_at = NOW()
		WHERE id = $14
	`

	result, err := r.db.DB.Exec(
		query,
		req.Name,
		req.TradeName,
		req.TaxLevelCodeID,
		req.TaxTypeID,
		req.TypeOrganizationID,
		req.TypeRegimeID,
		req.DepartmentID,
		req.MunicipalityID,
		req.AddressLine,
		req.PostalZone,
		req.Phone,
		req.Email,
		req.IsActive,
		id,
	)

	if err != nil {
		return fmt.Errorf("error updating customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

// Delete elimina (soft delete) un cliente
func (r *CustomerRepository) Delete(id int64) error {
	query := `UPDATE customers SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

// GetByIdentification obtiene un cliente por número de identificación
func (r *CustomerRepository) GetByIdentification(companyID int64, identification string) (*domain.Customer, error) {
	query := `
		SELECT 
			id, company_id, document_type_id, identification_number, dv, name, trade_name,
			tax_level_code_id, type_organization_id, type_regime_id,
			country_id, department_id, municipality_id, address_line, postal_zone,
			phone, email, is_active, created_at, updated_at
		FROM customers
		WHERE company_id = $1 AND identification_number = $2 AND is_active = true
	`

	customer := &domain.Customer{}
	err := r.db.DB.QueryRow(query, companyID, identification).Scan(
		&customer.ID,
		&customer.CompanyID,
		&customer.DocumentTypeID,
		&customer.IdentificationNumber,
		&customer.DV,
		&customer.Name,
		&customer.TradeName,
		&customer.TaxLevelCodeID,
		&customer.TypeOrganizationID,
		&customer.TypeRegimeID,
		&customer.CountryID,
		&customer.DepartmentID,
		&customer.MunicipalityID,
		&customer.AddressLine,
		&customer.PostalZone,
		&customer.Phone,
		&customer.Email,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting customer by identification: %w", err)
	}

	return customer, nil
}
