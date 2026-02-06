package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// SaveFile guarda un archivo en el sistema de archivos local
func (s *LocalStorage) SaveFile(file *multipart.FileHeader, subdir string) (string, error) {
	// Crear directorio si no existe
	fullPath := filepath.Join(s.basePath, subdir)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	// Generar nombre único para el archivo
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", timestamp, sanitizeFilename(file.Filename), ext)
	
	// Ruta completa del archivo
	filePath := filepath.Join(fullPath, filename)

	// Abrir archivo subido
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening uploaded file: %w", err)
	}
	defer src.Close()

	// Crear archivo destino
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer dst.Close()

	// Copiar contenido
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("error saving file: %w", err)
	}

	// Retornar ruta relativa
	return filepath.Join(subdir, filename), nil
}

// DeleteFile elimina un archivo del sistema
func (s *LocalStorage) DeleteFile(relativePath string) error {
	fullPath := filepath.Join(s.basePath, relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error deleting file: %w", err)
	}
	return nil
}

// FileExists verifica si un archivo existe
func (s *LocalStorage) FileExists(relativePath string) bool {
	fullPath := filepath.Join(s.basePath, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFilePath obtiene la ruta completa de un archivo
func (s *LocalStorage) GetFilePath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

// sanitizeFilename limpia el nombre del archivo
func sanitizeFilename(filename string) string {
	// Remover extensión
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Reemplazar caracteres no permitidos
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	return name
}
