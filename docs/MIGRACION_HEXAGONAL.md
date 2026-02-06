# 🏗️ PLAN DE MIGRACIÓN A HEXAGONAL PURO + MÓDULOS EXTERNOS

## 📊 ANÁLISIS INICIAL

### Estado Actual
```
apidian-go/
├── database/          → 55 archivos (migrations + seeds)
├── ubl21-dian/        → 89 archivos (ya tiene go.mod propio)
└── internal/          → Lógica de negocio
```

### Dependencias Actuales
```go
// go.mod línea 45
github.com/diegofxm/ubl21-dian v0.0.0

// go.mod línea 60
replace github.com/diegofxm/ubl21-dian => ./ubl21-dian  // ❌ Local
```

---

## ✅ RESPUESTA A PREGUNTAS CLAVE

### 1️⃣ ¿Es viable extraer `database` y `ubl21-dian` como módulos externos?

**SÍ, es viable y RECOMENDADO:**

#### `ubl21-dian` → GitHub Module
- ✅ **Ya tiene `go.mod` propio**
- ✅ **Es reutilizable** para otros proyectos DIAN
- ✅ **Reduce binario** (Go solo incluye lo usado)
- ✅ **Versionado independiente**

**Tamaño estimado del binario:**
- **Antes**: ~40-50 MB (con todo embebido)
- **Después**: ~15-20 MB (solo API + dependencias necesarias)

#### `database` → ¿Módulo externo?

**❌ NO recomendado como módulo:**
- Las migraciones y seeds son **específicas de tu aplicación**
- Cambiarán frecuentemente con tu negocio
- No son reutilizables en otros proyectos

**✅ MEJOR: Mantener en el proyecto pero reorganizar:**
```
apidian-go/
├── migrations/     → Solo archivos SQL
├── seeds/          → Solo archivos SQL
└── internal/
    └── infrastructure/
        └── database/
            └── connection.go
```

---

## 🎯 PLAN DE MIGRACIÓN COMPLETO

---

## FASE 1: EXTRAER `ubl21-dian` COMO MÓDULO EXTERNO

### Paso 1.1: Crear repositorio GitHub
```bash
# En GitHub, crear nuevo repo:
# github.com/diegofxm/ubl21-dian

# Mover contenido actual
cd /var/www/apidian-go/ubl21-dian
git init
git add .
git commit -m "Initial commit: UBL 2.1 DIAN library"
git remote add origin https://github.com/diegofxm/ubl21-dian.git
git push -u origin main
```

### Paso 1.2: Versionar el módulo
```bash
# Crear primera versión
git tag v0.1.0
git push origin v0.1.0
```

### Paso 1.3: Actualizar `go.mod` de apidian-go
```go
// go.mod - ANTES
replace github.com/diegofxm/ubl21-dian => ./ubl21-dian

// go.mod - DESPUÉS
require (
    github.com/diegofxm/ubl21-dian v0.1.0  // ✅ Desde GitHub
)
```

### Paso 1.4: Eliminar carpeta local
```bash
cd /var/www/apidian-go
rm -rf ubl21-dian/
go mod tidy
```

**Beneficio:** Binario reduce ~10-15 MB

---

## FASE 2: REORGANIZAR `database` Y EXTRAER `engine` COMO MÓDULO

### Análisis de la estructura actual de `database/`

**Estado actual:**
```
database/
├── cmd/migrate/main.go          # CLI de migraciones
├── engine/                      # ⭐ Motor reutilizable (4 archivos Go)
│   ├── migrator.go              # Lógica principal (504 líneas)
│   ├── parser.go                # YAML → SQL (4758 bytes)
│   ├── tracker.go               # Tracking de migraciones (3283 bytes)
│   └── types.go                 # Estructuras YAML (86 líneas)
├── migrations/                  # ⭐ 31 archivos YAML (específicos del proyecto)
│   ├── 000_create_extensions.yaml
│   ├── 101_create_companies.yaml
│   └── ...
└── seeds/                       # ⭐ 18 archivos CSV (catálogos DIAN)
    ├── municipalities.csv       # 350+ municipios
    ├── document_types.csv
    └── ...
```

