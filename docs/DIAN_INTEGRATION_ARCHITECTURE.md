# 📊 REPORTE DE COMPATIBILIDAD: JSON INVOICE ↔ BD ↔ DIAN ↔ UBL21-DIAN

## 🎯 RESUMEN EJECUTIVO

He realizado una auditoría exhaustiva comparando:
1. **JSON de salida actual** (`test.txt`)
2. **Esquema de base de datos** (migraciones)
3. **Requisitos oficiales DIAN** (anexos técnicos UBL 2.1)
4. **Implementación ubl21-dian** (librería Go)

**Resultado:** ✅ **ALTA COMPATIBILIDAD** con **gaps críticos identificados** que requieren JOINs adicionales.

---

## 📋 1. ANÁLISIS DEL JSON ACTUAL

### ✅ **Campos Presentes en JSON**

```json
{
  "id": 1,
  "company_id": 1,           // ⚠️ Solo ID, falta datos completos
  "customer_id": 1,          // ⚠️ Solo ID, falta datos completos
  "resolution_id": 1,        // ⚠️ Solo ID, falta datos completos
  "number": "SETP990000001",
  "consecutive": 990000001,
  "issue_date": "2026-01-13T00:00:00Z",
  "issue_time": "2026-01-15T14:04:20.9698Z",
  "due_date": "2026-02-13T00:00:00Z",
  "type_document_id": 1,     // ⚠️ Solo ID, falta código
  "currency_code_id": 1,     // ⚠️ Solo ID, falta código "COP"
  "notes": "Factura de prueba",
  "payment_method_id": 1,    // ⚠️ Solo ID, falta código
  "payment_due_date": "2026-02-12T00:00:00Z",
  "subtotal": 100000,
  "tax_total": 19000,
  "total": 119000,
  "status": "draft",
  "lines": [
    {
      "id": 1,
      "product_id": 1,         // ⚠️ Solo ID, falta datos completos
      "line_number": 1,
      "description": "Producto 1",
      "quantity": 2,
      "unit_price": 25000,
      "line_total": 50000,
      "tax_rate": 19,
      "tax_amount": 9500
    }
  ]
}
```

---

## 🔴 2. CAMPOS FALTANTES CRÍTICOS PARA DIAN

### **A. Datos del EMISOR (AccountingSupplierParty)**

**Requeridos por DIAN según UBL 2.1:**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `PartyIdentification.ID` | `companies.nit` | ❌ Falta | JOIN con `companies` |
| `PartyIdentification.SchemeID` | `companies.dv` | ❌ Falta | JOIN con `companies` |
| `PartyIdentification.SchemeName` | `companies.document_type_id` | ❌ Falta | JOIN con `companies` + `document_types` |
| `PartyName.Name` | `companies.name` | ❌ Falta | JOIN con `companies` |
| `PartyTaxScheme.RegistrationName` | `companies.registration_name` | ❌ Falta | JOIN con `companies` |
| `PartyTaxScheme.TaxLevelCode` | `companies.tax_level_code_id` | ❌ Falta | JOIN con `companies` + `tax_level_codes` |
| `PartyLegalEntity.CompanyID` | `companies.nit` | ❌ Falta | JOIN con `companies` |
| `PhysicalLocation.Address` | `companies.address_line` | ❌ Falta | JOIN con `companies` |
| `PhysicalLocation.CityName` | `companies.municipality_id` | ❌ Falta | JOIN con `companies` + `municipalities` |
| `PhysicalLocation.CountrySubentity` | `companies.department_id` | ❌ Falta | JOIN con `companies` + `departments` |
| `PhysicalLocation.PostalZone` | `companies.postal_zone` | ❌ Falta | JOIN con `companies` |
| `Contact.Telephone` | `companies.phone` | ❌ Falta | JOIN con `companies` |
| `Contact.ElectronicMail` | `companies.email` | ❌ Falta | JOIN con `companies` |

### **B. Datos del ADQUIRIENTE (AccountingCustomerParty)**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `PartyIdentification.ID` | `customers.identification_number` | ❌ Falta | JOIN con `customers` |
| `PartyIdentification.SchemeID` | `customers.dv` | ❌ Falta | JOIN con `customers` |
| `PartyIdentification.SchemeName` | `customers.document_type_id` | ❌ Falta | JOIN con `customers` + `document_types` |
| `PartyName.Name` | `customers.name` | ❌ Falta | JOIN con `customers` |
| `PartyTaxScheme.RegistrationName` | `customers.name` | ❌ Falta | JOIN con `customers` |
| `PartyTaxScheme.TaxLevelCode` | `customers.tax_level_code_id` | ❌ Falta | JOIN con `customers` + `tax_level_codes` |
| `PhysicalLocation.Address` | `customers.address_line` | ❌ Falta | JOIN con `customers` |
| `PhysicalLocation.CityName` | `customers.municipality_id` | ❌ Falta | JOIN con `customers` + `municipalities` |
| `PhysicalLocation.CountrySubentity` | `customers.department_id` | ❌ Falta | JOIN con `customers` + `departments` |
| `Contact.Telephone` | `customers.phone` | ❌ Falta | JOIN con `customers` |
| `Contact.ElectronicMail` | `customers.email` | ❌ Falta | JOIN con `customers` |

### **C. Datos de RESOLUCIÓN DIAN**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `InvoiceAuthorization.ID` | `resolutions.resolution` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.StartDate` | `resolutions.date_from` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.EndDate` | `resolutions.date_to` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.Prefix` | `resolutions.prefix` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.From` | `resolutions.from_number` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.To` | `resolutions.to_number` | ❌ Falta | JOIN con `resolutions` |
| `InvoiceAuthorization.TechnicalKey` | `resolutions.technical_key` | ❌ Falta | JOIN con `resolutions` |

