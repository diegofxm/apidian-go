package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type ProductRepository struct {
	db *database.Database
}

func NewProductRepository(db *database.Database) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create crea un nuevo producto
func (r *ProductRepository) Create(userID int64, req *domain.CreateProductRequest) (*domain.Product, error) {
	query := `
		INSERT INTO products (
			company_id, code, name, description, type_item_identification_id,
			standard_item_code, unspsc_code, unit_code_id, price, tax_type_id,
			tax_rate, brand_name, model_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, description, type_item_identification_id, standard_item_code, 
		            unspsc_code, brand_name, model_name, is_active, created_at, updated_at
	`

	product := &domain.Product{
		CompanyID:                req.CompanyID,
		Code:                     req.Code,
		Name:                     req.Name,
		Description:              req.Description,
		TypeItemIdentificationID: req.TypeItemIdentificationID,
		StandardItemCode:         req.StandardItemCode,
		UNSPSCCode:               req.UNSPSCCode,
		UnitCodeID:               req.UnitCodeID,
		Price:                    req.Price,
		TaxTypeID:                req.TaxTypeID,
		TaxRate:                  req.TaxRate,
		BrandName:                req.BrandName,
		ModelName:                req.ModelName,
	}

	err := r.db.DB.QueryRow(
		query,
		req.CompanyID,
		req.Code,
		req.Name,
		req.Description,
		req.TypeItemIdentificationID,
		req.StandardItemCode,
		req.UNSPSCCode,
		req.UnitCodeID,
		req.Price,
		req.TaxTypeID,
		req.TaxRate,
		req.BrandName,
		req.ModelName,
	).Scan(
		&product.ID,
		&product.Description,
		&product.TypeItemIdentificationID,
		&product.StandardItemCode,
		&product.UNSPSCCode,
		&product.BrandName,
		&product.ModelName,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("product with code %s already exists for this company", req.Code)
		}
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	return product, nil
}

// GetByID obtiene un producto por ID
func (r *ProductRepository) GetByID(id int64) (*domain.Product, error) {
	query := `
		SELECT 
			id, company_id, code, name, description, type_item_identification_id,
			standard_item_code, unspsc_code, unit_code_id, price, tax_type_id,
			tax_rate, brand_name, model_name, is_active, created_at, updated_at
		FROM products
		WHERE id = $1 AND is_active = true
	`

	product := &domain.Product{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&product.ID,
		&product.CompanyID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.TypeItemIdentificationID,
		&product.StandardItemCode,
		&product.UNSPSCCode,
		&product.UnitCodeID,
		&product.Price,
		&product.TaxTypeID,
		&product.TaxRate,
		&product.BrandName,
		&product.ModelName,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}

	return product, nil
}

// GetByCompanyID obtiene todos los productos de una empresa
func (r *ProductRepository) GetByCompanyID(companyID int64, page, pageSize int) ([]domain.Product, int, error) {
	offset := (page - 1) * pageSize

	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM products WHERE company_id = $1 AND is_active = true`
	err := r.db.DB.QueryRow(countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	// Obtener productos
	query := `
		SELECT 
			id, company_id, code, name, description, type_item_identification_id,
			standard_item_code, unspsc_code, unit_code_id, price, tax_type_id,
			tax_rate, brand_name, model_name, is_active, created_at, updated_at
		FROM products
		WHERE company_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, companyID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products := []domain.Product{}
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.CompanyID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.TypeItemIdentificationID,
			&product.StandardItemCode,
			&product.UNSPSCCode,
			&product.UnitCodeID,
			&product.Price,
			&product.TaxTypeID,
			&product.TaxRate,
			&product.BrandName,
			&product.ModelName,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, product)
	}

	return products, total, nil
}

// GetByUserID obtiene todos los productos de todas las empresas del usuario
func (r *ProductRepository) GetByUserID(userID int64, page, pageSize int) ([]domain.Product, int, error) {
	offset := (page - 1) * pageSize

	// Contar total
	var total int
	countQuery := `
		SELECT COUNT(*) 
		FROM products p
		INNER JOIN companies c ON p.company_id = c.id
		WHERE c.user_id = $1 AND p.is_active = true AND c.is_active = true
	`
	err := r.db.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	// Obtener productos
	query := `
		SELECT 
			p.id, p.company_id, p.code, p.name, p.description, p.type_item_identification_id,
			p.standard_item_code, p.unspsc_code, p.unit_code_id, p.price, p.tax_type_id,
			p.tax_rate, p.brand_name, p.model_name, p.is_active, p.created_at, p.updated_at
		FROM products p
		INNER JOIN companies c ON p.company_id = c.id
		WHERE c.user_id = $1 AND p.is_active = true AND c.is_active = true
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products := []domain.Product{}
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.CompanyID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.TypeItemIdentificationID,
			&product.StandardItemCode,
			&product.UNSPSCCode,
			&product.UnitCodeID,
			&product.Price,
			&product.TaxTypeID,
			&product.TaxRate,
			&product.BrandName,
			&product.ModelName,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, product)
	}

	return products, total, nil
}

// Update actualiza un producto
func (r *ProductRepository) Update(id int64, req *domain.UpdateProductRequest) error {
	query := `
		UPDATE products SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			type_item_identification_id = COALESCE($3, type_item_identification_id),
			standard_item_code = COALESCE($4, standard_item_code),
			unspsc_code = COALESCE($5, unspsc_code),
			unit_code_id = COALESCE($6, unit_code_id),
			price = COALESCE($7, price),
			tax_type_id = COALESCE($8, tax_type_id),
			tax_rate = COALESCE($9, tax_rate),
			brand_name = COALESCE($10, brand_name),
			model_name = COALESCE($11, model_name),
			is_active = COALESCE($12, is_active),
			updated_at = NOW()
		WHERE id = $13
	`

	result, err := r.db.DB.Exec(
		query,
		req.Name,
		req.Description,
		req.TypeItemIdentificationID,
		req.StandardItemCode,
		req.UNSPSCCode,
		req.UnitCodeID,
		req.Price,
		req.TaxTypeID,
		req.TaxRate,
		req.BrandName,
		req.ModelName,
		req.IsActive,
		id,
	)

	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Delete elimina (soft delete) un producto
func (r *ProductRepository) Delete(id int64) error {
	query := `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetByCode obtiene un producto por código
func (r *ProductRepository) GetByCode(companyID int64, code string) (*domain.Product, error) {
	query := `
		SELECT 
			id, company_id, code, name, description, type_item_identification_id,
			standard_item_code, unspsc_code, unit_code_id, price, tax_type_id,
			tax_rate, brand_name, model_name, is_active, created_at, updated_at
		FROM products
		WHERE company_id = $1 AND code = $2 AND is_active = true
	`

	product := &domain.Product{}
	err := r.db.DB.QueryRow(query, companyID, code).Scan(
		&product.ID,
		&product.CompanyID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.TypeItemIdentificationID,
		&product.StandardItemCode,
		&product.UNSPSCCode,
		&product.UnitCodeID,
		&product.Price,
		&product.TaxTypeID,
		&product.TaxRate,
		&product.BrandName,
		&product.ModelName,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting product by code: %w", err)
	}

	return product, nil
}