**Características del sistema:**
- ✅ **Migraciones en YAML** (no SQL) con parser Go
- ✅ **Seeds en CSV** con PostgreSQL `COPY FROM` (10-100x más rápido)
- ✅ **Engine reutilizable** independiente del proyecto
- ✅ **CLI independiente** (`go run database/cmd/migrate/main.go`)

---

### Paso 2.1: Extraer `database/engine` como módulo GitHub

**Razón:** El engine es **100% reutilizable** para cualquier proyecto Go + PostgreSQL

#### A) Crear repositorio GitHub
```bash
# Crear repo: github.com/diegofxm/go-yaml-migrator

cd /var/www/apidian-go/database/engine
git init
git add .
git commit -m "Initial commit: Go YAML Migration Engine for PostgreSQL"
git remote add origin https://github.com/diegofxm/go-yaml-migrator.git
git push -u origin main
git tag v0.1.0
git push origin v0.1.0
```

#### B) Estructura del módulo externo
```
go-yaml-migrator/
├── go.mod                       # module github.com/diegofxm/go-yaml-migrator
├── README.md                    # Documentación completa
├── migrator.go                  # Desde database/engine/migrator.go
├── parser.go                    # Desde database/engine/parser.go
├── tracker.go                   # Desde database/engine/tracker.go
├── types.go                     # Desde database/engine/types.go
└── examples/
    ├── migrations/
    │   └── 001_example.yaml
    └── seeds/
        └── example.csv
```

#### C) Actualizar `apidian-go`
```go
// go.mod - DESPUÉS
require (
    github.com/diegofxm/go-yaml-migrator v0.1.0  // ✅ Módulo externo
)
```

```go
// database/cmd/migrate/main.go - DESPUÉS
import (
    migrator "github.com/diegofxm/go-yaml-migrator"  // ⭐ Desde GitHub
)

func main() {
    m := migrator.NewMigrator(db, "database/migrations", "database/seeds")
    m.Migrate()
}
```

---

### Paso 2.2: Reorganizar estructura del proyecto

**Nueva estructura (después de extraer engine):**
```
apidian-go/
├── database/                    # ⭐ Solo datos del proyecto
│   ├── cmd/migrate/main.go      # CLI (usa go-yaml-migrator)
│   ├── migrations/              # 31 archivos YAML (específicos)
│   │   ├── 000_create_extensions.yaml
│   │   ├── 010_create_countries.yaml
│   │   ├── 101_create_companies.yaml
│   │   └── ...
│   └── seeds/                   # 18 archivos CSV (catálogos DIAN)
│       ├── municipalities.csv   # 350+ registros
│       ├── document_types.csv
│       └── ...
│
└── internal/
    └── infrastructure/
        └── database/
            └── connection.go    # Conexión a PostgreSQL
```

**Beneficios:**
- ✅ **Engine reutilizable** en otros proyectos
- ✅ **Binario más pequeño** (solo incluye lo necesario)
- ✅ **Datos del proyecto separados** (migrations/ y seeds/)
- ✅ **Versionado independiente** del engine

---

### Paso 2.3: Alternativa (si NO quieres extraer engine)

Si prefieres mantener el engine en el proyecto:

```
apidian-go/
├── database/
│   ├── cmd/migrate/main.go
│   ├── engine/                  # Mantener aquí
│   │   ├── migrator.go
│   │   ├── parser.go
│   │   ├── tracker.go
│   │   └── types.go
│   ├── migrations/              # 31 YAML
│   └── seeds/                   # 18 CSV
│
└── internal/
    └── infrastructure/
        └── database/
            └── connection.go
```

**Nota:** Esta opción NO reduce el binario, pero mantiene todo en un solo repo.

