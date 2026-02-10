# 🏛️ ARQUITECTURA HEXAGONAL COMPLETA - APIDIAN-GO

## 📋 DOCUMENTO DE DISEÑO DESDE CERO

**Versión**: 2.0  
**Fecha**: 2026-02-07  
**Tipo**: API REST con Arquitectura Hexagonal Pura  
**Documentos DIAN**: 6 tipos completos (Invoice, CreditNote, DebitNote, ApplicationResponse, AttachedDocument, Payroll)

---

## 🎯 PRINCIPIOS DE DISEÑO

### Premisas del Proyecto

1. ✅ **API REST** - Mantiene la categoría actual (Fiber)
2. ✅ **Arquitectura Hexagonal Pura** - Desde el día 1
3. ✅ **Módulos Externos en GitHub**:
   - `github.com/diegofxm/ubl21-dian` (Librería UBL 2.1)
   - `github.com/diegofxm/go-yaml-migrator` (Motor de migraciones)
   - `github.com/diegofxm/apidian-fixtures` (Datos ficticios/seeds para testing)
4. ✅ **Local solo datos del proyecto**:
   - `database/migrations/*.yaml` (DDL específico del proyecto)
   - `database/seeds/*.csv` (Catálogos DIAN reales)
5. ✅ **6 Documentos DIAN completos**

---

## 📐 ESTRUCTURA COMPLETA DEL PROYECTO

Ver archivo adjunto: `ESTRUCTURA_CARPETAS.md`

**Resumen de estructura**:
```
apidian-go/
├── cmd/                     # Puntos de entrada (api, worker)
├── database/                # Migraciones YAML + Seeds CSV (local)
├── internal/
│   ├── domain/              # Núcleo hexagonal
│   │   ├── entities/        # Invoice, CreditNote, DebitNote, etc.
│   │   ├── valueobjects/    # Money, Tax, NIT, CUFE, etc.
│   │   ├── aggregates/      # InvoiceAggregate, etc.
│   │   ├── events/          # Domain events
│   │   └── ports/           # Interfaces (input + output)
│   ├── application/         # Casos de uso (6 documentos × 5-6 casos)
│   └── adapters/
│       ├── input/http/      # REST API (Fiber)
│       └── output/          # Postgres, DIAN, Storage, PDF, Email, QR
├── pkg/                     # Utilidades (logger, validator, errors, config)
├── tests/                   # Unit, Integration, E2E
├── docs/                    # Documentación
└── storage/                 # Archivos generados (no versionados)
```

---

## 📦 MÓDULOS EXTERNOS EN GITHUB

### 1. `github.com/diegofxm/ubl21-dian` v1.0.0
Librería para generar y firmar documentos UBL 2.1 según DIAN.

**Contenido**:
- `documents/invoice/`, `documents/credit_note/`, etc.
- `signature/` - Firma digital
- `soap/` - Cliente SOAP DIAN
- `xml/` - Utilidades XML

**Uso**:
```go
require github.com/diegofxm/ubl21-dian v1.0.0
```

---

### 2. `github.com/diegofxm/go-yaml-migrator` v0.1.0
Motor de migraciones YAML → SQL para PostgreSQL (como Laravel pero en Go).

**Contenido**:
- `migrator.go` - Lógica principal
- `parser.go` - YAML → SQL
- `tracker.go` - Tracking de migraciones
- `types.go` - Estructuras YAML

**Uso en `database/cmd/migrate/main.go`**:
```go
import migrator "github.com/diegofxm/go-yaml-migrator"

func main() {
    m := migrator.NewMigrator(db, "database/migrations", "database/seeds")
    m.Migrate()
}
```

---

### 3. `github.com/diegofxm/apidian-fixtures` v0.1.0
Datos ficticios para testing (como Laravel factories/seeders).

**Contenido**:
- `factories/` - Generadores de datos (InvoiceFactory, CompanyFactory, etc.)
- `seeders/` - Seeders para tests
- `data/` - Datos estáticos JSON (companies.json, customers.json, products.json)
- `certificates/` - Certificados de prueba (.p12)

