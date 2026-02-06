# 🚀 APIDIAN-GO API

API REST en Go con Fiber para facturación electrónica DIAN Colombia.

## 📋 Tabla de Contenidos

- [Características](#características)
- [Arquitectura](#arquitectura)
- [Requisitos](#requisitos)
- [Instalación](#instalación)
- [Configuración](#configuración)
- [Uso](#uso)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [API Endpoints](#api-endpoints)

## ✨ Características

- ✅ **Fiber Framework** - HTTP framework ultra rápido
- ✅ **PostgreSQL** - Base de datos con conexión independiente
- ✅ **JWT Authentication** - Autenticación segura con tokens
- ✅ **Arquitectura limpia** - Separación de capas (domain, service, repository, handler)
- ✅ **Middleware** - CORS, Logger, Error Handler, Auth
- ✅ **Sistema de migraciones** - Migraciones YAML + Seeds CSV
- ✅ **Respuestas estandarizadas** - Sistema de respuestas HTTP consistente

## 🏗️ Arquitectura

```
apidian-go/
├── cmd/api/                    # Punto de entrada
├── internal/                   # Código privado
│   ├── config/                # Configuración
│   ├── domain/                # Entidades de negocio
│   ├── repository/            # Acceso a datos
│   ├── service/               # Lógica de negocio
│   ├── handler/               # HTTP handlers
│   ├── middleware/            # Middleware (auth, cors, logger)
│   └── infrastructure/        # Infraestructura (DB, crypto, storage)
├── pkg/                       # Código reutilizable
│   ├── response/             # Respuestas HTTP
│   └── errors/               # Errores personalizados
├── database/                  # Sistema de migraciones
│   ├── migrations/           # Migraciones YAML
│   └── seeds/                # Seeds CSV
└── storage/                   # Archivos generados
```

## 📦 Requisitos

### Software Base
- **Go 1.21+**
- **PostgreSQL 12+**

### Dependencias del Sistema

Para compilar y ejecutar correctamente la aplicación, necesitas instalar las siguientes dependencias:

#### 1. libxml2-dev y pkg-config
**Requerido para:** Canonicalización XML (C14N 1.0) con CGO
```bash
sudo apt-get install -y libxml2-dev pkg-config
```

#### 2. OpenSSL
**Requerido para:** Conversión de certificados P12 a PEM (fallback automático)
```bash
sudo apt-get install -y openssl
```

#### 3. build-essential (gcc)
**Requerido para:** Compilación con CGO
```bash
sudo apt-get install -y build-essential
```

#### Instalación rápida de todas las dependencias:
```bash
# Limpiar caché de APT (en caso de error de espacio en disco)
sudo apt-get clean
sudo apt-get autoclean
sudo apt-get autoremove

# Actualizar repositorios e instalar dependencias
sudo apt-get update
sudo apt-get install -y libxml2-dev pkg-config openssl build-essential postgresql-client
```

## 🔧 Instalación

### 1. Clonar el repositorio

```bash
cd /var/www/apidian-go
```

### 2. Instalar dependencias

```bash
go mod download
```

### 3. Configurar variables de entorno

Copia `.env.example` a `.env` y configura:

```bash
cp .env.example .env
```

### 4. Ejecutar migraciones

```bash
# Ejecutar todas las migraciones
go run database/cmd/migrate/main.go migrate

# Cargar seeds (catálogos DIAN)
go run database/cmd/migrate/main.go seed
```

## ⚙️ Configuración

### 1. Copiar archivo de configuración

```bash
cp .env.example .env
```

### 2. Generar ENCRYPTION_KEY

El sistema usa **AES-256-GCM** para cifrar passwords de certificados digitales. Debes generar una clave de 32 bytes (64 caracteres hexadecimales):

#### Opción A: OpenSSL (Linux/WSL/Git Bash)
```bash
openssl rand -hex 32
```

#### Opción B: PowerShell (Windows)
```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Maximum 256) })
```

**Resultado ejemplo:**
```
a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2
```

⚠️ **IMPORTANTE:**
- Genera claves diferentes para desarrollo y producción
- Guarda la clave en lugar seguro (password manager)
- Si pierdes la clave, NO podrás descifrar passwords existentes
- NUNCA subas el `.env` a Git

### 3. Configurar variables de entorno

Edita el archivo `.env`:

```env
# Server Configuration
SERVER_PORT=3000
APP_ENV=development
CORS_ALLOW_ORIGINS=*
TZ=America/Bogota

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=apidian
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION=24

# Invoice Configuration
KEEP_UNSIGNED_XML=false

# Encryption Configuration (AES-256-GCM)
# Generate with: openssl rand -hex 32
ENCRYPTION_KEY=your-generated-64-char-hex-key-here
```

## 🚀 Uso

### Iniciar el servidor

```bash
go run cmd/api/main.go
```

El servidor estará disponible en `http://localhost:3000`

### Health Check

```bash
curl http://localhost:3000/health
```

Respuesta:
```json
{
  "status": "ok",
  "env": "development"
}
```

## 📁 Estructura del Proyecto

### **cmd/api/**
Punto de entrada de la aplicación.

### **internal/**
Código privado de la aplicación (no exportable).

#### **internal/config/**
Configuración de la aplicación (carga variables de entorno).

#### **internal/domain/**
Entidades de negocio (Company, Customer, Product, Invoice, etc.).

#### **internal/repository/**
Capa de acceso a datos (queries SQL, CRUD).

#### **internal/service/**
Lógica de negocio (validaciones, orquestación).

#### **internal/handler/**
HTTP handlers (controllers) con Fiber.

#### **internal/middleware/**
- `auth.go` - Autenticación JWT
- `cors.go` - CORS
- `logger.go` - Logging de requests
- `error.go` - Manejo de errores

#### **internal/infrastructure/**
- `database/` - Conexión a PostgreSQL
- `crypto/` - Encriptación (certificados)
- `storage/` - Almacenamiento de archivos

### **pkg/**
Código reutilizable (exportable).

#### **pkg/response/**
Sistema de respuestas HTTP estandarizadas:
- `Success()` - 200 OK
- `Created()` - 201 Created
- `BadRequest()` - 400 Bad Request
- `Unauthorized()` - 401 Unauthorized
- `NotFound()` - 404 Not Found
- `InternalServerError()` - 500 Internal Server Error

#### **pkg/errors/**
Errores personalizados de la aplicación.

### **database/**
Sistema de migraciones independiente.

## ✨ Funcionalidades Implementadas

### Gestión de Facturas
- ✅ Crear facturas en estado "draft"
- ✅ Firmar facturas con certificado digital (XAdES-BES)
- ✅ Actualizar fechas automáticamente al firmar
- ✅ Calcular CUFE (Código Único de Factura Electrónica)
- ✅ Generar QR code para validación

### Generación de PDFs
- ✅ **On-demand:** PDFs se generan dinámicamente, no se guardan en disco
- ✅ **Preview:** Facturas en draft muestran CUFE/QR de ejemplo
- ✅ **Final:** Facturas firmadas muestran CUFE/QR reales
- ✅ **Sistema de templates:** Diseño modular y extensible
- ✅ **Generación nativa:** Usando Maroto (Go puro, sin dependencias externas)

### Firma Digital
- ✅ Soporte para certificados P12 (DER y BER)
- ✅ Conversión automática P12 → PEM con OpenSSL (fallback)
- ✅ Firma XAdES-BES según estándar DIAN
- ✅ Canonicalización C14N 1.0 con libxml2 (mismo comportamiento que PHP)
- ✅ Actualización automática de fechas (IssueDate = SigningTime)

### Seguridad y Cifrado
- ✅ **AES-256-GCM:** Cifrado fuerte para passwords de certificados
- ✅ **AEAD:** Autenticación y detección de manipulación
- ✅ **Nonce aleatorio:** Cada cifrado usa un nonce único de 12 bytes
- ✅ **Clave maestra:** Una sola `ENCRYPTION_KEY` para todo el sistema
- ✅ **Bcrypt:** Hashing irreversible para passwords de usuarios (recomendado)

### Envío a DIAN
- ✅ Cliente SOAP implementado
- ✅ WS-Security header configurado
- ✅ AttachedDocument generado
- ⏳ **Pendiente:** Validación completa con DIAN

## 🔌 API Endpoints

### **Públicos**

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/ping` | Ping |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/register` | Registro |

### **Protegidos** (requieren JWT)

#### **Companies**
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/companies` | Listar empresas |
| GET | `/api/v1/companies/:id` | Obtener empresa |
| POST | `/api/v1/companies` | Crear empresa |
| PUT | `/api/v1/companies/:id` | Actualizar empresa |
| DELETE | `/api/v1/companies/:id` | Eliminar empresa |

#### **Customers**
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/customers` | Listar clientes |
| GET | `/api/v1/customers/:id` | Obtener cliente |
| POST | `/api/v1/customers` | Crear cliente |
| PUT | `/api/v1/customers/:id` | Actualizar cliente |
| DELETE | `/api/v1/customers/:id` | Eliminar cliente |

#### **Products**
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/products` | Listar productos |
| GET | `/api/v1/products/:id` | Obtener producto |
| POST | `/api/v1/products` | Crear producto |
| PUT | `/api/v1/products/:id` | Actualizar producto |
| DELETE | `/api/v1/products/:id` | Eliminar producto |

#### **Invoices**
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/invoices` | Listar facturas |
| GET | `/api/v1/invoices/:id` | Obtener factura |
| POST | `/api/v1/invoices` | Crear factura |
| PUT | `/api/v1/invoices/:id` | Actualizar factura |
| DELETE | `/api/v1/invoices/:id` | Eliminar factura |
| POST | `/api/v1/invoices/:id/sign` | Firmar factura |
| POST | `/api/v1/invoices/:id/send` | Enviar a DIAN |
| GET | `/api/v1/invoices/:id/pdf` | Generar y visualizar PDF |

## 🔐 Autenticación

### Obtener Token JWT

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Respuesta:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 24
  }
}
```

### Usar Token en Requests

```bash
curl http://localhost:3000/api/v1/companies \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📊 Formato de Respuestas

### Éxito
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error
```json
{
  "success": false,
  "error": "Error message"
}
```

## 🗄️ Base de Datos

### Migraciones

```bash
# Ver estado de migraciones
go run database/cmd/migrate/main.go status

# Ejecutar migraciones pendientes
go run database/cmd/migrate/main.go migrate

# Recrear base de datos desde cero
go run database/cmd/migrate/main.go fresh

# Cargar seeds (catálogos DIAN)
go run database/cmd/migrate/main.go seed
```

### Conexión Independiente

La API tiene su propia conexión a PostgreSQL, **independiente** del sistema de migraciones:
- **Migraciones**: `database/cmd/migrate/main.go`
- **API**: `internal/infrastructure/database/postgres.go`

Ambas usan las mismas credenciales del `.env`.

## 🧪 Testing

```bash
# Ejecutar tests
go test ./...

# Con cobertura
go test -cover ./...

# Verbose
go test -v ./...
```

## 🔧 Solución de Problemas

### Error: "ENCRYPTION_KEY not set in environment"
**Causa:** La variable `ENCRYPTION_KEY` no está configurada en el archivo `.env`  
**Solución:** 
```bash
# Generar clave
openssl rand -hex 32
# Agregar al .env
echo "ENCRYPTION_KEY=tu_clave_generada_aqui" >> .env
```

### Error: "ENCRYPTION_KEY must be 32 bytes (64 hex chars)"
**Causa:** La clave no tiene el formato correcto  
**Solución:** La clave debe ser exactamente 64 caracteres hexadecimales (0-9, a-f). Genera una nueva:
```bash
openssl rand -hex 32
```

### Error: "failed to decrypt certificate password"
**Causa:** Password cifrado con una `ENCRYPTION_KEY` diferente  
**Solución:** 
- Si cambiaste la `ENCRYPTION_KEY`, los passwords existentes no se pueden descifrar
- Debes re-subir los certificados con la nueva clave
- En producción, NUNCA cambies la `ENCRYPTION_KEY` sin migrar los datos

### Error: "ASN.1 syntax error" al cargar certificado
**Causa:** Certificado P12 en formato BER (no DER)  
**Solución:** El sistema convierte automáticamente usando OpenSSL. Asegúrate de tener OpenSSL instalado:
```bash
sudo apt-get install -y openssl
```

### Error: "C14N canonicalization failed"
**Causa:** libxml2-dev no está instalado o pkg-config no encuentra la librería  
**Solución:**
```bash
sudo apt-get install -y libxml2-dev pkg-config
# Luego recompilar
go build ./...
```

### Error: "No tiene suficiente espacio libre en /var/cache/apt/archives/"
**Causa:** Disco lleno o caché de APT ocupando mucho espacio  
**Solución:**
```bash
# Limpiar caché de paquetes
sudo apt-get clean
sudo apt-get autoclean
sudo apt-get autoremove

# Intentar instalar nuevamente
sudo apt-get install -y libxml2-dev pkg-config
```

### Error: "Package libxml-2.0 was not found"
**Causa:** libxml2-dev no está instalado en el sistema  
**Solución:**
```bash
sudo apt-get update
sudo apt-get install -y libxml2-dev pkg-config
```

### Error: "InvalidSecurity" al enviar a DIAN
**Causa:** Certificado no autorizado para servicios SOAP de DIAN  
**Solución:** Contactar con DIAN para autorizar el certificado para WS-Security

## 📝 Notas de Producción

1. **ENCRYPTION_KEY:** 
   - Genera una clave única y fuerte con `openssl rand -hex 32`
   - Guárdala en un password manager seguro
   - NUNCA la subas a Git ni la compartas
   - Si la pierdes, NO podrás descifrar passwords de certificados existentes
   - Usa claves diferentes para desarrollo y producción

2. **Storage:** 
   - Certificados: `storage/app/companies/{NIT}/certificates/`
   - Documentos: `storage/app/companies/{NIT}/documents/{InvoiceNumber}/`
   - Logos: `storage/app/companies/{NIT}/profile/logo.{ext}`
   - Logo por defecto: `storage/app/assets/logo_default.png`

3. **Permisos:** El usuario que ejecuta la aplicación debe tener permisos de escritura en `storage/`

4. **Base de datos:** Configurar conexiones pooling para mejor rendimiento

5. **CORS:** Configurar `CORS_ALLOW_ORIGINS` según dominios permitidos

6. **JWT:** Usar un `JWT_SECRET` fuerte y único en producción

7. **CGO:** La aplicación requiere CGO habilitado para compilar (libxml2). Asegúrate de tener `gcc` instalado

8. **Timezone:** Configurar `TZ=America/Bogota` para zona horaria de Colombia

9. **XMLs sin firmar:** Configurar `KEEP_UNSIGNED_XML=false` en producción para ahorrar espacio

## 📚 Dependencias

### Librerías Go
- [Fiber](https://gofiber.io/) - HTTP framework ultra rápido
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Tokens
- [PostgreSQL Driver](https://github.com/lib/pq) - Driver PostgreSQL
- [godotenv](https://github.com/joho/godotenv) - Cargar variables de entorno
- [Maroto](https://github.com/johnfercher/maroto) - Generación de PDFs nativa en Go
- [beevik/etree](https://github.com/beevik/etree) - Manipulación de XML

### Dependencias del Sistema
- **libxml2-dev** - Canonicalización C14N 1.0 (CGO)
- **pkg-config** - Detección de librerías para CGO
- **OpenSSL** - Conversión de certificados P12
- **build-essential (gcc)** - Compilador C para CGO

## 🤝 Contribuir

Este es un proyecto privado. Para contribuir, contacta al equipo de desarrollo.

## 📄 Licencia

Privado - Todos los derechos reservados.