---

### Paso 2.4: Actualizar imports (si extraes engine)

```go
// Antes
import "apidian-go/database/engine"

// Después
import migrator "github.com/diegofxm/go-yaml-migrator"
```

**Beneficio:** 
- **Con extracción**: Binario reduce ~2-3 MB + engine reutilizable
- **Sin extracción**: Estructura más simple, todo en un repo

---

## FASE 3: MIGRAR A HEXAGONAL PURO

### Paso 3.1: Crear estructura de puertos

```
internal/
├── domain/
│   ├── entities/                    → Entidades puras
│   │   ├── invoice.go
│   │   ├── company.go
│   │   ├── customer.go
│   │   └── product.go
│   │
│   ├── valueobjects/                → Objetos de valor
│   │   ├── money.go
│   │   ├── tax.go
│   │   └── address.go
│   │
│   └── ports/                       → ⭐ INTERFACES (contratos)
│       ├── input/                   → Casos de uso
│       │   ├── invoice_usecase.go
│       │   ├── company_usecase.go
│       │   └── customer_usecase.go
│       │
│       └── output/                  → Repositorios y servicios externos
│           ├── invoice_repository.go
│           ├── company_repository.go
│           ├── dian_service.go      → ⭐ Abstracción de ubl21-dian
│           ├── storage_service.go
│           └── pdf_service.go
│
├── application/                     → Implementación de casos de uso
│   ├── invoice_service.go           → Implementa InvoiceUseCase
│   ├── company_service.go
│   └── customer_service.go
│
└── adapters/                        → Implementaciones concretas
    ├── input/
    │   └── http/
    │       ├── invoice_handler.go   → Fiber handlers
    │       ├── company_handler.go
    │       └── routes.go
    │
    └── output/
        ├── postgres/
        │   ├── invoice_repository.go    → Implementa InvoiceRepository
        │   ├── company_repository.go
        │   └── customer_repository.go
        │
        ├── dian/
        │   └── dian_adapter.go          → ⭐ Wrapper de ubl21-dian
        │
        ├── filesystem/
        │   └── storage_adapter.go       → Implementa StorageService
        │
        └── pdf/
            └── pdf_adapter.go           → Implementa PDFService
```

---

### Paso 3.2: Definir Puertos (Interfaces)

#### A) Puerto de Entrada: Casos de Uso
```go
// internal/domain/ports/input/invoice_usecase.go
package input

import "apidian-go/internal/domain/entities"

type InvoiceUseCase interface {
    CreateInvoice(req CreateInvoiceRequest) (*entities.Invoice, error)
    GetInvoiceByID(id uint) (*entities.Invoice, error)
    SignInvoice(id uint) error
    SendInvoiceToDIAN(id uint) error
    GetInvoiceStatus(id uint) (*DIANStatus, error)
}

type CreateInvoiceRequest struct {
    CompanyID  uint
    CustomerID uint
    Lines      []InvoiceLineRequest
}
```

#### B) Puerto de Salida: Repositorio
```go
// internal/domain/ports/output/invoice_repository.go
package output

import "apidian-go/internal/domain/entities"

type InvoiceRepository interface {
    Create(invoice *entities.Invoice) error
    GetByID(id uint) (*entities.Invoice, error)
    Update(invoice *entities.Invoice) error
    Delete(id uint) error
    GetAll(filters InvoiceFilters) ([]*entities.Invoice, error)
    UpdateStatus(id uint, status string) error
    UpdateTrackId(id uint, trackId string) error
}
```

