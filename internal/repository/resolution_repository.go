package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"errors"
	"time"
)

type ResolutionRepository struct {
	db *database.Database
}

func NewResolutionRepository(db *database.Database) *ResolutionRepository {
	return &ResolutionRepository{db: db}
}

// Create crea una nueva resolución
func (r *ResolutionRepository) Create(req *domain.CreateResolutionRequest) (*domain.Resolution, error) {
	query := `
		INSERT INTO resolutions (
			company_id, type_document_id, prefix, resolution, technical_key,
			from_number, to_number, current_number, date_from, date_to
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, is_active, created_at, updated_at
	`

	// Parsear fechas
	dateFrom, _ := time.Parse("2006-01-02", req.DateFrom)
	dateTo, _ := time.Parse("2006-01-02", req.DateTo)

	resolution := &domain.Resolution{
		CompanyID:      req.CompanyID,
		TypeDocumentID: req.TypeDocumentID,
		Prefix:         req.Prefix,
		Resolution:     req.Resolution,
		TechnicalKey:   req.TechnicalKey,
		FromNumber:     req.FromNumber,
		ToNumber:       req.ToNumber,
		CurrentNumber:  req.FromNumber, // Inicia en from_number
		DateFrom:       dateFrom,
		DateTo:         dateTo,
	}

	err := r.db.DB.QueryRow(
		query,
		req.CompanyID,
		req.TypeDocumentID,
		req.Prefix,
		req.Resolution,
		req.TechnicalKey,
		req.FromNumber,
		req.ToNumber,
		req.FromNumber, // current_number inicia en from_number
		dateFrom,
		dateTo,
	).Scan(
		&resolution.ID,
		&resolution.IsActive,
		&resolution.CreatedAt,
		&resolution.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return resolution, nil
}

// GetByID obtiene una resolución por ID
func (r *ResolutionRepository) GetByID(id int64) (*domain.Resolution, error) {
	query := `
		SELECT 
			id, company_id, type_document_id, prefix, resolution, technical_key,
			from_number, to_number, current_number, date_from, date_to,
			is_active, created_at, updated_at
		FROM resolutions
		WHERE id = $1 AND is_active = true
	`

	resolution := &domain.Resolution{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&resolution.ID,
		&resolution.CompanyID,
		&resolution.TypeDocumentID,
		&resolution.Prefix,
		&resolution.Resolution,
		&resolution.TechnicalKey,
		&resolution.FromNumber,
		&resolution.ToNumber,
		&resolution.CurrentNumber,
		&resolution.DateFrom,
		&resolution.DateTo,
		&resolution.IsActive,
		&resolution.CreatedAt,
		&resolution.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("resolution not found")
		}
		return nil, err
	}

	return resolution, nil
}