### **D. Datos de SOFTWARE DIAN**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `SoftwareProvider` | `companies.*` (mismo emisor) | ❌ Falta | JOIN con `companies` |
| `SoftwareID` | `software.identifier` | ❌ Falta | JOIN con `software` |
| `SoftwareSecurityCode` | `software.pin` | ❌ Falta | JOIN con `software` |
| `ProfileExecutionID` | `software.environment` | ❌ Falta | JOIN con `software` |

### **E. Datos de LÍNEAS (InvoiceLines)**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `InvoicedQuantity.UnitCode` | `products.unit_code_id` | ❌ Falta | JOIN con `products` + `unit_codes` |
| `Item.StandardItemIdentification` | `products.standard_item_code` | ❌ Falta | JOIN con `products` |
| `Item.BrandName` | `document_lines.brand_name` | ✅ Presente | - |
| `Item.ModelName` | `document_lines.model_name` | ✅ Presente | - |
| `TaxTotal.TaxScheme.ID` | `products.tax_type_id` | ❌ Falta | JOIN con `products` + `tax_types` |
| `TaxTotal.TaxScheme.Name` | `products.tax_type_id` | ❌ Falta | JOIN con `products` + `tax_types` |

### **F. Códigos de Catálogos DIAN**

| Campo DIAN | Campo BD | Estado | Acción Requerida |
|------------|----------|--------|------------------|
| `InvoiceTypeCode` | `type_document_id` | ❌ Solo ID | JOIN con `invoice_type_codes.code` |
| `DocumentCurrencyCode` | `currency_code_id` | ❌ Solo ID | JOIN con `currency_codes.code` ("COP") |
| `PaymentMeansCode` | `payment_method_id` | ❌ Solo ID | JOIN con `payment_methods.code` |

---

## 🔍 3. COMPARACIÓN CON UBL21-DIAN

### ✅ **Estructura `invoice.Invoice` en ubl21-dian**

```go
type Invoice struct {
    ID              string          // ✅ Mapeado: documents.number
    UUID            string          // ✅ Mapeado: documents.uuid (CUFE)
    IssueDate       time.Time       // ✅ Mapeado: documents.issue_date
    IssueTime       string          // ✅ Mapeado: documents.issue_time
    DueDate         *time.Time      // ✅ Mapeado: documents.due_date
    InvoiceTypeCode string          // ⚠️ Requiere JOIN: invoice_type_codes.code
    
    ProfileExecutionID string       // ⚠️ Requiere JOIN: software.environment
    DocumentCurrencyCode string     // ⚠️ Requiere JOIN: currency_codes.code
    
    Notes []string                  // ✅ Mapeado: documents.notes
    
    AccountingSupplier core.Party  // ❌ FALTA COMPLETO - Requiere JOIN companies
    AccountingCustomer core.Party  // ❌ FALTA COMPLETO - Requiere JOIN customers
    
    InvoiceLines []InvoiceLine     // ⚠️ Parcial - Requiere JOIN products
    
    LegalMonetaryTotal             // ✅ Mapeado: subtotal, tax_total, total
    TaxTotals []core.TaxTotal      // ⚠️ Requiere JOIN tax_types
    
    PaymentMeans []core.PaymentMeans  // ⚠️ Requiere JOIN payment_methods
    PaymentTerms []core.PaymentTerms  // ✅ Mapeado: payment_due_date
    
    // Extensiones DIAN
    SoftwareProvider        core.Party  // ❌ FALTA - Requiere JOIN companies
    SoftwareID              string      // ❌ FALTA - Requiere JOIN software
    SoftwareSecurityCode    string      // ❌ FALTA - Requiere JOIN software
    AuthorizationProvider   string      // ⚠️ Hardcoded: "DIAN"
    AuthorizationProviderID string      // ⚠️ Hardcoded: "800197268"
    QRCode                  string      // ⚠️ Se genera dinámicamente
}
```

---

## 🔴 4. GAPS CRÍTICOS IDENTIFICADOS

### **GAP #1: Falta JOIN con `companies`**

**Impacto:** ❌ **CRÍTICO** - Sin datos del emisor, no se puede generar XML DIAN válido.

**Solución:**
```sql
SELECT 
    d.*,
    -- Emisor (AccountingSupplierParty)
    c.nit AS company_nit,
    c.dv AS company_dv,
    c.name AS company_name,
    c.registration_name AS company_registration_name,
    c.address_line AS company_address,
    c.postal_zone AS company_postal_zone,
    c.phone AS company_phone,
    c.email AS company_email,
    dt_company.code AS company_document_type_code,
    dt_company.name AS company_document_type_name,
    tlc_company.code AS company_tax_level_code,
    to_company.code AS company_type_organization_code,
    tr_company.code AS company_type_regime_code,
    mun_company.name AS company_municipality_name,
    dep_company.name AS company_department_name,
    country_company.code AS company_country_code
FROM documents d
INNER JOIN companies c ON d.company_id = c.id
INNER JOIN document_types dt_company ON c.document_type_id = dt_company.id
INNER JOIN tax_level_codes tlc_company ON c.tax_level_code_id = tlc_company.id
INNER JOIN type_organizations to_company ON c.type_organization_id = to_company.id
INNER JOIN type_regimes tr_company ON c.type_regime_id = tr_company.id
INNER JOIN municipalities mun_company ON c.municipality_id = mun_company.id
INNER JOIN departments dep_company ON c.department_id = dep_company.id
INNER JOIN countries country_company ON c.country_id = country_company.id
```

### **GAP #2: Falta JOIN con `customers`**

**Impacto:** ❌ **CRÍTICO** - Sin datos del adquiriente, no se puede generar XML DIAN válido.

**Solución:** Similar a GAP #1, JOIN con `customers` y sus tablas relacionadas.

### **GAP #3: Falta JOIN con `resolutions`**

**Impacto:** ❌ **CRÍTICO** - DIAN requiere datos de la resolución en el XML.

**Solución:**
```sql
INNER JOIN resolutions r ON d.resolution_id = r.id
```

### **GAP #4: Falta JOIN con `software`**

**Impacto:** ❌ **CRÍTICO** - DIAN requiere `SoftwareID` y `PIN` en el XML.