#### C) Puerto de Salida: Servicio DIAN
```go
// internal/domain/ports/output/dian_service.go
package output

import "apidian-go/internal/domain/entities"

type DIANService interface {
    // Envío de documentos
    SendInvoiceSync(invoice *entities.Invoice) (*DIANResponse, error)
    SendInvoiceAsync(invoice *entities.Invoice) (zipKey string, error)
    SendTestSet(invoices []*entities.Invoice, testSetID string) (zipKey string, error)
    
    // Consultas
    GetStatus(trackId string) (*DIANStatus, error)
    GetStatusZip(zipKey string) (*DIANStatus, error)
    
    // Generación XML
    GenerateXML(invoice *entities.Invoice) ([]byte, error)
    SignXML(xmlBytes []byte, certificate []byte, password string) ([]byte, error)
}

type DIANResponse struct {
    IsValid           bool
    StatusCode        string
    StatusDescription string
    XmlDocumentKey    string
    XmlBase64Bytes    string
}
```

---

### Paso 3.3: Implementar Adaptadores

#### A) Adaptador DIAN (Wrapper de ubl21-dian)
```go
// internal/adapters/output/dian/dian_adapter.go
package dian

import (
    "apidian-go/internal/domain/entities"
    "apidian-go/internal/domain/ports/output"
    
    "github.com/diegofxm/ubl21-dian/soap"           // ⭐ Módulo externo
    "github.com/diegofxm/ubl21-dian/soap/types"
    "github.com/diegofxm/ubl21-dian/documents/invoice"
)

type DIANAdapter struct {
    client *soap.Client
    config *DIANConfig
}

func NewDIANAdapter(config *DIANConfig) (output.DIANService, error) {
    client, err := soap.NewClient(&soap.Config{
        WSDLURL:     config.WSDLURL,
        Certificate: config.CertificatePath,
        PrivateKey:  config.PrivateKeyPath,
    })
    if err != nil {
        return nil, err
    }
    
    return &DIANAdapter{
        client: client,
        config: config,
    }, nil
}

func (d *DIANAdapter) SendInvoiceSync(inv *entities.Invoice) (*output.DIANResponse, error) {
    // Convertir entities.Invoice → types.SendBillSyncRequest
    request := d.buildSyncRequest(inv)
    
    // Llamar a ubl21-dian
    response, err := d.client.SendBillSync(request)
    if err != nil {
        return nil, err
    }
    
    // Convertir types.Response → output.DIANResponse
    return &output.DIANResponse{
        IsValid:           response.IsValid,
        StatusCode:        response.StatusCode,
        StatusDescription: response.StatusDescription,
        XmlDocumentKey:    response.XmlDocumentKey,
        XmlBase64Bytes:    response.XmlBase64Bytes,
    }, nil
}

// ... otros métodos
```

#### B) Adaptador PostgreSQL
```go
// internal/adapters/output/postgres/invoice_repository.go
package postgres

import (
    "apidian-go/internal/domain/entities"
    "apidian-go/internal/domain/ports/output"
    "database/sql"
)

type InvoiceRepository struct {
    db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) output.InvoiceRepository {
    return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(invoice *entities.Invoice) error {
    query := `INSERT INTO invoices (company_id, customer_id, number, ...) VALUES ($1, $2, $3, ...)`
    _, err := r.db.Exec(query, invoice.CompanyID, invoice.CustomerID, invoice.Number)
    return err
}

// ... otros métodos
```

#### C) Adaptador HTTP (Handlers)
```go
// internal/adapters/input/http/invoice_handler.go
package http

import (
    "apidian-go/internal/domain/ports/input"
    "github.com/gofiber/fiber/v2"
)

type InvoiceHandler struct {
    invoiceUseCase input.InvoiceUseCase  // ⭐ Interfaz, no implementación
}

func NewInvoiceHandler(useCase input.InvoiceUseCase) *InvoiceHandler {
    return &InvoiceHandler{
        invoiceUseCase: useCase,
    }
}

func (h *InvoiceHandler) SendToDIAN(c *fiber.Ctx) error {
    id, _ := c.ParamsInt("id")
    
    // Delegar al caso de uso
    err := h.invoiceUseCase.SendInvoiceToDIAN(uint(id))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(fiber.Map{"success": true})
}
```

---

