package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"fmt"
)

type UserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create crea un nuevo usuario
func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (name, email, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.DB.QueryRow(
		query,
		user.Name,
		user.Email,
		user.Password,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID obtiene un usuario por ID
func (r *UserRepository) GetByID(id int64) (*domain.User, error) {
	query := `
		SELECT 
			id, name, email, email_verified_at, password, 
			is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`
	user := &domain.User{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.Password,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

// GetByEmail obtiene un usuario por email
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT 
			id, name, email, email_verified_at, password, 
			is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`
	user := &domain.User{}
	err := r.db.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.Password,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

// GetAll obtiene todos los usuarios activos con paginación
func (r *UserRepository) GetAll(page, pageSize int) ([]*domain.User, int, error) {
	offset := (page - 1) * pageSize
	
	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE is_active = true`
	err := r.db.DB.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query := `
		SELECT 
			id, name, email, email_verified_at, 
			is_active, created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.EmailVerifiedAt,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// Update actualiza un usuario
func (r *UserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5 AND is_active = true
		RETURNING updated_at
	`
	return r.db.DB.QueryRow(
		query,
		user.Name,
		user.Email,
		user.Password,
		user.IsActive,
		user.ID,
	).Scan(&user.UpdatedAt)
}

// Delete realiza soft delete de un usuario
func (r *UserRepository) Delete(id int64) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND is_active = true
	`
	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// EmailExists verifica si un email ya existe
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_active = true)`
	var exists bool
	err := r.db.DB.QueryRow(query, email).Scan(&exists)
	return exists, err
}