**Uso en tests**:
```go
import "github.com/diegofxm/apidian-fixtures/factories"

func TestCreateInvoice(t *testing.T) {
    company := factories.NewCompany().WithNIT("900123456").Build()
    customer := factories.NewCustomer().Build()
    invoice := factories.NewInvoice().
        WithCompany(company).
        WithCustomer(customer).
        WithLines(3).
        Build()
    // Test...
}
```

---

## 🔄 FLUJO DE DEPENDENCIAS (HEXAGONAL PURO)

```
cmd/api/main.go (Inyección de dependencias)
    ↓
adapters/input/http/handlers (HTTP)
    ↓
application/invoice/create_invoice.go (Caso de uso)
    ↓
domain/ports/output/repositories/invoice_repository.go (Interfaz)
    ↑
adapters/output/postgres/repositories/invoice_repository.go (Implementación)
    ↓
PostgreSQL (Infraestructura)
```

**Regla de oro**: Las dependencias apuntan SIEMPRE hacia el dominio (núcleo).

---

## 🚀 ENDPOINTS API REST COMPLETOS

**Total de endpoints: ~70**

### Estructura de Rutas

```go
// internal/adapters/input/http/routes.go

api := app.Group("/api/v1")

// ==================== AUTH ====================
auth := api.Group("/auth")
auth.Post("/register", handlers.Auth.Register)
auth.Post("/login", handlers.Auth.Login)
auth.Post("/refresh", handlers.Auth.RefreshToken)

// Middleware de autenticación
api.Use(middleware.Auth())

// ==================== INVOICES (Facturas) ====================
invoices := api.Group("/invoices")
invoices.Post("/", handlers.Invoice.Create)
invoices.Get("/:id", handlers.Invoice.GetByID)
invoices.Get("/", handlers.Invoice.List)
invoices.Post("/:id/sign", handlers.Invoice.Sign)
invoices.Post("/:id/send", handlers.Invoice.SendToDIAN)
invoices.Get("/:id/status", handlers.Invoice.GetStatus)
invoices.Get("/:id/pdf", handlers.Invoice.GeneratePDF)
invoices.Post("/:id/email", handlers.Invoice.SendEmail)
invoices.Get("/:id/xml", handlers.Invoice.GetXML)

// ==================== CREDIT NOTES ====================
creditNotes := api.Group("/credit-notes")
// ... (9 endpoints similares a invoices)

// ==================== DEBIT NOTES ====================
debitNotes := api.Group("/debit-notes")
// ... (9 endpoints similares)

// ==================== APPLICATION RESPONSES ====================
appResponses := api.Group("/application-responses")
// ... (5 endpoints)

// ==================== ATTACHED DOCUMENTS ====================
attachedDocs := api.Group("/attached-documents")
// ... (7 endpoints)

// ==================== PAYROLLS ====================
payrolls := api.Group("/payrolls")
// ... (9 endpoints)

// ==================== COMPANIES ====================
companies := api.Group("/companies")
// ... (8 endpoints)

// ==================== CUSTOMERS ====================
customers := api.Group("/customers")
// ... (6 endpoints)

// ==================== PRODUCTS ====================
products := api.Group("/products")
// ... (6 endpoints)

// ==================== CATALOGS (Solo lectura) ====================
catalogs := api.Group("/catalogs")
// ... (7 endpoints)
```

---

## 📊 ESTIMACIÓN DE LÍNEAS DE CÓDIGO

### Resumen

| Categoría | Archivos | Líneas |
|-----------|----------|--------|
| **Domain** (entities, valueobjects, ports) | ~60 | ~7,000 |
| **Application** (casos de uso) | ~50 | ~7,000 |
| **Adapters Input** (HTTP handlers, DTOs) | ~40 | ~5,500 |
| **Adapters Output** (Postgres, DIAN, etc.) | ~50 | ~9,000 |
| **Infrastructure** (pkg) | ~5 | ~650 |
| **Main & Setup** | ~2 | ~600 |
| **Tests** | ~100 | ~17,750 |
| **TOTAL** | **~307 archivos** | **~47,500 líneas** |