### Paso 3.4: Implementar Casos de Uso

```go
// internal/application/invoice_service.go
package application

import (
    "apidian-go/internal/domain/entities"
    "apidian-go/internal/domain/ports/input"
    "apidian-go/internal/domain/ports/output"
)

type InvoiceService struct {
    invoiceRepo   output.InvoiceRepository  // ⭐ Interfaz
    companyRepo   output.CompanyRepository  // ⭐ Interfaz
    dianService   output.DIANService        // ⭐ Interfaz
    storageService output.StorageService    // ⭐ Interfaz
}

func NewInvoiceService(
    invoiceRepo output.InvoiceRepository,
    companyRepo output.CompanyRepository,
    dianService output.DIANService,
    storageService output.StorageService,
) input.InvoiceUseCase {
    return &InvoiceService{
        invoiceRepo:    invoiceRepo,
        companyRepo:    companyRepo,
        dianService:    dianService,
        storageService: storageService,
    }
}

func (s *InvoiceService) SendInvoiceToDIAN(id uint) error {
    // 1. Obtener factura
    invoice, err := s.invoiceRepo.GetByID(id)
    if err != nil {
        return err
    }
    
    // 2. Validar (lógica en domain)
    if !invoice.CanBeSent() {
        return errors.New("invoice must be signed first")
    }
    
    // 3. Enviar a DIAN (delegado al adaptador)
    response, err := s.dianService.SendInvoiceSync(invoice)
    if err != nil {
        return err
    }
    
    // 4. Actualizar estado
    if response.IsValid {
        invoice.MarkAsSent()
        s.invoiceRepo.Update(invoice)
    }
    
    return nil
}
```

---

### Paso 3.5: Inyección de Dependencias (main.go)

```go
// cmd/api/main.go
package main

import (
    "apidian-go/internal/adapters/input/http"
    "apidian-go/internal/adapters/output/dian"
    "apidian-go/internal/adapters/output/postgres"
    "apidian-go/internal/application"
    "apidian-go/internal/infrastructure/database"
)

func main() {
    // 1. Configuración
    cfg := loadConfig()
    
    // 2. Conexión a DB
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. Crear adaptadores de SALIDA (implementaciones)
    invoiceRepo := postgres.NewInvoiceRepository(db)
    companyRepo := postgres.NewCompanyRepository(db)
    dianService, _ := dian.NewDIANAdapter(&dian.DIANConfig{
        WSDLURL:         cfg.DIAN.WSDLURL,
        CertificatePath: cfg.DIAN.CertPath,
        PrivateKeyPath:  cfg.DIAN.KeyPath,
    })
    storageService := filesystem.NewStorageAdapter(cfg.Storage.Path)
    
    // 4. Crear casos de uso (inyectar interfaces)
    invoiceUseCase := application.NewInvoiceService(
        invoiceRepo,      // output.InvoiceRepository
        companyRepo,      // output.CompanyRepository
        dianService,      // output.DIANService
        storageService,   // output.StorageService
    )
    
    // 5. Crear adaptadores de ENTRADA (handlers)
    invoiceHandler := http.NewInvoiceHandler(invoiceUseCase)
    
    // 6. Setup routes
    app := fiber.New()
    http.SetupRoutes(app, invoiceHandler)
    
    // 7. Start server
    app.Listen(":8080")
}
```

---

## � DATABASE EN ARQUITECTURA HEXAGONAL

### ¿Dónde encaja `database/` en hexagonal?

**`database/` NO es parte de la arquitectura hexagonal de la aplicación.**

Es un **sistema independiente** de gestión de esquema que:
- ✅ Se ejecuta **antes** de la aplicación (setup inicial)
- ✅ Tiene su propio CLI (`go run database/cmd/migrate/main.go`)
- ✅ No depende de la lógica de negocio
- ✅ Es **infraestructura pura** (DDL, no DML)

### Relación con la arquitectura hexagonal