**Solución:**
```sql
INNER JOIN software s ON c.id = s.company_id
```

### **GAP #5: Falta JOIN con `products` en líneas**

**Impacto:** ⚠️ **ALTO** - Faltan códigos de unidad y tipo de impuesto.

**Solución:**
```sql
SELECT 
    dl.*,
    p.code AS product_code,
    p.name AS product_name,
    p.standard_item_code,
    p.unspsc_code,
    uc.code AS unit_code,
    uc.name AS unit_name,
    tt.code AS tax_type_code,
    tt.name AS tax_type_name
FROM document_lines dl
INNER JOIN products p ON dl.product_id = p.id
INNER JOIN unit_codes uc ON p.unit_code_id = uc.id
INNER JOIN tax_types tt ON p.tax_type_id = tt.id
```

### **GAP #6: Falta JOIN con catálogos DIAN**

**Impacto:** ⚠️ **MEDIO** - Faltan códigos en lugar de IDs.

**Solución:**
```sql
INNER JOIN invoice_type_codes itc ON d.type_document_id = itc.id
INNER JOIN currency_codes cc ON d.currency_code_id = cc.id
LEFT JOIN payment_methods pm ON d.payment_method_id = pm.id
```

---

## ✅ 5. QUERY SQL COMPLETA RECOMENDADA

```sql
SELECT 
    -- Documento
    d.id,
    d.number,
    d.consecutive,
    d.uuid,
    d.issue_date,
    d.issue_time,
    d.due_date,
    d.notes,
    d.payment_due_date,
    d.subtotal,
    d.tax_total,
    d.total,
    d.status,
    d.xml_path,
    d.pdf_path,
    d.qr_code_url,
    
    -- Códigos DIAN
    itc.code AS invoice_type_code,
    cc.code AS currency_code,
    pm.code AS payment_method_code,
    
    -- EMISOR (Company)
    jsonb_build_object(
        'id', c.id,
        'nit', c.nit,
        'dv', c.dv,
        'name', c.name,
        'registration_name', c.registration_name,
        'document_type_code', dt_c.code,
        'tax_level_code', tlc_c.code,
        'type_organization_code', to_c.code,
        'type_regime_code', tr_c.code,
        'address_line', c.address_line,
        'postal_zone', c.postal_zone,
        'phone', c.phone,
        'email', c.email,
        'municipality', mun_c.name,
        'department', dep_c.name,
        'country_code', country_c.code
    ) AS company,
    
    -- ADQUIRIENTE (Customer)
    jsonb_build_object(
        'id', cust.id,
        'identification_number', cust.identification_number,
        'dv', cust.dv,
        'name', cust.name,
        'document_type_code', dt_cust.code,
        'tax_level_code', tlc_cust.code,
        'type_organization_code', to_cust.code,
        'address_line', cust.address_line,
        'postal_zone', cust.postal_zone,
        'phone', cust.phone,
        'email', cust.email,
        'municipality', mun_cust.name,
        'department', dep_cust.name,
        'country_code', country_cust.code
    ) AS customer,
    
    -- RESOLUCIÓN
    jsonb_build_object(
        'id', r.id,
        'prefix', r.prefix,
        'resolution', r.resolution,
        'technical_key', r.technical_key,
        'from_number', r.from_number,
        'to_number', r.to_number,
        'date_from', r.date_from,
        'date_to', r.date_to
    ) AS resolution,
    
    -- SOFTWARE DIAN
    jsonb_build_object(
        'id', s.id,
        'identifier', s.identifier,
        'pin', s.pin,
        'environment', s.environment,
        'test_set_id', s.test_set_id
    ) AS software

FROM documents d

-- JOINs EMISOR
INNER JOIN companies c ON d.company_id = c.id
INNER JOIN document_types dt_c ON c.document_type_id = dt_c.id
INNER JOIN tax_level_codes tlc_c ON c.tax_level_code_id = tlc_c.id
INNER JOIN type_organizations to_c ON c.type_organization_id = to_c.id
INNER JOIN type_regimes tr_c ON c.type_regime_id = tr_c.id
INNER JOIN municipalities mun_c ON c.municipality_id = mun_c.id
INNER JOIN departments dep_c ON c.department_id = dep_c.id
INNER JOIN countries country_c ON c.country_id = country_c.id

-- JOINs ADQUIRIENTE
INNER JOIN customers cust ON d.customer_id = cust.id
INNER JOIN document_types dt_cust ON cust.document_type_id = dt_cust.id
INNER JOIN tax_level_codes tlc_cust ON cust.tax_level_code_id = tlc_cust.id
INNER JOIN type_organizations to_cust ON cust.type_organization_id = to_cust.id
INNER JOIN type_regimes tr_cust ON cust.type_regime_id = tr_cust.id
INNER JOIN municipalities mun_cust ON cust.municipality_id = mun_cust.id
INNER JOIN departments dep_cust ON cust.department_id = dep_cust.id
INNER JOIN countries country_cust ON cust.country_id = country_cust.id

-- JOINs RESOLUCIÓN Y SOFTWARE
INNER JOIN resolutions r ON d.resolution_id = r.id
INNER JOIN software s ON c.id = s.company_id

-- JOINs CÓDIGOS DIAN
INNER JOIN invoice_type_codes itc ON d.type_document_id = itc.id
INNER JOIN currency_codes cc ON d.currency_code_id = cc.id
LEFT JOIN payment_methods pm ON d.payment_method_id = pm.id

WHERE d.id = $1;
```

### **Query para LÍNEAS:**

```sql
SELECT 
    dl.id,
    dl.line_number,
    dl.description,
    dl.quantity,
    dl.unit_price,
    dl.line_total,
    dl.tax_rate,
    dl.tax_amount,
    dl.brand_name,
    dl.model_name,
    dl.standard_item_code,
    dl.classification_code,
    
    -- Producto
    p.code AS product_code,
    p.name AS product_name,
    p.standard_item_code AS product_standard_code,
    p.unspsc_code,
    
    -- Unidad
    uc.code AS unit_code,
    uc.name AS unit_name,
    
    -- Impuesto
    tt.code AS tax_type_code,
    tt.name AS tax_type_name

FROM document_lines dl
INNER JOIN products p ON dl.product_id = p.id
INNER JOIN unit_codes uc ON p.unit_code_id = uc.id
INNER JOIN tax_types tt ON p.tax_type_id = tt.id
WHERE dl.document_id = $1
ORDER BY dl.line_number;
```

