# 📋 API Routes Documentation - APIDIAN-GO

## ✅ Estructura FLAT Profesional (Actualizado 2026-01-14)

Todos los endpoints siguen el estándar REST profesional:
- **GET**: Usa `?company_id=1` en query params para filtrar
- **POST/PUT**: Usa `company_id` en JSON body
- **DELETE**: Usa solo el ID del recurso

---

## 🔐 Auth (Públicas)

```bash
POST /api/v1/auth/register
POST /api/v1/auth/login
```

**Ejemplo - Register:**
```json
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Ejemplo - Login:**
```json
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

---

## 🔐 Auth (Protegidas)

```bash
POST /api/v1/auth/logout
GET  /api/v1/auth/me
POST /api/v1/auth/change-password
```

---

## 🏢 Companies

```bash
GET    /api/v1/companies
GET    /api/v1/companies/:id
POST   /api/v1/companies
PUT    /api/v1/companies/:id
DELETE /api/v1/companies/:id
POST   /api/v1/companies/:id/certificate  # ⚠️ DEPRECATED - Use /certificates
```

**Ejemplo - Listar empresas:**
```bash
GET /api/v1/companies?page=1&page_size=10
Authorization: Bearer {token}
```

**Ejemplo - Crear empresa:**
```json
POST /api/v1/companies
Authorization: Bearer {token}

{
  "document_type_id": 31,
  "nit": "900123456",
  "dv": "7",
  "name": "Mi Empresa SAS",
  "registration_name": "Mi Empresa SAS",
  "tax_level_code_id": 1,
  "type_organization_id": 1,
  "type_regime_id": 1,
  "industry_codes": ["4711"],
  "country_id": 46,
  "department_id": 149,
  "municipality_id": 1,
  "address_line": "Calle 123 # 45-67"
}
```

---

## 👥 Customers (FLAT)

```bash
GET    /api/v1/customers?company_id=1
GET    /api/v1/customers/:id
POST   /api/v1/customers
PUT    /api/v1/customers/:id
DELETE /api/v1/customers/:id
```

**Ejemplo - Listar customers:**
```bash
GET /api/v1/customers?company_id=1&page=1&page_size=10
Authorization: Bearer {token}
```

**Ejemplo - Crear customer:**
```json
POST /api/v1/customers
Authorization: Bearer {token}

{
  "company_id": 1,
  "document_type_id": 13,
  "identification_number": "123456789",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "3001234567"
}
```

---

## 📦 Products (FLAT)

```bash
GET    /api/v1/products?company_id=1
GET    /api/v1/products/:id
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id
```

**Ejemplo - Listar products:**
```bash
GET /api/v1/products?company_id=1&page=1&page_size=10
Authorization: Bearer {token}
```

**Ejemplo - Crear product:**
```json
POST /api/v1/products
Authorization: Bearer {token}

{
  "company_id": 1,
  "code": "PROD001",
  "name": "Producto de Prueba",
  "price": 100000,
  "tax_id": 1
}
```

---

## 📄 Invoices (FLAT)

```bash
GET    /api/v1/invoices?company_id=1&status=draft
GET    /api/v1/invoices/:id
POST   /api/v1/invoices
PUT    /api/v1/invoices/:id
DELETE /api/v1/invoices/:id
POST   /api/v1/invoices/:id/sign
POST   /api/v1/invoices/:id/send
```

**Ejemplo - Listar invoices con filtros:**
```bash
GET /api/v1/invoices?company_id=1&status=draft&page=1&page_size=10
Authorization: Bearer {token}
```

**Ejemplo - Crear invoice:**
```json
POST /api/v1/invoices
Authorization: Bearer {token}

{
  "company_id": 1,
  "customer_id": 5,
  "resolution_id": 2,
  "invoice_number": "SETT-1",
  "items": [
    {
      "product_id": 10,
      "quantity": 2,
      "unit_price": 50000
    }
  ]
}
```

---

## 🔐 Certificates (FLAT)

```bash
GET    /api/v1/certificates?company_id=1
GET    /api/v1/certificates/all?company_id=1
GET    /api/v1/certificates/:id
POST   /api/v1/certificates
DELETE /api/v1/certificates/:id
```

**Ejemplo - Obtener certificado activo:**
```bash
GET /api/v1/certificates?company_id=1
Authorization: Bearer {token}
```