```
┌─────────────────────────────────────────────────────────┐
│ SETUP (Una vez)                                         │
│                                                          │
│  database/                                              │
│  ├── cmd/migrate/main.go  → Crea esquema PostgreSQL    │
│  ├── migrations/*.yaml    → Definiciones de tablas     │
│  └── seeds/*.csv          → Catálogos DIAN             │
│                                                          │
│  Resultado: Base de datos lista con esquema y datos    │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│ APLICACIÓN (Hexagonal)                                  │
│                                                          │
│  internal/                                              │
│  ├── domain/                                            │
│  │   └── ports/output/                                 │
│  │       └── invoice_repository.go  ← Interfaz         │
│  │                                                      │
│  ├── adapters/output/postgres/                         │
│  │   └── invoice_repository.go  ← Implementación       │
│  │       (Usa tablas creadas por migrations)           │
│  │                                                      │
│  └── infrastructure/database/                          │
│      └── connection.go  ← Conexión a PostgreSQL        │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Flujo de trabajo

```bash
# 1. Setup inicial (una vez)
go run database/cmd/migrate/main.go migrate
go run database/cmd/migrate/main.go seed

# 2. Ejecutar aplicación (usa esquema creado)
go run cmd/api/main.go
```

### Estructura final recomendada

```
apidian-go/
├── database/                        # ⭐ Sistema independiente de migraciones
│   ├── cmd/migrate/main.go          # CLI de migraciones
│   ├── migrations/*.yaml            # 31 archivos (DDL)
│   └── seeds/*.csv                  # 18 archivos (catálogos)
│
├── internal/                        # ⭐ Arquitectura hexagonal
│   ├── domain/
│   │   ├── entities/
│   │   └── ports/
│   ├── application/
│   └── adapters/
│       ├── input/http/
│       └── output/
│           └── postgres/            # Usa tablas de migrations/
│
└── cmd/api/main.go                  # Aplicación principal
```

**Conclusión:** `database/` permanece separado. Solo extraes el `engine/` como módulo si quieres reutilizarlo.

---

## �📋 PLAN DE EJECUCIÓN PASO A PASO

### Semana 1: Preparación y Módulos Externos
- [ ] **Día 1-2**: Crear repo GitHub para `ubl21-dian`
  - Crear repositorio en GitHub
  - Inicializar git en carpeta local
  - Push inicial
- [ ] **Día 3**: Versionar y publicar `ubl21-dian` v0.1.0
  - Crear tag v0.1.0
  - Verificar que se puede importar desde GitHub
- [ ] **Día 4**: Actualizar `go.mod` y probar compilación
  - Eliminar `replace` directive
  - Agregar `require github.com/diegofxm/ubl21-dian v0.1.0`
  - Ejecutar `go mod tidy`
  - Compilar y verificar que funciona
- [ ] **Día 5**: (OPCIONAL) Extraer `database/engine` como módulo
  - Crear repo `go-yaml-migrator`
  - Publicar v0.1.0
  - Actualizar `database/cmd/migrate/main.go`

### Semana 2: Estructura Hexagonal
- [ ] **Día 1-2**: Crear carpetas `domain/ports/` y definir interfaces
  - Crear `internal/domain/ports/input/`
  - Crear `internal/domain/ports/output/`
  - Definir interfaces de casos de uso
  - Definir interfaces de repositorios
  - Definir interfaz DIANService
- [ ] **Día 3-4**: Crear `adapters/output/dian/` (wrapper)
  - Implementar DIANAdapter
  - Convertir tipos entre domain y ubl21-dian
  - Probar envío a DIAN
- [ ] **Día 5**: Crear `adapters/output/postgres/` (repositories)
  - Implementar InvoiceRepository
  - Implementar CompanyRepository
  - Implementar CustomerRepository

### Semana 3: Migración de Lógica
- [ ] **Día 1-2**: Mover lógica de `service/` a `application/`
  - Crear InvoiceService en application/
  - Migrar lógica de negocio
  - Usar interfaces en lugar de implementaciones
- [ ] **Día 3-4**: Refactorizar `handler/` a `adapters/input/http/`
  - Mover handlers a nueva ubicación
  - Actualizar para usar interfaces de casos de uso
  - Actualizar routes.go
- [ ] **Día 5**: Actualizar `main.go` con inyección de dependencias
  - Implementar patrón de inyección
  - Crear todos los adaptadores
  - Conectar todo el flujo

### Semana 4: Testing y Ajustes
- [ ] **Día 1-3**: Crear tests unitarios para cada capa
  - Tests de domain (entidades y value objects)
  - Tests de application (casos de uso con mocks)
  - Tests de adapters (con mocks de interfaces)
- [ ] **Día 4**: Reorganizar `database/` a `migrations/` y `seeds/`
  - Mover archivos SQL
  - Actualizar migrator y seeder
  - Probar migraciones
- [ ] **Día 5**: Documentación y deployment
  - Actualizar README.md
  - Documentar arquitectura
  - Preparar para producción

---

## 💰 BENEFICIOS FINALES

### Tamaño del Binario
- **Antes**: ~45 MB
- **Después**: ~18 MB (-60%)

### Arquitectura
- ✅ Hexagonal pura (100%)
- ✅ Testeable (mocks fáciles)
- ✅ Mantenible (cambios aislados)
- ✅ Escalable (agregar adaptadores sin tocar dominio)

### Módulos
- ✅ `ubl21-dian` reutilizable en otros proyectos
- ✅ Versionado independiente
- ✅ Binario más ligero

---

## ⚠️ CONSIDERACIONES IMPORTANTES

1. **Tiempo estimado**: 3-4 semanas
2. **Riesgo**: Medio (muchos cambios)
3. **Recomendación**: Hacerlo en **branch separado** y mergear cuando esté estable
4. **Testing**: Crear tests antes de migrar para asegurar que todo funciona igual

---

## 🎯 RECOMENDACIÓN FINAL

**Hazlo AHORA antes de producción**, pero:
1. Crea branch `feature/hexagonal-architecture`
2. Migra por fases (1 módulo a la vez)
3. Mantén `main` funcional
4. Mergea cuando todo esté probado

---

## 📚 RECURSOS ADICIONALES

### Lecturas Recomendadas
- [Hexagonal Architecture by Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Modules Reference](https://go.dev/ref/mod)

### Ejemplos de Código
- [Go Hexagonal Architecture Example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [Go Clean Architecture](https://github.com/bxcodec/go-clean-arch)

---

## 📝 CHECKLIST DE MIGRACIÓN

### Pre-requisitos
- [ ] Backup completo del proyecto
- [ ] Tests existentes funcionando
- [ ] Crear branch `feature/hexagonal-architecture`

### Fase 1: Módulos Externos
- [ ] Publicar `ubl21-dian` en GitHub
- [ ] Actualizar `go.mod`
- [ ] Verificar compilación

### Fase 2: Estructura
- [ ] Crear carpetas de arquitectura hexagonal
- [ ] Definir todas las interfaces (ports)
- [ ] Mover entidades a `domain/entities/`

### Fase 3: Adaptadores
- [ ] Implementar adaptadores de salida (DB, DIAN, Storage)
- [ ] Implementar adaptadores de entrada (HTTP handlers)
- [ ] Actualizar casos de uso

### Fase 4: Testing
- [ ] Tests unitarios de domain
- [ ] Tests de application con mocks
- [ ] Tests de integración
- [ ] Tests end-to-end

### Fase 5: Deployment
- [ ] Documentación actualizada
- [ ] CI/CD configurado
- [ ] Merge a main
- [ ] Deploy a producción

---

**Fecha de creación**: 2026-02-06  
**Versión**: 1.0  
**Autor**: Equipo de Desarrollo apidian-go