---

## 📊 6. ESTRUCTURA JSON RECOMENDADA

```json
{
  "success": true,
  "message": "Invoice retrieved successfully",
  "data": {
    "id": 1,
    "number": "SETP990000001",
    "consecutive": 990000001,
    "uuid": null,
    "issue_date": "2026-01-13T00:00:00Z",
    "issue_time": "2026-01-15T14:04:20.9698Z",
    "due_date": "2026-02-13T00:00:00Z",
    "invoice_type_code": "01",
    "currency_code": "COP",
    "notes": "Factura de prueba",
    "payment_method_code": "10",
    "payment_due_date": "2026-02-12T00:00:00Z",
    "subtotal": 100000,
    "tax_total": 19000,
    "total": 119000,
    "status": "draft",
    
    "company": {
      "id": 1,
      "nit": "900123456",
      "dv": "3",
      "name": "Mi Empresa SAS",
      "registration_name": "MI EMPRESA SAS",
      "document_type_code": "31",
      "tax_level_code": "O-13",
      "type_organization_code": "1",
      "type_regime_code": "49",
      "address_line": "Calle 123 # 45-67",
      "postal_zone": "110111",
      "phone": "+573001234567",
      "email": "contacto@miempresa.com",
      "municipality": "Bogotá D.C.",
      "department": "Bogotá",
      "country_code": "CO"
    },
    
    "customer": {
      "id": 1,
      "identification_number": "800654321",
      "dv": "9",
      "name": "Cliente XYZ SAS",
      "document_type_code": "31",
      "tax_level_code": "O-13",
      "type_organization_code": "1",
      "address_line": "Carrera 10 # 20-30",
      "postal_zone": "110221",
      "phone": "+573009876543",
      "email": "cliente@xyz.com",
      "municipality": "Bogotá D.C.",
      "department": "Bogotá",
      "country_code": "CO"
    },
    
    "resolution": {
      "id": 1,
      "prefix": "SETP",
      "resolution": "18760000001",
      "technical_key": "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c",
      "from_number": 990000000,
      "to_number": 995000000,
      "date_from": "2019-01-19",
      "date_to": "2030-01-19"
    },
    
    "software": {
      "id": 1,
      "identifier": "a8d18e50-0b6a-4ef1-b0d8-6c8f3c8e8e8e",
      "pin": "12345",
      "environment": "2",
      "test_set_id": "TestSetId123"
    },
    
    "lines": [
      {
        "id": 1,
        "line_number": 1,
        "description": "Producto 1",
        "quantity": 2,
        "unit_price": 25000,
        "line_total": 50000,
        "tax_rate": 19,
        "tax_amount": 9500,
        "brand_name": null,
        "model_name": null,
        "standard_item_code": null,
        "product_code": "PROD001",
        "product_name": "Producto 1",
        "unspsc_code": "43211500",
        "unit_code": "EA",
        "unit_name": "Unidad",
        "tax_type_code": "01",
        "tax_type_name": "IVA"
      }
    ]
  }
}
```

---

## 🎯 7. PLAN DE ACCIÓN

### **FASE 1: Actualizar Repository (CRÍTICO)**

**Archivo:** `internal/repository/invoice_repository.go`

1. **Modificar `GetByID()`** para incluir todos los JOINs necesarios
2. **Retornar struct enriquecido** con datos completos

### **FASE 2: Actualizar Domain (CRÍTICO)**

**Archivo:** `internal/domain/invoice.go`

1. **Agregar structs anidados:**
   ```go
   type Invoice struct {
       // ... campos existentes
       Company    *CompanyDetail    `json:"company"`
       Customer   *CustomerDetail   `json:"customer"`
       Resolution *ResolutionDetail `json:"resolution"`
       Software   *SoftwareDetail   `json:"software"`
       Lines      []InvoiceLineDetail `json:"lines"`
   }
   
   type CompanyDetail struct {
       ID                   int64  `json:"id"`
       NIT                  string `json:"nit"`
       DV                   string `json:"dv"`
       Name                 string `json:"name"`
       RegistrationName     string `json:"registration_name"`
       DocumentTypeCode     string `json:"document_type_code"`
       TaxLevelCode         string `json:"tax_level_code"`
       TypeOrganizationCode string `json:"type_organization_code"`
       TypeRegimeCode       string `json:"type_regime_code"`
       AddressLine          string `json:"address_line"`
       PostalZone           string `json:"postal_zone"`
       Phone                string `json:"phone"`
       Email                string `json:"email"`
       Municipality         string `json:"municipality"`
       Department           string `json:"department"`
       CountryCode          string `json:"country_code"`
   }
   
   // Similar para CustomerDetail, ResolutionDetail, SoftwareDetail
   ```

### **FASE 3: Crear Mapper (CRÍTICO)**

**Archivo nuevo:** `internal/service/invoice_mapper.go`