### Desglose por Documento DIAN

Cada documento (Invoice, CreditNote, DebitNote, ApplicationResponse, AttachedDocument, Payroll) requiere aproximadamente:

```
Por documento:
├── Domain Entity                    ~200 líneas
├── Use Cases (5-6 archivos)         ~750 líneas
├── HTTP Handler                     ~200 líneas
├── DTOs (Request + Response)        ~140 líneas
├── Mappers (HTTP)                   ~100 líneas
├── Repository (Postgres)            ~300 líneas
├── Mapper (Postgres)                ~150 líneas
├── XML Builder (DIAN)               ~250 líneas
├── Mapper (DIAN)                    ~200 líneas
└── Tests                            ~1,000 líneas
────────────────────────────────────────────────
TOTAL por documento:                 ~3,290 líneas
```

**6 documentos × 3,290 = ~19,740 líneas**

---

## ⏱️ TIEMPO DE DESARROLLO ESTIMADO

### Opción 1: Desarrollo Manual (1 desarrollador)

```
Semana 1-2:  Estructura y Domain (2 semanas)
Semana 3-5:  Casos de Uso (3 semanas)
Semana 6-8:  Adaptadores Output (3 semanas)
Semana 9-10: Adaptadores Input (2 semanas)
Semana 11-13: Testing (3 semanas)
Semana 14:   Migración y Ajustes (1 semana)

TOTAL: 14 semanas (~3.5 meses) con 1 dev
       7 semanas (~1.75 meses) con 2 devs
```

### Opción 2: Con Asistencia de IA (Cascade)

```
Semana 1:   Estructura y Domain (5 días)
Semana 2-3: Casos de Uso (10 días)
Semana 4:   Adaptadores Output (5 días)
Semana 5:   Adaptadores Input (5 días)
Semana 6-7: Testing (10 días)
Semana 8:   Migración y Deployment (5 días)

TOTAL: 8 semanas (~2 meses) con asistencia IA
```

---

## 🔧 CONFIGURACIÓN Y SETUP

### go.mod

```go
module apidian-go

go 1.25.1

require (
    // Framework HTTP
    github.com/gofiber/fiber/v2 v2.52.10
    
    // Módulos externos propios ⭐
    github.com/diegofxm/ubl21-dian v1.0.0
    github.com/diegofxm/go-yaml-migrator v0.1.0
    
    // Database
    github.com/lib/pq v1.10.9
    
    // Auth
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.46.0
    
    // Config
    github.com/joho/godotenv v1.5.1
    gopkg.in/yaml.v3 v3.0.1
    
    // PDF, QR, Validation
    github.com/SebastiaanKlippert/go-wkhtmltopdf v1.9.3
    github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
    github.com/go-playground/validator/v10 v10.16.0
    github.com/google/uuid v1.6.0
)

require (
    // Testing (solo en dev) ⭐
    github.com/diegofxm/apidian-fixtures v0.1.0
    github.com/stretchr/testify v1.8.4
)
```

### Makefile

```makefile
.PHONY: help run build test migrate seed

help: ## Mostrar ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Ejecutar aplicación
	go run cmd/api/main.go

build: ## Compilar binario
	go build -o bin/apidian-go cmd/api/main.go

test: ## Ejecutar tests
	go test -v -race -coverprofile=coverage.out ./...

migrate: ## Ejecutar migraciones
	go run database/cmd/migrate/main.go migrate

seed: ## Ejecutar seeds
	go run database/cmd/migrate/main.go seed

docker-up: ## Levantar PostgreSQL
	docker-compose up -d

.DEFAULT_GOAL := help
```

---

## 📝 VENTAJAS DE ESTE DISEÑO