**Ejemplo - Subir certificado (JSON con base64):**
```json
POST /api/v1/certificates
Authorization: Bearer {token}
Content-Type: application/json

{
  "company_id": 1,
  "certificate": "MIIKpAIBAzCCCmAGCSqGSIb3DQEHAaCCClEEggpNMIIKSTC...",
  "password": "mi_password_certificado"
}
```

**Ejemplo - Histórico de certificados:**
```bash
GET /api/v1/certificates/all?company_id=1
Authorization: Bearer {token}
```

---

## 📋 Resolutions (FLAT)

```bash
GET    /api/v1/resolutions?company_id=1
GET    /api/v1/resolutions/:id
POST   /api/v1/resolutions
DELETE /api/v1/resolutions/:id
```

**Ejemplo - Listar resolutions:**
```bash
GET /api/v1/resolutions?company_id=1
Authorization: Bearer {token}
```

**Ejemplo - Crear resolution:**
```json
POST /api/v1/resolutions
Authorization: Bearer {token}

{
  "company_id": 1,
  "resolution_number": "18760000001",
  "prefix": "SETT",
  "from": 1,
  "to": 5000,
  "date_from": "2024-01-01",
  "date_to": "2025-12-31"
}
```

---

## 💻 Software (FLAT)

```bash
GET    /api/v1/software?company_id=1
GET    /api/v1/software/:id
POST   /api/v1/software
PUT    /api/v1/software/:id
DELETE /api/v1/software/:id
```

**Ejemplo - Obtener software:**
```bash
GET /api/v1/software?company_id=1
Authorization: Bearer {token}
```

**Ejemplo - Crear software:**
```json
POST /api/v1/software
Authorization: Bearer {token}

{
  "company_id": 1,
  "identifier": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "pin": "12345"
}
```

---

## 👤 Users

```bash
GET    /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

---

## 🏥 System

```bash
GET /health
GET /api/v1/ping
```

---

## 📤 Cómo Subir Certificado en Postman

### Opción 1: JSON con Base64 (Actual - Recomendado)

1. **Convertir .p12 a base64:**
```bash
# En Linux/Mac
base64 -i certificado.p12 | tr -d '\n' > certificado_base64.txt

# En Windows PowerShell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("certificado.p12")) | Out-File -Encoding ASCII certificado_base64.txt
```

2. **En Postman:**
   - Method: `POST`
   - URL: `http://localhost:8080/api/v1/certificates`
   - Headers:
     - `Authorization: Bearer {tu_token}`
     - `Content-Type: application/json`
   - Body → raw → JSON:
```json
{
  "company_id": 1,
  "certificate": "MIIKpAIBAzCCCmAGCSqGSIb3DQEHAaCCClEEggpNMIIKSTC...",
  "password": "mi_password"
}
```

### Opción 2: Multipart/Form-Data (No implementado actualmente)

Si prefieres subir el archivo directamente sin convertir a base64, necesitarías:

**En Postman:**
- Method: `POST`
- URL: `http://localhost:8080/api/v1/certificates/upload`
- Headers:
  - `Authorization: Bearer {tu_token}`
- Body → form-data:
  - `company_id`: `1` (text)
  - `certificate`: `[Select File] certificado.p12` (file)
  - `password`: `mi_password` (text)

**⚠️ NOTA:** Esta opción requiere modificar el handler para aceptar `multipart/form-data`. Actualmente solo acepta JSON con base64.

---

## 🔄 Migración de Rutas Antiguas

| ❌ Ruta Antigua (Deprecated) | ✅ Ruta Nueva (FLAT) |
|------------------------------|---------------------|
| `GET /api/v1/companies/900123456/7/customers` | `GET /api/v1/customers?company_id=1` |
| `POST /api/v1/companies/900123456/7/products` | `POST /api/v1/products` (company_id en body) |
| `GET /api/v1/companies/900123456/7/invoices` | `GET /api/v1/invoices?company_id=1` |
| `GET /api/v1/certificates/company/1` | `GET /api/v1/certificates?company_id=1` |
| `POST /api/v1/companies/1/certificate` | `POST /api/v1/certificates` |

---

## 📝 Notas Importantes

1. ✅ Todos los endpoints requieren autenticación (excepto `/auth/register` y `/auth/login`)
2. ✅ El `company_id` en query params es **obligatorio** para GET
3. ✅ El `company_id` en JSON body es **obligatorio** para POST/PUT
4. ✅ El sistema valida que la empresa pertenezca al usuario autenticado
5. ✅ Paginación estándar: `?page=1&page_size=10`
6. ✅ Certificados se guardan encriptados en la base de datos (mejora de seguridad vs apidian)