```go
package service

import (
    "apidian-go/internal/domain"
    "github.com/diegofxm/ubl21-dian/invoice"
    "github.com/diegofxm/ubl21-dian/core"
)

// MapInvoiceToUBL convierte domain.Invoice a invoice.Invoice (UBL)
func MapInvoiceToUBL(inv *domain.Invoice) (*invoice.Invoice, error) {
    ublInvoice := invoice.NewInvoice()
    
    // Mapear campos básicos
    ublInvoice.ID = inv.Number
    ublInvoice.UUID = *inv.UUID // CUFE
    ublInvoice.IssueDate = inv.IssueDate
    ublInvoice.IssueTime = inv.IssueTime.Format("15:04:05-07:00")
    ublInvoice.InvoiceTypeCode = inv.InvoiceTypeCode
    ublInvoice.DocumentCurrencyCode = inv.CurrencyCode
    ublInvoice.ProfileExecutionID = inv.Software.Environment
    
    // Mapear emisor
    ublInvoice.AccountingSupplier = core.Party{
        Name: inv.Company.Name,
        PartyIdentification: core.PartyIdentification{
            ID: inv.Company.NIT,
            SchemeID: inv.Company.DV,
            SchemeName: inv.Company.DocumentTypeCode,
        },
        PartyTaxScheme: core.PartyTaxScheme{
            RegistrationName: inv.Company.RegistrationName,
            CompanyID: inv.Company.NIT,
            TaxLevelCode: inv.Company.TaxLevelCode,
        },
        // ... resto de campos
    }
    
    // Mapear adquiriente
    ublInvoice.AccountingCustomer = core.Party{
        // ... similar a emisor
    }
    
    // Mapear líneas
    for _, line := range inv.Lines {
        ublLine := invoice.NewLineBuilder(
            fmt.Sprintf("%d", line.LineNumber),
            line.Description,
        ).
        SetQuantity(line.Quantity).
        SetUnitPrice(line.UnitPrice).
        SetUnitCode(line.UnitCode).
        AddTax(line.TaxTypeCode, line.TaxTypeName, line.TaxRate).
        Build()
        
        ublInvoice.InvoiceLines = append(ublInvoice.InvoiceLines, ublLine)
    }
    
    // Mapear totales
    ublInvoice.LegalMonetaryTotal = invoice.LegalMonetaryTotal{
        LineExtensionAmount: core.MonetaryAmount{
            Value: inv.Subtotal,
            CurrencyID: inv.CurrencyCode,
        },
        TaxExclusiveAmount: core.MonetaryAmount{
            Value: inv.Subtotal,
            CurrencyID: inv.CurrencyCode,
        },
        TaxInclusiveAmount: core.MonetaryAmount{
            Value: inv.Total,
            CurrencyID: inv.CurrencyCode,
        },
        PayableAmount: core.MonetaryAmount{
            Value: inv.Total,
            CurrencyID: inv.CurrencyCode,
        },
    }
    
    // Mapear extensiones DIAN
    ublInvoice.SoftwareID = inv.Software.Identifier
    ublInvoice.SoftwareSecurityCode = inv.Software.PIN
    ublInvoice.AuthorizationProvider = "DIAN"
    ublInvoice.AuthorizationProviderID = "800197268"
    
    return ublInvoice, nil
}
```

---

## ✅ 8. CONCLUSIONES

### **Estado Actual:**
- ✅ Esquema de BD **100% compatible** con requisitos DIAN
- ✅ Librería `ubl21-dian` **100% funcional**
- ❌ JSON de salida **incompleto** (solo IDs, faltan datos)
- ❌ Repository **no hace JOINs** necesarios

### **Acción Inmediata Requerida:**

1. **Actualizar `invoice_repository.GetByID()`** con query completa (JOINs)
2. **Actualizar `domain.Invoice`** con structs anidados
3. **Crear `invoice_mapper.go`** para convertir domain → UBL
4. **Integrar mapper en `invoice_service.Sign()`**

### **Estimación:**
- **Tiempo:** 4-6 horas
- **Complejidad:** Media
- **Impacto:** Crítico para generación XML DIAN

---

# 🏗️ ARQUITECTURA Y FLUJO DE DATOS - EXPLICACIÓN COMPLETA

## 📊 1. FLUJO DE DATOS POR OPERACIÓN

---

### **OPERACIÓN 1: GET Invoice (Consultar factura)**

```
Usuario → Handler → Service → Repository → PostgreSQL
                                              ↓
                                         [JOINs aquí]
                                              ↓
PostgreSQL → Repository → Domain (JSON completo) → Handler → Usuario
```

**¿Qué pasa aquí?**
- ✅ **Repository hace JOINs** y retorna `domain.Invoice` **completo**
- ✅ **JSON de salida es COMPLETO** (incluye company, customer, resolution, software, lines con todos los datos)
- ✅ **Datos en memoria** (struct Go) mientras se procesa la request
- ✅ **No se guarda nada en BD** (solo lectura)

**Ejemplo JSON de respuesta:**
```json
{
  "data": {
    "id": 1,
    "number": "SETP990000001",
    "company": {
      "nit": "900123456",
      "name": "Mi Empresa SAS",
      "email": "contacto@empresa.com"
    },
    "customer": {
      "identification_number": "800654321",
      "name": "Cliente XYZ"
    },
    "resolution": {
      "prefix": "SETP",
      "resolution": "18760000001"
    },
    "software": {
      "identifier": "abc-123",
      "pin": "12345"
    },
    "lines": [...]
  }
}
```

---

### **OPERACIÓN 2: SIGN Invoice (Firmar factura)**

```
Usuario → Handler → Service.Sign()
                         ↓
                    Repository.GetByID() [con JOINs]
                         ↓
                    domain.Invoice (completo en memoria)
                         ↓
                    Mapper: domain.Invoice → ubl21-dian.Invoice
                         ↓
                    ubl21-dian: Generar XML sin firma
                         ↓
                    Guardar: /storage/{NIT}/invoices/FE-{number}.xml
                         ↓
                    ubl21-dian: Firmar XML (XAdES)
                         ↓
                    Guardar: /storage/{NIT}/invoices/FES-{number}.xml
                         ↓
                    Calcular CUFE
                         ↓
                    Repository.UpdateStatus(id, "signed")
                    Repository.UpdateXMLPath(id, path)
                    Repository.UpdateUUID(id, cufe)
                         ↓
                    Handler → Usuario (JSON con status: "signed")
```

