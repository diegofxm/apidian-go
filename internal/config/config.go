package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Invoice  InvoiceConfig
}

type ServerConfig struct {
	Port         string
	Env          string
	AllowOrigins string
	Timezone     string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	Expiration int
}

type StorageConfig struct {
	Path string // Ruta base de storage (./storage)
}

type InvoiceConfig struct {
	KeepUnsignedXML bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	jwtExpiration, err := strconv.Atoi(getEnv("JWT_EXPIRATION", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "3000"),
			Env:          getEnv("APP_ENV", "development"),
			AllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
			Timezone:     getEnv("TZ", "America/Bogota"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "apidian"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: jwtExpiration,
		},
		Storage: StorageConfig{
			Path: getEnv("STORAGE_PATH", "./storage"),
		},
		Invoice: InvoiceConfig{
			KeepUnsignedXML: getEnvBool("KEEP_UNSIGNED_XML", false),
		},
	}, nil
}

// AppRoot retorna la ruta de storage/app donde se guardan todos los archivos
func (s StorageConfig) AppRoot() string {
	return filepath.Join(s.Path, "app")
}

// LogsRoot retorna la ruta de storage/logs
func (s StorageConfig) LogsRoot() string {
	return filepath.Join(s.Path, "logs")
}

// TempPath retorna la ruta de archivos temporales
func (s StorageConfig) TempPath() string {
	return filepath.Join(s.AppRoot(), "temp")
}

// TempUploadsPath retorna la ruta de uploads temporales
func (s StorageConfig) TempUploadsPath() string {
	return filepath.Join(s.TempPath(), "uploads")
}

// === Compañía ===

// CompanyPath retorna la ruta de una empresa específica
func (s StorageConfig) CompanyPath(nit string) string {
	return filepath.Join(s.AppRoot(), "companies", nit)
}

// CompanyProfilePath retorna la ruta del perfil de una empresa
func (s StorageConfig) CompanyProfilePath(nit string) string {
	return filepath.Join(s.CompanyPath(nit), "profile")
}

// CompanyLogoPath retorna la ruta del logo de una empresa
func (s StorageConfig) CompanyLogoPath(nit string) string {
	return filepath.Join(s.CompanyProfilePath(nit), "logo.png")
}

// === Certificados ===

// CertificatesPath retorna la ruta de certificados de una empresa
func (s StorageConfig) CertificatesPath(nit string) string {
	return filepath.Join(s.CompanyPath(nit), "certificates")
}

// CertificatePath retorna la ruta de un certificado específico
func (s StorageConfig) CertificatePath(nit, filename string) string {
	return filepath.Join(s.CertificatesPath(nit), filename)
}

// === Documentos ===

// DocumentsPath retorna la ruta de documentos de una empresa
func (s StorageConfig) DocumentsPath(nit string) string {
	return filepath.Join(s.CompanyPath(nit), "documents")
}

// InvoicesPath retorna la ruta de facturas de una empresa
func (s StorageConfig) InvoicesPath(nit string) string {
	return filepath.Join(s.DocumentsPath(nit), "invoices")
}

// InvoicePath retorna la ruta de una factura específica
func (s StorageConfig) InvoicePath(nit, numero string) string {
	return filepath.Join(s.InvoicesPath(nit), numero)
}

// InvoiceXMLPath retorna la ruta del XML sin firmar de una factura
func (s StorageConfig) InvoiceXMLPath(nit, numero string) string {
	return filepath.Join(s.InvoicePath(nit, numero), numero+".xml")
}

// InvoiceSignedXMLPath retorna la ruta del XML firmado de una factura
func (s StorageConfig) InvoiceSignedXMLPath(nit, numero string) string {
	return filepath.Join(s.InvoicePath(nit, numero), numero+"_signed.xml")
}

// InvoiceZIPPath retorna la ruta del ZIP de una factura
func (s StorageConfig) InvoiceZIPPath(nit, numero string) string {
	return filepath.Join(s.InvoicePath(nit, numero), numero+".zip")
}

// InvoiceApplicationResponsePath retorna la ruta del ApplicationResponse de una factura
func (s StorageConfig) InvoiceApplicationResponsePath(nit, numero string) string {
	return filepath.Join(s.InvoicePath(nit, numero), "ApplicationResponse-"+numero+".xml")
}

// === Logs ===

// DebugSoapPath retorna la ruta de debug SOAP
func (s StorageConfig) DebugSoapPath() string {
	return filepath.Join(s.LogsRoot(), "debug", "soap")
}

// === Assets (privados) ===

// AssetsPath retorna la ruta de assets de la aplicación (logo por defecto, etc)
func (s StorageConfig) AssetsPath() string {
	return filepath.Join(s.AppRoot(), "assets")
}

// DefaultLogoPath retorna la ruta del logo por defecto
func (s StorageConfig) DefaultLogoPath() string {
	return filepath.Join(s.AssetsPath(), "logo_default.png")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
