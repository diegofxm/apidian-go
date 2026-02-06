# 🗺️ Roadmap: Sistema Completo de Documentos Electrónicos DIAN

**Última actualización:** 29 de enero de 2026

## 📊 Progreso General

### ✅ Métodos SOAP Implementados: 4/15 (27%)
- ✅ SendBillSync
- ✅ SendBillAsync  
- ✅ SendTestSetAsync
- ✅ **GetStatus** (Implementado 29/01/2026)

### 🎯 Estado Actual
- **Facturación básica:** ✅ Funcional
- **Envío a DIAN:** ✅ Funcional (Sync, Async, TestSet)
- **Consulta de estado:** ✅ Funcional (GetStatus)
- **Validación completa:** ✅ Funcional
- **Descuentos/Retenciones:** ⏳ Pendiente
- **Notas Crédito/Débito:** ⏳ Pendiente

---

## 📋 Tabla de Contenidos
- [Progreso General](#progreso-general)
- [Métodos SOAP DIAN](#métodos-soap-dian)
- [Pregunta 1: Campos UBL 2.1 Completos](#pregunta-1-campos-ubl-21-completos)
- [Pregunta 2: Arquitectura Multi-Documento](#pregunta-2-arquitectura-multi-documento)
- [Plan de Acción](#plan-de-acción)

---

## 📡 Métodos SOAP DIAN (15 métodos del WSDL)

### 🔵 Grupo 1: Envío de Documentos (5 métodos)

#### 1. `SendBillAsync` ⭐ **(IMPLEMENTADO - En uso)**
- **Para:** Enviar facturas electrónicas de venta (asíncrono)
- **Uso:** Producción - Envío masivo
- **Respuesta:** TrackId para consultar estado después
- **Estado:** ✅ Implementado en `internal/service/invoice/invoice_service.go:SendToDIAN()`

#### 2. `SendBillSync` ⭐ **(IMPLEMENTADO)**
- **Para:** Enviar facturas electrónicas de venta (síncrono)
- **Uso:** Desarrollo/Testing - Respuesta inmediata
- **Respuesta:** Validación completa en el mismo request
- **Estado:** ✅ Implementado en `internal/service/invoice/invoice_service.go:SendToDIAN()`
- **Nota:** Se usa automáticamente en ambiente de producción cuando no es TestSet

#### 3. `SendTestSetAsync` ⭐ **(IMPLEMENTADO)**
- **Para:** Enviar set de pruebas para habilitación
- **Uso:** Certificación ante DIAN (obligatorio antes de producción)
- **Respuesta:** TrackId + validación de set de pruebas
- **Estado:** ✅ Implementado en `internal/service/invoice/invoice_service.go:SendToDIAN()`
- **Nota:** Se usa automáticamente cuando la factura tiene `test_set_id` configurado

#### 4. `SendBillAttachmentAsync`
- **Para:** Enviar documentos soporte (anexos PDF, imágenes)
- **Uso:** Adjuntar archivos adicionales a facturas
- **Respuesta:** TrackId
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (funcionalidad avanzada)

#### 5. `SendNominaSync`
- **Para:** Enviar nómina electrónica (síncrono)
- **Uso:** Documentos de nómina
- **Respuesta:** Validación inmediata
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (módulo separado)

---

### 🟢 Grupo 2: Consulta de Estado (3 métodos)

#### 6. `GetStatus` ⭐ **(IMPLEMENTADO - CRÍTICO)**
- **Para:** Consultar estado de un documento por TrackId
- **Uso:** Verificar si DIAN aceptó/rechazó el documento
- **Entrada:** TrackId (XmlDocumentKey recibido de SendBillSync/SendTestSetAsync)
- **Respuesta:** 
  - `IsValid` (bool)
  - `StatusCode` (string)
  - `StatusDescription` (string)
  - `StatusMessage` (string)
  - `XmlBase64Bytes` (ApplicationResponse FINAL firmado por DIAN en base64)
- **Estado:** ✅ Implementado en `internal/service/invoice/invoice_status.go:GetInvoiceStatus()`
- **Endpoint:** `POST /api/v1/invoices/:id/status` con body `{"track_id": "..."}`
- **Cambios en BD:**
  - Campo `track_id` agregado a tabla `documents` (VARCHAR 255, nullable)
  - Índice `idx_documents_track_id` para consultas rápidas
  - Se guarda automáticamente al enviar factura a DIAN
- **Flujo:**
  1. Enviar factura: `POST /invoices/:id/send` → Guarda `track_id` automáticamente
  2. Esperar 5-30 segundos (DIAN procesa)
  3. Consultar estado: `POST /invoices/:id/status` con `{"track_id": "..."}`
  4. Sistema actualiza `dian_status` (accepted/rejected) y guarda ApplicationResponse FINAL

#### 7. `GetStatusZip`
- **Para:** Consultar estado y descargar ZIP con ApplicationResponse
- **Uso:** Obtener respuesta completa de DIAN en formato ZIP
- **Respuesta:** ZIP con ApplicationResponse firmado por DIAN
- **Estado:** ⏳ Pendiente
- **Prioridad:** Media (alternativa a GetStatus)

#### 8. `GetStatusEvent`
- **Para:** Consultar estado de eventos (acuse, reclamo, aceptación)
- **Uso:** Verificar eventos de documentos
- **Respuesta:** Estado del evento
- **Estado:** ⏳ Pendiente
- **Prioridad:** Media (después de implementar eventos)

---

### 🟡 Grupo 3: Eventos de Documentos (1 método)

#### 9. `SendEventUpdateStatus`
- **Para:** Enviar eventos de documentos
- **Tipos de eventos:**
  - Acuse de recibo (030)
  - Aceptación expresa (032)
  - Aceptación tácita (033)
  - Rechazo (031)
  - Reclamo (034)
- **Uso:** Receptor notifica al emisor sobre el documento
- **Respuesta:** TrackId del evento
- **Estado:** ⏳ Pendiente
- **Prioridad:** Alta (requerido para flujo completo)

---

### 🟠 Grupo 4: Consultas de Información (6 métodos)

#### 10. `GetNumberingRange`
- **Para:** Consultar rangos de numeración autorizados
- **Uso:** Verificar resoluciones vigentes de un NIT
- **Respuesta:** Lista de rangos activos con fechas de vigencia
- **Estado:** ⏳ Pendiente
- **Prioridad:** Media (útil para validación)

#### 11. `GetXmlByDocumentKey`
- **Para:** Descargar XML de un documento por CUFE/CUDE
- **Uso:** Recuperar documento firmado desde DIAN
- **Respuesta:** XML completo del documento
- **Estado:** ⏳ Pendiente
- **Prioridad:** Media (útil para auditoría)

#### 12. `GetReferenceNotes`
- **Para:** Consultar notas crédito/débito asociadas a una factura
- **Uso:** Ver historial de ajustes de un documento
- **Respuesta:** Lista de notas relacionadas
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (después de implementar notas)

#### 13. `GetDocumentInfo`
- **Para:** Consultar información completa de un documento
- **Uso:** Ver detalles, estado, eventos de un documento
- **Respuesta:** Objeto completo con toda la información
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (funcionalidad avanzada)

#### 14. `GetAcquirer`
- **Para:** Consultar información del adquiriente (comprador)
- **Uso:** Validar datos del receptor antes de facturar
- **Respuesta:** Datos del adquiriente registrados en DIAN
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (validación opcional)

#### 15. `GetExchangeEmails`
- **Para:** Consultar correos de intercambio configurados
- **Uso:** Verificar emails para notificaciones automáticas
- **Respuesta:** Lista de emails configurados
- **Estado:** ⏳ Pendiente
- **Prioridad:** Baja (configuración avanzada)

---

## 📝 Pregunta 1: Campos UBL 2.1 Completos

### 🎯 Situación Actual vs. Completa

**✅ Lo que tienes ahora:**
- Factura básica funcional (emisor, receptor, líneas, totales, impuestos)
- Campos mínimos obligatorios para DIAN
- Estructura UBL 2.1 base

**⏳ Lo que falta para factura COMPLETA:**

### 📦 Campos Adicionales Importantes

#### 1. Descuentos y Cargos
```xml
<cac:AllowanceCharge>
  <cbc:ChargeIndicator>false</cbc:ChargeIndicator> <!-- false=descuento, true=cargo -->
  <cbc:AllowanceChargeReason>Descuento comercial</cbc:AllowanceChargeReason>
  <cbc:MultiplierFactorNumeric>10.00</cbc:MultiplierFactorNumeric> <!-- % -->
  <cbc:Amount currencyID="COP">50000.00</cbc:Amount>
  <cbc:BaseAmount currencyID="COP">500000.00</cbc:BaseAmount>
</cac:AllowanceCharge>
```

**Casos de uso:**
- Descuentos por pronto pago
- Descuentos por volumen
- Descuentos comerciales
- Cargos por transporte
- Cargos por embalaje
- Cargos por seguros

**Cambios en BD:**
```sql
CREATE TABLE document_allowance_charges (
  id BIGSERIAL PRIMARY KEY,
  document_id BIGINT NOT NULL REFERENCES documents(id),
  line_id BIGINT REFERENCES document_lines(id), -- NULL si es a nivel documento
  charge_indicator BOOLEAN NOT NULL, -- false=descuento, true=cargo
  allowance_charge_reason VARCHAR(255),
  multiplier_factor_numeric DECIMAL(15,2), -- porcentaje
  amount DECIMAL(15,2) NOT NULL,
  base_amount DECIMAL(15,2),
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### 2. Retenciones
```xml
<cac:WithholdingTaxTotal>
  <cbc:TaxAmount currencyID="COP">25000.00</cbc:TaxAmount>
  <cac:TaxSubtotal>
    <cbc:TaxableAmount currencyID="COP">500000.00</cbc:TaxableAmount>
    <cbc:TaxAmount currencyID="COP">25000.00</cbc:TaxAmount>
    <cac:TaxCategory>
      <cbc:Percent>5.00</cbc:Percent>
      <cac:TaxScheme>
        <cbc:ID>06</cbc:ID> <!-- Renta -->
        <cbc:Name>ReteRenta</cbc:Name>
      </cac:TaxScheme>
    </cac:TaxCategory>
  </cac:TaxSubtotal>
</cac:WithholdingTaxTotal>
```

**Tipos de retenciones:**
- Retención en la fuente (Renta)
- Retención de IVA
- Retención de ICA
- Retención CREE

**Cambios en BD:**
```sql
CREATE TABLE document_withholding_taxes (
  id BIGSERIAL PRIMARY KEY,
  document_id BIGINT NOT NULL REFERENCES documents(id),
  tax_scheme_id INT NOT NULL REFERENCES tax_schemes(id), -- 06=Renta, 05=IVA, etc.
  taxable_amount DECIMAL(15,2) NOT NULL,
  tax_amount DECIMAL(15,2) NOT NULL,
  percent DECIMAL(5,2) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### 3. Anticipos y Prepagos
```xml
<cac:PrepaidPayment>
  <cbc:ID>ANTICIPO-001</cbc:ID>
  <cbc:PaidAmount currencyID="COP">100000.00</cbc:PaidAmount>
  <cbc:PaidDate>2024-01-15</cbc:PaidDate>
  <cbc:InstructionID>Anticipo del 20%</cbc:InstructionID>
</cac:PrepaidPayment>
```

**Cambios en BD:**
```sql
CREATE TABLE document_prepaid_payments (
  id BIGSERIAL PRIMARY KEY,
  document_id BIGINT NOT NULL REFERENCES documents(id),
  payment_id VARCHAR(50),
  paid_amount DECIMAL(15,2) NOT NULL,
  paid_date DATE NOT NULL,
  instruction_id VARCHAR(255),
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### 4. Información de Entrega
```xml
<cac:Delivery>
  <cbc:ActualDeliveryDate>2024-02-01</cbc:ActualDeliveryDate>
  <cac:DeliveryLocation>
    <cac:Address>
      <cbc:AddressLine>Calle 123 #45-67</cbc:AddressLine>
      <cbc:CityName>Medellín</cbc:CityName>
    </cac:Address>
  </cac:DeliveryLocation>
  <cac:DeliveryTerms>
    <cbc:ID>FOB</cbc:ID> <!-- Incoterm -->
  </cac:DeliveryTerms>
</cac:Delivery>
```

**Cambios en BD:**
```sql
ALTER TABLE documents ADD COLUMN delivery_date DATE;
ALTER TABLE documents ADD COLUMN delivery_address TEXT;
ALTER TABLE documents ADD COLUMN delivery_city_code VARCHAR(10);
ALTER TABLE documents ADD COLUMN delivery_terms_code VARCHAR(10); -- Incoterms
```

#### 5. Medios de Pago Detallados
```xml
<cac:PaymentMeans>
  <cbc:ID>1</cbc:ID>
  <cbc:PaymentMeansCode>42</cbc:PaymentMeansCode> <!-- Transferencia -->
  <cbc:PaymentDueDate>2024-02-15</cbc:PaymentDueDate>
  <cac:PayeeFinancialAccount>
    <cbc:ID>1234567890</cbc:ID>
    <cbc:Name>Cuenta de ahorros</cbc:Name>
    <cac:FinancialInstitutionBranch>
      <cbc:ID>001</cbc:ID>
      <cbc:Name>Bancolombia</cbc:Name>
    </cac:FinancialInstitutionBranch>
  </cac:PayeeFinancialAccount>
</cac:PaymentMeans>
```

**Cambios en BD:**
```sql
CREATE TABLE company_bank_accounts (
  id BIGSERIAL PRIMARY KEY,
  company_id BIGINT NOT NULL REFERENCES companies(id),
  account_number VARCHAR(50) NOT NULL,
  account_type VARCHAR(50),
  bank_code VARCHAR(10),
  bank_name VARCHAR(100),
  is_default BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE documents ADD COLUMN payment_account_id BIGINT REFERENCES company_bank_accounts(id);
```

#### 6. Documentos de Referencia
```xml
<cac:BillingReference>
  <cac:InvoiceDocumentReference>
    <cbc:ID>SETT-001</cbc:ID> <!-- Número de orden de compra -->
    <cbc:UUID>abc123...</cbc:UUID>
    <cbc:IssueDate>2024-01-10</cbc:IssueDate>
  </cac:InvoiceDocumentReference>
</cac:BillingReference>
```

**Cambios en BD:**
```sql
CREATE TABLE document_references (
  id BIGSERIAL PRIMARY KEY,
  source_document_id BIGINT NOT NULL REFERENCES documents(id),
  reference_type VARCHAR(50) NOT NULL, -- 'purchase_order', 'contract', 'despatch', etc.
  reference_number VARCHAR(100) NOT NULL,
  reference_uuid VARCHAR(255),
  reference_date DATE,
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### 7. Información Adicional de Líneas
```xml
<cac:InvoiceLine>
  <!-- ... campos actuales ... -->
  <cac:Item>
    <cac:AdditionalItemProperty>
      <cbc:Name>Color</cbc:Name>
      <cbc:Value>Rojo</cbc:Value>
    </cac:AdditionalItemProperty>
    <cac:AdditionalItemProperty>
      <cbc:Name>Talla</cbc:Name>
      <cbc:Value>M</cbc:Value>
    </cac:AdditionalItemProperty>
  </cac:Item>
</cac:InvoiceLine>
```

**Cambios en BD:**
```sql
CREATE TABLE document_line_properties (
  id BIGSERIAL PRIMARY KEY,
  line_id BIGINT NOT NULL REFERENCES document_lines(id),
  property_name VARCHAR(100) NOT NULL,
  property_value TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Campos adicionales en document_lines
ALTER TABLE document_lines ADD COLUMN discount_amount DECIMAL(15,2) DEFAULT 0;
ALTER TABLE document_lines ADD COLUMN charge_amount DECIMAL(15,2) DEFAULT 0;
ALTER TABLE document_lines ADD COLUMN free_of_charge_indicator BOOLEAN DEFAULT false;
```

---

### 💡 Estrategia de Implementación

#### **Opción A: Incremental (RECOMENDADA)**
```
Fase 1: Factura actual (✅ Ya tienes)
Fase 2: + Descuentos/Cargos (más común)
Fase 3: + Retenciones (obligatorio para muchos)
Fase 4: + Anticipos y referencias
Fase 5: + Campos avanzados (entrega, propiedades)
```

**Ventajas:**
- ✅ Entregas rápidas
- ✅ Validación incremental
- ✅ Menor riesgo
- ✅ Aprendes en el camino

#### **Opción B: Completa desde el inicio**
- Agregar TODOS los campos opcionales ahora
- BD más compleja desde el principio
- Más flexible pero más trabajo inicial

**Mi recomendación:** **Opción A** - Ve agregando campos conforme los necesites. La mayoría de empresas usan solo el 30% de los campos disponibles.

---

## 🏗️ Pregunta 2: Arquitectura Multi-Documento

### 📊 Situación Actual

**En `ubl21-dian`:**
- ✅ Constructor XML específico para Invoice
- ✅ Lógica de firma y canonicalización genérica (reutilizable)

**En `apidian-go`:**
- ✅ Modelo `Invoice` específico
- ✅ Servicio `InvoiceService` específico
- ✅ Handler `InvoiceHandler` específico

### 🎯 Estrategia Recomendada: Arquitectura Polimórfica

#### **Opción 1: Tabla Única Polimórfica (RECOMENDADA)**

```sql
-- Tabla única para TODOS los documentos electrónicos
CREATE TABLE documents (
  id BIGSERIAL PRIMARY KEY,
  type_document_id INT NOT NULL, -- 1=Invoice, 2=CreditNote, 3=DebitNote, 4=Payroll, etc.
  company_id BIGINT NOT NULL,
  customer_id BIGINT, -- NULL para nómina
  resolution_id BIGINT, -- NULL para nómina
  number VARCHAR(50) NOT NULL,
  consecutive BIGINT NOT NULL,
  uuid VARCHAR(255), -- CUFE/CUDE/CUNE
  issue_date DATE NOT NULL,
  issue_time TIME NOT NULL,
  due_date DATE,
  currency_code_id INT NOT NULL,
  notes TEXT,
  payment_method_id INT,
  payment_form_id INT,
  
  -- Totales (comunes a todos)
  subtotal DECIMAL(15,2) NOT NULL,
  tax_total DECIMAL(15,2) NOT NULL,
  total DECIMAL(15,2) NOT NULL,
  
  -- Archivos
  xml_path TEXT,
  pdf_path TEXT,
  zip_path TEXT,
  qr_code_url TEXT,
  
  -- Estado
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  dian_status VARCHAR(50),
  dian_response TEXT,
  dian_status_code VARCHAR(10),
  dian_status_description TEXT,
  sent_to_dian_at TIMESTAMP,
  accepted_by_dian_at TIMESTAMP,
  
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  
  UNIQUE(company_id, type_document_id, number)
);

-- Líneas de detalle (para facturas, notas, etc.)
CREATE TABLE document_lines (
  id BIGSERIAL PRIMARY KEY,
  document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  product_id BIGINT REFERENCES products(id),
  line_number INT NOT NULL,
  description TEXT NOT NULL,
  quantity DECIMAL(15,4) NOT NULL,
  unit_price DECIMAL(15,2) NOT NULL,
  line_total DECIMAL(15,2) NOT NULL,
  tax_rate DECIMAL(5,2) NOT NULL,
  tax_amount DECIMAL(15,2) NOT NULL,
  discount_amount DECIMAL(15,2) DEFAULT 0,
  charge_amount DECIMAL(15,2) DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Referencias entre documentos (para notas que referencian facturas)
CREATE TABLE document_references (
  id BIGSERIAL PRIMARY KEY,
  source_document_id BIGINT NOT NULL REFERENCES documents(id), -- La nota
  referenced_document_id BIGINT NOT NULL REFERENCES documents(id), -- La factura original
  reference_type_code VARCHAR(10) NOT NULL, -- '01'=Factura, '91'=NotaCredito, etc.
  created_at TIMESTAMP DEFAULT NOW()
);
```

**Ventajas:**
- ✅ Un solo flujo de firma, envío, consulta
- ✅ Queries más simples
- ✅ Menos duplicación de código
- ✅ Fácil agregar nuevos tipos de documentos
- ✅ Reportes consolidados simples

**Desventajas:**
- ⚠️ Algunos campos específicos quedan NULL (ej: customer_id en nómina)
- ⚠️ Validaciones por tipo en código (no en BD)

#### **Opción 2: Herencia con Tablas Separadas**

```sql
-- Tabla base
CREATE TABLE documents (
  id BIGSERIAL PRIMARY KEY,
  -- campos comunes
);

-- Tablas específicas
CREATE TABLE invoices (
  document_id BIGINT PRIMARY KEY REFERENCES documents(id),
  -- campos específicos de factura
);

CREATE TABLE credit_notes (
  document_id BIGINT PRIMARY KEY REFERENCES documents(id),
  referenced_invoice_id BIGINT REFERENCES invoices(document_id),
  correction_concept_id INT,
  -- campos específicos
);

CREATE TABLE payrolls (
  document_id BIGINT PRIMARY KEY REFERENCES documents(id),
  employee_id BIGINT,
  period_start DATE,
  period_end DATE,
  -- campos específicos de nómina
);
```

**Ventajas:**
- ✅ Campos específicos bien tipados
- ✅ Validaciones a nivel BD
- ✅ Separación clara

**Desventajas:**
- ❌ Mucha duplicación de código
- ❌ Queries complejas (UNION)
- ❌ Difícil mantener
- ❌ Más tablas = más complejidad

---

### 🔧 Arquitectura en `ubl21-dian` (Constructor XML)

#### **Estrategia: Interface + Implementaciones Específicas**

```go
// Interface genérica para todos los documentos
type DocumentBuilder interface {
    BuildXML() ([]byte, error)
    GetDocumentType() string
    GetDocumentKey() string // CUFE/CUDE/CUNE
    Validate() error
}

// Implementaciones específicas
type InvoiceBuilder struct {
    Invoice *Invoice
    Company *Company
    Customer *Customer
    // ...
}

type CreditNoteBuilder struct {
    CreditNote *CreditNote
    Company *Company
    Customer *Customer
    ReferencedInvoice *Invoice // La factura que corrige
    // ...
}

type DebitNoteBuilder struct {
    DebitNote *DebitNote
    Company *Company
    Customer *Customer
    ReferencedInvoice *Invoice
    // ...
}

type PayrollBuilder struct {
    Payroll *Payroll
    Company *Company
    Employee *Employee
    // ...
}
```

#### **Reutilización de Código:**

| Componente | Reutilizable | Notas |
|------------|--------------|-------|
| Firma digital (XAdES) | ✅ 100% | Mismo proceso para todos |
| Canonicalización C14N | ✅ 100% | Mismo algoritmo |
| Cálculo de hash SHA-256 | ✅ 100% | Mismo algoritmo |
| Estructura base UBL | ✅ 80% | Namespaces, header similar |
| Elementos específicos | ⚠️ Variable | Cada documento tiene particularidades |
| Cálculo de CUFE/CUDE | ✅ 90% | Fórmula similar, cambian algunos campos |

---

### 📋 Documentos DIAN y Similitud con Invoice

| Documento | Similitud | Complejidad | Prioridad | Notas |
|-----------|-----------|-------------|-----------|-------|
| **Invoice** | 100% | - | ✅ Hecho | Base actual |
| **CreditNote** | 90% | Baja | 🔥 Alta | Casi idéntico + BillingReference |
| **DebitNote** | 90% | Baja | 🔥 Alta | Casi idéntico + BillingReference |
| **ApplicationResponse** | 30% | Media | 🔥 Alta | Para eventos (acuse, aceptación) |
| **AttachedDocument** | 40% | Media | ⏳ Media | Wrapper de otros documentos |
| **Payroll** | 20% | Alta | ⏳ Baja | Estructura muy diferente |

#### **CreditNote vs Invoice:**
```
Hereda de Invoice:
✅ Misma estructura base (emisor, receptor, líneas, totales)
✅ Mismo proceso de firma
✅ Mismo cálculo de impuestos
+ BillingReference (factura que corrige)
+ DiscrepancyResponse (razón de la nota)
± Totales pueden ser negativos
± CUDE en lugar de CUFE (fórmula similar)
```

#### **DebitNote vs Invoice:**
```
Hereda de Invoice:
✅ Misma estructura base
✅ Mismo proceso de firma
+ BillingReference (factura que ajusta)
+ DiscrepancyResponse (razón del ajuste)
± CUDE en lugar de CUFE
```

#### **ApplicationResponse (Eventos):**
```
Estructura diferente:
- No tiene líneas de detalle
- No tiene totales
+ DocumentResponse (referencia al documento)
+ Response (código de respuesta: aceptado/rechazado)
+ Note (observaciones)
± Firma similar pero sobre estructura diferente
```

---

### 🔧 Arquitectura en `apidian-go`

#### **Servicios Genéricos:**

```go
// Servicio genérico para todos los documentos
type DocumentService struct {
    repo       *repository.DocumentRepository
    soapClient *soap.DIANClient
    storage    *config.StorageConfig
    xmlBuilder DocumentBuilderFactory
}

// Factory para crear builders según tipo
type DocumentBuilderFactory interface {
    CreateBuilder(docType string, data interface{}) (ubl.DocumentBuilder, error)
}

// Métodos genéricos
func (s *DocumentService) Create(docType string, data interface{}) (*Document, error)
func (s *DocumentService) Sign(documentID int64, certPath, password string) error
func (s *DocumentService) SendToDIAN(documentID int64) error
func (s *DocumentService) GetStatus(trackID string) (*DIANStatus, error)
```

#### **Handlers Específicos:**

```go
// Cada tipo de documento tiene su handler para validaciones específicas
type InvoiceHandler struct {
    documentService *service.DocumentService
}

type CreditNoteHandler struct {
    documentService *service.DocumentService
}

// Pero comparten la lógica de firma, envío, consulta
```

---

## 🚀 Plan de Acción Recomendado

### ✅ **Fase 1: Factura Básica (COMPLETADA)**
- [x] Estructura UBL 2.1 base
- [x] Firma electrónica
- [x] Envío a DIAN (SendBillSync, SendBillAsync, SendTestSetAsync)
- [x] Generación de PDF
- [x] Generación de ZIP
- [x] **GetStatus implementado** (29/01/2026)
- [x] Campo `track_id` en BD
- [x] Actualización automática de estado DIAN

### 🔄 **Fase 2: Validación Completa (SIGUIENTE)**
1. ⏳ **Descuentos y Retenciones**
   - Agregar campos a BD (`document_allowance_charges`, `document_withholding_taxes`)
   - Actualizar builder XML
   - Validar cálculos
   - Actualizar totales

2. ⏳ **Mejoras de validación**
   - Validar contra XSD
   - Mensajes de error más claros
   - Validación de rangos de numeración

3. ⏳ **Agregar retenciones básicas**
   - Tabla `document_withholding_taxes`
   - Soporte en XML builder
   - Cálculo automático de retenciones

4. ⏳ **Mejorar validaciones XML**
   - Validar contra XSD de DIAN
   - Validar reglas de negocio
   - Mensajes de error claros

---

### **Fase 2: Notas Crédito/Débito (2-3 semanas)**
**Objetivo:** Soportar ajustes a facturas

1. ⏳ **Extender modelo `documents`**
   - Agregar `type_document_id`
   - Migrar datos existentes
   - Tabla `document_references`

2. ⏳ **Crear builders en `ubl21-dian`**
   - `CreditNoteBuilder` (hereda 90% de Invoice)
   - `DebitNoteBuilder` (hereda 90% de Invoice)
   - Reutilizar firma y canonicalización

3. ⏳ **Implementar endpoints**
   - `POST /api/v1/credit-notes`
   - `POST /api/v1/debit-notes`
   - `POST /api/v1/credit-notes/:id/sign`
   - `POST /api/v1/credit-notes/:id/send`

4. ⏳ **Validaciones específicas**
   - Nota debe referenciar factura válida
   - Totales no pueden exceder factura original (crédito)
   - Razón de corrección obligatoria

---

### **Fase 3: Eventos y ApplicationResponse (1-2 semanas)**
**Objetivo:** Soportar ciclo completo de eventos

1. ⏳ **Implementar `SendEventUpdateStatus`**
   - Acuse de recibo (030)
   - Aceptación expresa (032)
   - Rechazo (031)
   - Reclamo (034)

2. ⏳ **Crear tabla `document_events`**
   ```sql
   CREATE TABLE document_events (
     id BIGSERIAL PRIMARY KEY,
     document_id BIGINT REFERENCES documents(id),
     event_type_code VARCHAR(10), -- 030, 031, 032, 033, 034
     event_date TIMESTAMP,
     notes TEXT,
     xml_path TEXT,
     status VARCHAR(20),
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

3. ⏳ **Endpoints para eventos**
   - `POST /api/v1/documents/:id/events/acknowledge`
   - `POST /api/v1/documents/:id/events/accept`
   - `POST /api/v1/documents/:id/events/reject`
   - `POST /api/v1/documents/:id/events/claim`

---

### **Fase 4: Documentos Avanzados (3-4 semanas)**
**Objetivo:** Soportar otros tipos de documentos

1. ⏳ **Nómina electrónica**
   - Estructura muy diferente
   - Campos específicos de empleado
   - `SendNominaSync`

2. ⏳ **Documentos soporte**
   - Para no obligados a facturar
   - Similar a factura pero con diferencias

3. ⏳ **Documentos equivalentes**
   - Tiquetes de máquina registradora
   - Facturas de servicios públicos

---

### **Fase 5: Métodos SOAP Adicionales (1-2 semanas)**
**Objetivo:** Completar integración con DIAN

1. ⏳ **Consultas de información**
   - `GetNumberingRange`
   - `GetXmlByDocumentKey`
   - `GetDocumentInfo`

2. ⏳ **Set de pruebas**
   - `SendTestSetAsync`
   - Proceso de certificación

3. ⏳ **Adjuntos**
   - `SendBillAttachmentAsync`

---

## 📊 Resumen de Prioridades

### 🔥 **CRÍTICO (Completado)**
1. ✅ `GetStatus` - **IMPLEMENTADO** (29/01/2026)
   - Endpoint: `POST /api/v1/invoices/:id/status`
   - Campo `track_id` en BD
   - Actualización automática de `dian_status`

### 🟡 **ALTA (Hacer YA)**
2. ⏳ Descuentos/Retenciones - Muy comunes en facturas reales

### 🟡 **ALTA (Próximas 2-4 semanas)**
3. ⏳ CreditNote/DebitNote - Necesarios para ajustes
4. ⏳ Eventos (SendEventUpdateStatus) - Ciclo completo

### 🟢 **MEDIA (1-2 meses)**
5. ⏳ Consultas adicionales (GetNumberingRange, GetXmlByDocumentKey)
6. ⏳ SendTestSetAsync (certificación)

### 🔵 **BAJA (Futuro)**
7. ⏳ Nómina electrónica
8. ⏳ Documentos equivalentes
9. ⏳ Adjuntos

---

## 💡 Respuestas Directas

### **1. ¿Agregar campos nuevos a Invoice?**
**SÍ**, pero de forma incremental:
- Primero: Descuentos y retenciones (más comunes)
- Después: Anticipos, referencias
- Último: Campos avanzados

### **2. ¿Nuevos campos en BD?**
**SÍ**, necesitas:
- `document_allowance_charges` (descuentos/cargos)
- `document_withholding_taxes` (retenciones)
- `document_references` (referencias entre documentos)
- `document_prepaid_payments` (anticipos)
- `document_events` (eventos)

### **3. ¿Formatear XML base?**
**SÍ**, `ubl21-dian` ya tiene la estructura base, solo falta:
- Agregar elementos opcionales (AllowanceCharge, WithholdingTaxTotal, etc.)
- Validar contra XSD de DIAN
- Mejorar mensajes de error

### **4. ¿Modelos generales para todos los documentos?**
**SÍ**, usa tabla `documents` polimórfica:
- `type_document_id` para diferenciar tipos
- Campos comunes para todos
- Tablas auxiliares para datos específicos

### **5. ¿Con Invoice se pueden hacer los demás?**
- **CreditNote/DebitNote: SÍ** (90% reutilizable)
- **ApplicationResponse: PARCIAL** (30% reutilizable)
- **Nómina: NO** (20% reutilizable, estructura muy diferente)

### **6. ¿Preparar los 15 métodos SOAP?**
**SÍ**, crea interfaces genéricas:
```go
// Abstracción genérica
SendDocument(docType string, xml []byte) (trackID string, error)
GetDocumentStatus(trackID string) (*DIANStatus, error)
SendEvent(eventType string, documentKey string, xml []byte) (trackID string, error)
```

---

## 📚 Referencias

- [Resolución 000042 de 2020 - DIAN](https://www.dian.gov.co)
- [UBL 2.1 Specification](http://docs.oasis-open.org/ubl/UBL-2.1.html)
- [Anexo Técnico Factura Electrónica](https://www.dian.gov.co/impuestos/factura-electronica)

---

**Fecha de creación:** 2026-01-29  
**Última actualización:** 2026-01-29  
**Versión:** 1.0