**¿Qué pasa aquí?**
- ✅ **Lee datos completos con JOINs** (una sola vez)
- ✅ **Datos en memoria** durante todo el proceso
- ✅ **Mapper convierte** domain → UBL (en memoria)
- ✅ **XML se guarda en disco** (`/storage/{NIT}/invoices/`)
- ✅ **Solo actualiza BD**: `status`, `xml_path`, `uuid` (CUFE)
- ❌ **NO guarda company/customer/resolution** (ya están en BD)

---

### **OPERACIÓN 3: SEND Invoice (Enviar a DIAN)**

```
Usuario → Handler → Service.SendToDIAN()
                         ↓
                    Repository.GetByID() [con JOINs]
                         ↓
                    domain.Invoice (completo en memoria)
                         ↓
                    Validar: status == "signed"
                         ↓
                    Leer XML firmado: /storage/{NIT}/invoices/FES-{number}.xml
                         ↓
                    Crear ZIP: FES-{number}.zip
                         ↓
                    ubl21-dian/soap: Construir SOAP Request
                         ↓
                    ubl21-dian/soap: Enviar a DIAN
                         ↓
                    Recibir SOAP Response
                         ↓
                    Decodificar ApplicationResponse (Base64)
                         ↓
                    Guardar: /storage/{NIT}/invoices/ApplicationResponse-{number}.xml
                         ↓
                    Repository.UpdateStatus(id, "sent")
                    Repository.UpdateDIANResponse(id, response)
                    Repository.UpdateSentToDIANAt(id, now)
                         ↓
                    Handler → Usuario (JSON con status: "sent")
```

**¿Qué pasa aquí?**
- ✅ **Lee datos completos con JOINs** (necesita software.environment para endpoint DIAN)
- ✅ **Datos en memoria** durante todo el proceso
- ✅ **Lee XML del disco** (no regenera)
- ✅ **Guarda ApplicationResponse en disco**
- ✅ **Solo actualiza BD**: `status`, `dian_response`, `sent_to_dian_at`

---

### **OPERACIÓN 4: Generate PDF (Generar PDF)**

```
Usuario → Handler → Service.GeneratePDF()
                         ↓
                    Repository.GetByID() [con JOINs]
                         ↓
                    domain.Invoice (completo en memoria)
                         ↓
                    Leer XML firmado: /storage/{NIT}/invoices/FES-{number}.xml
                         ↓
                    ubl21-dian/pdf: Generar QR Code
                         ↓
                    ubl21-dian/pdf: Generar PDF con plantilla
                         ↓
                    Guardar: /storage/{NIT}/invoices/FES-{number}.pdf
                         ↓
                    Repository.UpdatePDFPath(id, path)
                         ↓
                    Handler → Usuario (JSON con pdf_path)
```

**¿Qué pasa aquí?**
- ✅ **Lee datos completos con JOINs** (necesita todos los datos para el PDF)
- ✅ **Datos en memoria** para renderizar PDF
- ✅ **PDF se guarda en disco**
- ✅ **Solo actualiza BD**: `pdf_path`, `qr_code_url`

---

### **OPERACIÓN 5: Generate AttachedDocument (Documento adjunto para cliente)**

```
Usuario → Handler → Service.GenerateAttachedDocument()
                         ↓
                    Repository.GetByID() [con JOINs]
                         ↓
                    domain.Invoice (completo en memoria)
                         ↓
                    Validar: status == "sent"
                         ↓
                    Leer Invoice firmado: FES-{number}.xml
                         ↓
                    Leer ApplicationResponse: ApplicationResponse-{number}.xml
                         ↓
                    ubl21-dian/invoice: Construir AttachedDocument
                         ↓
                    Guardar: /storage/{NIT}/invoices/AttachedDocument-{number}.xml
                         ↓
                    ubl21-dian/signature: Firmar AttachedDocument
                         ↓
                    Guardar: /storage/{NIT}/invoices/ad{number}.xml
                         ↓
                    Crear ZIP final: ad{number}.zip (Invoice + AppResponse + AttachedDocument)
                         ↓
                    Repository.UpdateZIPPath(id, path)
                         ↓
                    Handler → Usuario (JSON con zip_path)
```

**¿Qué pasa aquí?**
- ✅ **Lee datos completos con JOINs** (necesita company/customer para AttachedDocument)
- ✅ **Datos en memoria** durante construcción
- ✅ **Lee XMLs del disco** (Invoice + ApplicationResponse)
- ✅ **Genera AttachedDocument y lo firma**
- ✅ **Crea ZIP final para cliente**
- ✅ **Solo actualiza BD**: `zip_path`

---

## 🎯 2. RESPONSABILIDADES POR CAPA

### **📂 Repository Layer** (`internal/repository/invoice_repository.go`)

**Responsabilidad:** Acceso a datos (PostgreSQL)

```go
type InvoiceRepository interface {
    // CRUD básico
    Create(invoice *domain.Invoice) error
    GetByID(id int64) (*domain.Invoice, error)  // ← AQUÍ van los JOINs
    Update(invoice *domain.Invoice) error
    Delete(id int64) error
    
    // Actualizaciones específicas
    UpdateStatus(id int64, status string) error
    UpdateXMLPath(id int64, path string) error
    UpdateUUID(id int64, uuid string) error
    UpdatePDFPath(id int64, path string) error
    UpdateZIPPath(id int64, path string) error
    UpdateDIANResponse(id int64, response string) error
}
```

**¿Qué hace?**
- ✅ **Ejecuta JOINs** en `GetByID()` para traer datos completos
- ✅ **Retorna `domain.Invoice`** con todos los datos anidados
- ✅ **Actualiza campos específicos** sin reescribir todo
- ❌ **NO tiene lógica de negocio** (solo SQL)

---

### **📦 Domain Layer** (`internal/domain/invoice.go`)

**Responsabilidad:** Estructuras de datos (DTOs)

