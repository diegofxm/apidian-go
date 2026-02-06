package service

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/repository"
	"apidian-go/pkg/utils"
	"fmt"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create crea un nuevo producto
func (s *ProductService) Create(userID int64, req *domain.CreateProductRequest) (*domain.Product, error) {
	// Validar que no exista producto con el mismo código en la empresa
	existing, err := s.repo.GetByCode(req.CompanyID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("error checking existing product: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("product with code %s already exists for this company", req.Code)
	}

	// Crear producto
	product, err := s.repo.Create(userID, req)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// GetByID obtiene un producto por ID
func (s *ProductService) GetByID(id int64, companyID int64) (*domain.Product, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && product.CompanyID != companyID {
		return nil, fmt.Errorf("unauthorized access to product")
	}

	return product, nil
}

// GetByCompanyID obtiene todos los productos de una empresa con paginación
func (s *ProductService) GetByCompanyID(companyID int64, page, pageSize int) (*domain.ProductListResponse, error) {
	// Normalizar paginación
	page, pageSize = utils.NormalizePagination(page, pageSize)

	products, total, err := s.repo.GetByCompanyID(companyID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &domain.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update actualiza un producto
func (s *ProductService) Update(id int64, companyID int64, req *domain.UpdateProductRequest) error {
	// Verificar que el producto existe
	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && product.CompanyID != companyID {
		return fmt.Errorf("unauthorized access to product")
	}

	// Actualizar producto
	if err := s.repo.Update(id, req); err != nil {
		return err
	}

	return nil
}

// Delete elimina (soft delete) un producto
func (s *ProductService) Delete(id int64, companyID int64) error {
	// Verificar que el producto existe
	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Si se proporciona companyID, verificar pertenencia
	if companyID > 0 && product.CompanyID != companyID {
		return fmt.Errorf("unauthorized access to product")
	}

	// Eliminar producto
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	return nil
}