### 1. Testabilidad Máxima
```go
// Test con mocks
mockInvoiceRepo := &mocks.InvoiceRepository{}
mockDIANService := &mocks.DIANService{}

useCase := invoice.NewSendInvoiceUseCase(
    mockInvoiceRepo,
    mockDIANService,
)

err := useCase.Execute(ctx, 1)
assert.NoError(t, err)
```

### 2. Cambiar Adaptadores Sin Tocar Dominio
- Cambiar de PostgreSQL a MongoDB → Solo cambias `adapters/output/postgres/`
- Cambiar de Fiber a Gin → Solo cambias `adapters/input/http/`
- Cambiar de ubl21-dian a otra lib → Solo cambias `adapters/output/dian/`

### 3. Lógica de Negocio Centralizada
```go
invoice.CalculateTotals()  // En entity, no en servicio
invoice.Validate()         // En entity, no en handler
invoice.CanBeSent()        // En entity, no en repository
```

### 4. Independencia de Frameworks
El dominio NO depende de Fiber, PostgreSQL, ni ubl21-dian.

---

## 🎓 COMPARACIÓN: ACTUAL vs HEXAGONAL

| Métrica | Actual | Hexagonal | Cambio |
|---------|--------|-----------|--------|
| Líneas de código | ~15,000 | ~47,500 | +217% |
| Archivos | ~120 | ~307 | +156% |
| Testabilidad | Media | Muy Alta | +++ |
| Mantenibilidad | Media | Muy Alta | +++ |
| Acoplamiento | Alto | Bajo | --- |
| Documentos soportados | 1 (FV) | 6 (todos) | +500% |
| Categoría | API REST | API REST | = |

---

## 📋 CHECKLIST DE IMPLEMENTACIÓN

### Fase 1: Fundamentos (Semana 1)
- [ ] Crear estructura de carpetas completa
- [ ] Definir entities en `domain/entities/`
- [ ] Definir value objects en `domain/valueobjects/`
- [ ] Definir interfaces en `domain/ports/`

### Fase 2: Casos de Uso (Semana 2-3)
- [ ] Implementar casos de uso de Invoice
- [ ] Implementar casos de uso de CreditNote
- [ ] Implementar casos de uso de DebitNote
- [ ] Implementar casos de uso de ApplicationResponse
- [ ] Implementar casos de uso de AttachedDocument
- [ ] Implementar casos de uso de Payroll
- [ ] Implementar casos de uso de Company, Customer, Product, Auth

### Fase 3: Adaptadores de Salida (Semana 4)
- [ ] Implementar PostgreSQL repositories
- [ ] Implementar DIAN adapter (wrapper ubl21-dian)
- [ ] Implementar Storage, PDF, Email, QR adapters

### Fase 4: Adaptadores de Entrada (Semana 5)
- [ ] Implementar HTTP handlers
- [ ] Implementar DTOs y mappers
- [ ] Implementar middleware
- [ ] Configurar routes

### Fase 5: Testing (Semana 6-7)
- [ ] Tests unitarios de domain
- [ ] Tests de casos de uso con mocks
- [ ] Tests de integración
- [ ] Tests E2E

### Fase 6: Deployment (Semana 8)
- [ ] Migración de datos
- [ ] Documentación
- [ ] Deploy a producción

---

## 🎯 CONCLUSIÓN

Este diseño te da:
- ✅ **100% Hexagonal** desde el día 1
- ✅ **API REST** (NO cambia la categoría)
- ✅ **6 Documentos DIAN** completos
- ✅ **Testeable** con mocks fáciles
- ✅ **Escalable** (agregar features sin romper nada)
- ✅ **Mantenible** (cambios aislados por capa)
- ✅ **Independiente** de frameworks
- ✅ **Binario optimizado** (~18 MB con módulos externos)
- ✅ **Módulos reutilizables** (ubl21-dian, go-yaml-migrator, apidian-fixtures)

---

**Fecha de creación**: 2026-02-07  
**Versión**: 2.0  
**Autor**: Equipo de Desarrollo apidian-go