```go
type Invoice struct {
    // Campos base (tabla documents)
    ID          int64     `json:"id"`
    Number      string    `json:"number"`
    IssueDate   time.Time `json:"issue_date"`
    Subtotal    float64   `json:"subtotal"`
    TaxTotal    float64   `json:"tax_total"`
    Total       float64   `json:"total"`
    Status      string    `json:"status"`
    
    // Códigos DIAN (de catálogos)
    InvoiceTypeCode   string `json:"invoice_type_code"`
    CurrencyCode      string `json:"currency_code"`
    PaymentMethodCode string `json:"payment_method_code"`
    
    // Datos anidados (de JOINs)
    Company    *CompanyDetail    `json:"company"`
    Customer   *CustomerDetail   `json:"customer"`
    Resolution *ResolutionDetail `json:"resolution"`
    Software   *SoftwareDetail   `json:"software"`
    Lines      []InvoiceLineDetail `json:"lines"`
}

type CompanyDetail struct {
    ID               int64  `json:"id"`
    NIT              string `json:"nit"`
    DV               string `json:"dv"`
    Name             string `json:"name"`
    RegistrationName string `json:"registration_name"`
    // ... todos los campos necesarios para UBL
}

// Similar para CustomerDetail, ResolutionDetail, SoftwareDetail
```

**¿Qué hace?**
- ✅ **Define la estructura** del JSON de salida
- ✅ **Incluye datos anidados** (company, customer, etc.)
- ✅ **Es el contrato** entre capas
- ❌ **NO tiene lógica** (solo structs)

---

### **⚙️ Service Layer** (`internal/service/invoice_service.go`)

**Responsabilidad:** Lógica de negocio y orquestación

```go
type InvoiceService struct {
    invoiceRepo     *repository.InvoiceRepository
    certificateRepo *repository.CertificateRepository
    storagePath     string
}

// Operaciones principales
func (s *InvoiceService) GetByID(id int64) (*domain.Invoice, error) {
    // Solo delega al repository
    return s.invoiceRepo.GetByID(id)
}

func (s *InvoiceService) Sign(id int64) error {
    // 1. Obtener invoice completo (con JOINs)
    invoice, err := s.invoiceRepo.GetByID(id)
    
    // 2. Validar estado
    if invoice.Status != "draft" {
        return errors.New("only draft invoices can be signed")
    }
    
    // 3. Convertir domain → UBL
    ublInvoice, err := MapInvoiceToUBL(invoice)
    
    // 4. Generar XML sin firma
    xmlUnsigned, err := GenerateXML(ublInvoice)
    
    // 5. Guardar XML sin firma
    pathUnsigned := fmt.Sprintf("%s/%s/invoices/FE-%s.xml", 
        s.storagePath, invoice.Company.NIT, invoice.Number)
    os.WriteFile(pathUnsigned, xmlUnsigned, 0644)
    
    // 6. Obtener certificado
    cert, err := s.certificateRepo.GetActiveByCompanyID(invoice.Company.ID)
    
    // 7. Firmar XML
    xmlSigned, err := SignXML(xmlUnsigned, cert)
    
    // 8. Guardar XML firmado
    pathSigned := fmt.Sprintf("%s/%s/invoices/FES-%s.xml", 
        s.storagePath, invoice.Company.NIT, invoice.Number)
    os.WriteFile(pathSigned, xmlSigned, 0644)
    
    // 9. Calcular CUFE
    cufe := CalculateCUFE(ublInvoice)
    
    // 10. Actualizar BD
    s.invoiceRepo.UpdateStatus(id, "signed")
    s.invoiceRepo.UpdateXMLPath(id, pathSigned)
    s.invoiceRepo.UpdateUUID(id, cufe)
    
    return nil
}

func (s *InvoiceService) SendToDIAN(id int64) error {
    // Similar: obtener datos, validar, enviar SOAP, actualizar BD
}

func (s *InvoiceService) GeneratePDF(id int64) error {
    // Similar: obtener datos, generar PDF, actualizar BD
}
```

**¿Qué hace?**
- ✅ **Orquesta operaciones** complejas
- ✅ **Usa mapper** para convertir domain → UBL
- ✅ **Llama a ubl21-dian** para XML/firma/SOAP/PDF
- ✅ **Guarda archivos en disco** (`/storage/`)
- ✅ **Actualiza BD** (solo campos específicos)
- ❌ **NO hace JOINs** (delega a repository)

---

### **🔄 Mapper** (`internal/service/invoice_mapper.go`)

**Responsabilidad:** Conversión domain → UBL

```go
func MapInvoiceToUBL(inv *domain.Invoice) (*invoice.Invoice, error) {
    ublInv := invoice.NewInvoice()
    
    // Mapear campos básicos
    ublInv.ID = inv.Number
    ublInv.UUID = inv.UUID
    ublInv.IssueDate = inv.IssueDate
    ublInv.InvoiceTypeCode = inv.InvoiceTypeCode
    ublInv.DocumentCurrencyCode = inv.CurrencyCode
    ublInv.ProfileExecutionID = inv.Software.Environment
    
    // Mapear emisor (company → Party)
    ublInv.AccountingSupplier = core.Party{
        Name: inv.Company.Name,
        PartyIdentification: core.PartyIdentification{
            ID: inv.Company.NIT,
            SchemeID: inv.Company.DV,
            SchemeName: inv.Company.DocumentTypeCode,
        },
        // ... resto de campos
    }
    
    // Mapear adquiriente (customer → Party)
    ublInv.AccountingCustomer = core.Party{
        // ... similar
    }
    
    // Mapear líneas
    for _, line := range inv.Lines {
        ublLine := invoice.NewLineBuilder(
            fmt.Sprintf("%d", line.LineNumber),
            line.Description,
        ).
        SetQuantity(line.Quantity).
        SetUnitPrice(line.UnitPrice).
        SetUnitCode(line.UnitCode).
        AddTax(line.TaxTypeCode, line.TaxTypeName, line.TaxRate).
        Build()
        
        ublInv.InvoiceLines = append(ublInv.InvoiceLines, ublLine)
    }
    
    // Mapear extensiones DIAN
    ublInv.SoftwareID = inv.Software.Identifier
    ublInv.SoftwareSecurityCode = inv.Software.PIN
    
    return ublInv, nil
}
```