// GetByCompanyID obtiene las resoluciones de una empresa con paginación
func (r *ResolutionRepository) GetByCompanyID(companyID int64, page, pageSize int) (*domain.ResolutionListResponse, error) {
	offset := (page - 1) * pageSize

	// Contar total de resoluciones
	var total int
	countQuery := `SELECT COUNT(*) FROM resolutions WHERE company_id = $1 AND is_active = true`
	err := r.db.DB.QueryRow(countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Obtener resoluciones paginadas
	query := `
		SELECT 
			id, company_id, type_document_id, prefix, resolution, technical_key,
			from_number, to_number, current_number, date_from, date_to,
			is_active, created_at, updated_at
		FROM resolutions
		WHERE company_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, companyID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resolutions []domain.Resolution
	for rows.Next() {
		var resolution domain.Resolution
		err := rows.Scan(
			&resolution.ID,
			&resolution.CompanyID,
			&resolution.TypeDocumentID,
			&resolution.Prefix,
			&resolution.Resolution,
			&resolution.TechnicalKey,
			&resolution.FromNumber,
			&resolution.ToNumber,
			&resolution.CurrentNumber,
			&resolution.DateFrom,
			&resolution.DateTo,
			&resolution.IsActive,
			&resolution.CreatedAt,
			&resolution.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resolutions = append(resolutions, resolution)
	}

	return &domain.ResolutionListResponse{
		Resolutions: resolutions,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetByUserID obtiene las resoluciones de todas las empresas del usuario con paginación
func (r *ResolutionRepository) GetByUserID(userID int64, page, pageSize int) (*domain.ResolutionListResponse, error) {
	offset := (page - 1) * pageSize

	// Contar total de resoluciones
	var total int
	countQuery := `
		SELECT COUNT(*) 
		FROM resolutions res
		INNER JOIN companies c ON res.company_id = c.id
		WHERE c.user_id = $1 AND res.is_active = true AND c.is_active = true
	`
	err := r.db.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Obtener resoluciones paginadas
	query := `
		SELECT 
			res.id, res.company_id, res.type_document_id, res.prefix, res.resolution, 
			res.technical_key, res.from_number, res.to_number, res.current_number, 
			res.date_from, res.date_to, res.is_active, res.created_at, res.updated_at
		FROM resolutions res
		INNER JOIN companies c ON res.company_id = c.id
		WHERE c.user_id = $1 AND res.is_active = true AND c.is_active = true
		ORDER BY res.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resolutions []domain.Resolution
	for rows.Next() {
		var resolution domain.Resolution
		err := rows.Scan(
			&resolution.ID,
			&resolution.CompanyID,
			&resolution.TypeDocumentID,
			&resolution.Prefix,
			&resolution.Resolution,
			&resolution.TechnicalKey,
			&resolution.FromNumber,
			&resolution.ToNumber,
			&resolution.CurrentNumber,
			&resolution.DateFrom,
			&resolution.DateTo,
			&resolution.IsActive,
			&resolution.CreatedAt,
			&resolution.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resolutions = append(resolutions, resolution)
	}

	return &domain.ResolutionListResponse{
		Resolutions: resolutions,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// Delete elimina (soft delete) una resolución
func (r *ResolutionRepository) Delete(id int64) error {
	query := `
		UPDATE resolutions SET
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
		return errors.New("resolution not found")
	}

	return nil
}

// GetByCompanyAndPrefix obtiene una resolución por empresa y prefijo
func (r *ResolutionRepository) GetByCompanyAndPrefix(companyID int64, prefix string) (*domain.Resolution, error) {
	query := `
		SELECT 
			id, company_id, type_document_id, prefix, resolution, technical_key,
			from_number, to_number, current_number, date_from, date_to,
			is_active, created_at, updated_at
		FROM resolutions
		WHERE company_id = $1 AND prefix = $2 AND is_active = true
	`

	resolution := &domain.Resolution{}
	err := r.db.DB.QueryRow(query, companyID, prefix).Scan(
		&resolution.ID,
		&resolution.CompanyID,
		&resolution.TypeDocumentID,
		&resolution.Prefix,
		&resolution.Resolution,
		&resolution.TechnicalKey,
		&resolution.FromNumber,
		&resolution.ToNumber,
		&resolution.CurrentNumber,
		&resolution.DateFrom,
		&resolution.DateTo,
		&resolution.IsActive,
		&resolution.CreatedAt,
		&resolution.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No existe, pero no es error
		}
		return nil, err
	}

	return resolution, nil
}

// GetAndIncrementConsecutive obtiene el consecutivo actual y lo incrementa de forma atómica
// Este método es thread-safe y previene race conditions usando transacciones con FOR UPDATE
func (r *ResolutionRepository) GetAndIncrementConsecutive(resolutionID int64) (int64, error) {
	// Iniciar transacción
	tx, err := r.db.DB.Begin()
	if err != nil {
		return 0, errors.New("error starting transaction: " + err.Error())
	}
	defer tx.Rollback()

	// Obtener resolución y bloquear la fila (FOR UPDATE previene lecturas concurrentes)
	query := `
		SELECT current_number, to_number, is_active
		FROM resolutions
		WHERE id = $1
		FOR UPDATE
	`
	
	var currentNumber, toNumber int64
	var isActive bool
	err = tx.QueryRow(query, resolutionID).Scan(&currentNumber, &toNumber, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("resolution not found")
		}
		return 0, err
	}

	// Validar que la resolución esté activa
	if !isActive {
		return 0, errors.New("resolution is not active")
	}

	// Validar que hay consecutivos disponibles
	if currentNumber >= toNumber {
		return 0, errors.New("no hay consecutivos disponibles en esta resolución")
	}

	// Incrementar current_number en la base de datos
	updateQuery := `
		UPDATE resolutions 
		SET current_number = current_number + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(updateQuery, resolutionID)
	if err != nil {
		return 0, errors.New("error incrementing consecutive: " + err.Error())
	}

	// Commit de la transacción
	if err = tx.Commit(); err != nil {
		return 0, errors.New("error committing transaction: " + err.Error())
	}

	// Retornar el consecutivo que se usó (antes del incremento)
	return currentNumber, nil
}

// Placeholder para evitar error de compilación
type _ domain.Resolution
type _ database.Database