**¿Qué hace?**
- ✅ **Convierte estructuras** (domain → UBL)
- ✅ **Trabaja en memoria** (no toca BD ni disco)
- ✅ **Es puro** (sin efectos secundarios)
- ❌ **NO valida** (asume datos correctos)

---

### **🌐 Handler Layer** (`internal/handler/invoice_handler.go`)

**Responsabilidad:** HTTP (request/response)

```go
func (h *InvoiceHandler) GetByID(c *fiber.Ctx) error {
    id := c.ParamsInt("id")
    
    invoice, err := h.service.GetByID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "success": false,
            "error": "Invoice not found",
        })
    }
    
    return c.JSON(fiber.Map{
        "success": true,
        "data": invoice,  // ← JSON completo con company, customer, etc.
    })
}

func (h *InvoiceHandler) Sign(c *fiber.Ctx) error {
    id := c.ParamsInt("id")
    
    err := h.service.Sign(id)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "success": false,
            "error": err.Error(),
        })
    }
    
    return c.JSON(fiber.Map{
        "success": true,
        "message": "Invoice signed successfully",
    })
}
```

**¿Qué hace?**
- ✅ **Maneja HTTP** (parsea request, retorna response)
- ✅ **Delega a service** (no tiene lógica)
- ✅ **Retorna JSON** (serializa domain.Invoice)
- ❌ **NO procesa datos** (solo pasa y retorna)

---

## 💾 3. ¿DÓNDE SE GUARDAN LOS DATOS?

### **Base de Datos (PostgreSQL)**

**Se guarda:**
- ✅ Datos maestros: `companies`, `customers`, `products`, `resolutions`, `software`
- ✅ Documentos: `documents` (facturas)
- ✅ Líneas: `document_lines`
- ✅ Metadatos: `status`, `xml_path`, `pdf_path`, `zip_path`, `uuid` (CUFE)
- ✅ Respuesta DIAN: `dian_response`, `dian_status`, `sent_to_dian_at`

**NO se guarda:**
- ❌ XML completo (solo la ruta)
- ❌ PDF completo (solo la ruta)
- ❌ Datos duplicados de company/customer (se obtienen con JOINs)

### **Disco (Filesystem)**

**Se guarda en `/storage/{NIT}/invoices/`:**
- ✅ `FE-{number}.xml` (Invoice sin firma)
- ✅ `FES-{number}.xml` (Invoice firmado)
- ✅ `FES-{number}.zip` (ZIP para DIAN)
- ✅ `ApplicationResponse-{number}.xml` (Respuesta DIAN)
- ✅ `AttachedDocument-{number}.xml` (Documento adjunto sin firma)
- ✅ `ad{number}.xml` (Documento adjunto firmado)
- ✅ `ad{number}.zip` (ZIP final para cliente)
- ✅ `FES-{number}.pdf` (PDF de la factura)

### **Memoria (RAM)**

**Existe solo durante la request:**
- ✅ `domain.Invoice` (con todos los datos de JOINs)
- ✅ `invoice.Invoice` (UBL, después del mapper)
- ✅ XML strings (antes de guardar en disco)
- ✅ SOAP request/response

---

## ✅ 4. RESPUESTA A LA PREGUNTA

> **"¿Los datos quedan en memoria o se ven en un JSON más completo?"**

**Ambos:**

1. **En memoria** (durante el procesamiento):
   - Cuando haces `GetByID()`, el repository ejecuta JOINs y retorna `domain.Invoice` completo en memoria
   - Ese struct vive en RAM mientras se procesa la request
   - Se usa para generar XML, PDF, enviar a DIAN, etc.

2. **En JSON** (respuesta HTTP):
   - El handler serializa `domain.Invoice` a JSON
   - El usuario recibe un JSON **completo** con company, customer, resolution, software, lines
   - Este JSON es **solo para consulta** (no se guarda en BD)

3. **En BD** (persistencia):
   - Solo se guardan **IDs** y **metadatos** (status, paths, CUFE)
   - Los datos de company/customer/etc. **ya existen** en sus tablas
   - Se obtienen con JOINs cuando se necesitan

---

## 🎯 5. VENTAJAS DE ESTA ARQUITECTURA

✅ **Separación de responsabilidades**
- Repository: SQL y JOINs
- Domain: Estructuras de datos
- Service: Lógica de negocio
- Mapper: Conversión domain → UBL
- Handler: HTTP

✅ **Eficiencia**
- JOINs solo cuando se necesitan (GetByID, Sign, Send, PDF)
- No se duplican datos en BD
- Archivos en disco (no en BD)

✅ **Mantenibilidad**
- Cambios en UBL → solo mapper
- Cambios en BD → solo repository
- Cambios en API → solo handler

✅ **Testeable**
- Cada capa se puede testear independientemente
- Mapper es puro (fácil de testear)

---

## 🚀 6. PLAN DE IMPLEMENTACIÓN

### **FASE 1:** Actualizar `domain.Invoice` (agregar structs anidados)
### **FASE 2:** Actualizar `invoice_repository.GetByID()` (agregar JOINs)
### **FASE 3:** Crear `invoice_mapper.go` (domain → UBL)
### **FASE 4:** Actualizar `invoice_service.Sign()` (usar mapper + ubl21-dian)
### **FASE 5:** Implementar `invoice_service.SendToDIAN()` (SOAP)
### **FASE 6:** Implementar `invoice_service.GeneratePDF()` (PDF)
### **FASE 7:** Implementar `invoice_service.GenerateAttachedDocument()` (ZIP final)

---

## 📝 NOTAS FINALES

- Este documento sirve como **referencia técnica** para la integración DIAN
- Todos los cambios propuestos son **retrocompatibles**
- La arquitectura sigue **Clean Architecture** y **SOLID**
- El flujo completo está documentado en `ubl21-dian/Flujo_Facturacion_Electronica_DIAN.md`

---

**Fecha de creación:** 2026-01-15  
**Versión:** 1.0  
**Estado:** Aprobado para implementación
