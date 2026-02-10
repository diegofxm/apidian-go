# 🔍 ANÁLISIS DE LÓGICA DISPERSA - APIDIAN-GO

## 📋 DIAGNÓSTICO COMPLETO

**Fecha**: 2026-02-10  
**Estado actual**: Arquitectura en capas con lógica desorganizada  
**Objetivo**: Centralizar y organizar la lógica de negocio

---

## 🚨 PROBLEMAS IDENTIFICADOS

### **1. LÓGICA DE NEGOCIO DISPERSA EN MÚLTIPLES CAPAS**

#### **Problema Principal:**
La lógica de negocio está **regada** en 3 capas diferentes:

```
❌ ACTUAL (Lógica dispersa):

1. Handler (internal/handler/invoice_handler.go)
   - Validaciones de negocio (líneas 77-85)
   - Construcción de respuestas con lógica (líneas 284-314)
   - Formateo de datos (línea 26-28: formatMoney)
   - Manejo de errores de negocio (múltiples if/else)

2. Service (internal/service/invoice/invoice_service.go)
   - Validaciones de negocio (líneas 52-84)
   - Cálculos de totales (líneas 122-178)
   - Lógica de estados (líneas 269-272, 303-305, 319-321, 426-428)
   - Generación de números de factura (línea 181)
   - Determinación automática de payment_form (líneas 107-119)

3. Helpers (internal/service/invoice/invoice_helpers.go)
   - Utilidades de conversión (líneas 372-463)
   - Lógica de mapeo de códigos DIAN (líneas 394-451)
   - Formateo de datos (línea 453-463)

4. Template Builder (internal/service/invoice/template_builder.go)
   - Cálculos de impuestos (líneas 29-38)
   - Generación de CUFE (líneas 45-58)
   - Generación de QR (líneas 67-78)
   - Lógica de formateo de fechas (líneas 17-26)

5. Domain (internal/domain/invoice.go)
   - SOLO structs, SIN métodos de negocio
   - NO tiene CalculateTotals()
   - NO tiene Validate()
   - NO tiene CanBeSigned(), CanBeSent()
```

---

### **2. VIOLACIONES DE PRINCIPIOS SOLID**

#### **Single Responsibility Principle (SRP)**
```go
// ❌ InvoiceService hace DEMASIADO:
type InvoiceService struct {
    // 1. Persistencia (repositories)
    invoiceRepo, companyRepo, customerRepo, resolutionRepo, productRepo, certificateRepo
    
    // 2. Configuración
    storage, keepUnsignedXML
}

// Métodos del servicio:
- Create()              // Validación + cálculos + persistencia
- Sign()                // Validación + XML + firma + persistencia + archivos
- SendToDIAN()          // Validación + ZIP + SOAP + persistencia
- GeneratePDF()         // Validación + PDF + persistencia
- GenerateAttachedDocument() // Validación + XML + firma + ZIP + persistencia
```

#### **Open/Closed Principle (OCP)**
```go
// ❌ Código cerrado a extensión:
// Para agregar CreditNote, DebitNote, etc., hay que duplicar TODO el código
// No hay abstracción común para "Document"
```

#### **Dependency Inversion Principle (DIP)**
```go
// ❌ Service depende de implementaciones concretas:
invoiceRepo := repository.NewInvoiceRepository(db)  // Implementación concreta
companyRepo := repository.NewCompanyRepository(db)  // Implementación concreta

// ✅ Debería depender de interfaces:
invoiceRepo InvoiceRepository  // Interfaz
companyRepo CompanyRepository  // Interfaz
```

---

### **3. LÓGICA DE NEGOCIO EN LUGARES INCORRECTOS**

#### **A) Handler con lógica de negocio**
```go
// ❌ internal/handler/invoice_handler.go:284-314
func (h *InvoiceHandler) Sign(c *fiber.Ctx) error {
    // ...
    
    // LÓGICA DE NEGOCIO EN EL HANDLER (debería estar en domain)
    data := &domain.DocumentData{
        InvoiceID:     invoice.ID,
        Number:        invoice.Number,
        URLInvoiceXML: "FES-" + invoice.Number + ".xml",
        URLInvoicePDF: "FES-" + invoice.Number + ".pdf",
    }
    
    if invoice.UUID != nil && *invoice.UUID != "" {
        data.CUFE = *invoice.UUID
        
        // CONSTRUCCIÓN DE QR STRING (lógica de negocio)
        qrStr := "NumFac: " + invoice.Number + "\n"
        qrStr += "FecFac: " + invoice.IssueDate.Format("2006-01-02") + "\n"
        qrStr += "NitFac: " + invoice.Company.NIT + "\n"
        // ... más líneas
        data.QRStr = qrStr
    }
    
    resp := domain.NewSuccessResponse("Factura #"+invoice.Number+" firmada con éxito", data)
    return c.Status(fiber.StatusOK).JSON(resp)
}
```

#### **B) Service con cálculos que deberían estar en Entity**
```go
// ❌ internal/service/invoice/invoice_service.go:122-178
func (s *InvoiceService) Create(req *domain.CreateInvoiceRequest, userID int64) (*domain.Invoice, error) {
    // ...
    
    // CÁLCULOS DE TOTALES (debería estar en invoice.CalculateTotals())
    var subtotal, taxTotal float64
    var lines []domain.InvoiceLine
    
    for i, lineReq := range req.Lines {
        lineTotal := lineReq.Quantity * unitPrice
        taxAmount := lineTotal * (taxRate / 100)
        
        subtotal += lineTotal
        taxTotal += taxAmount
        
        line := domain.InvoiceLine{
            ProductID:   lineReq.ProductID,
            Description: description,
            Quantity:    lineReq.Quantity,
            UnitPrice:   unitPrice,
            LineTotal:   lineTotal,
            TaxRate:     taxRate,
            TaxAmount:   taxAmount,
        }
        lines = append(lines, line)
    }
    
    total := subtotal + taxTotal
    
    // ...
}
```

#### **C) Template Builder con lógica de negocio**
```go
// ❌ internal/service/invoice/template_builder.go:29-58
func (s *InvoiceService) BuildInvoiceWithTemplates(inv *domain.Invoice) ([]byte, string, error) {
    // ...
    
    // CÁLCULO DE IMPUESTOS POR TIPO (debería estar en invoice.GetTaxByType())
    var ivaAmount, incAmount, icaAmount float64
    for _, line := range inv.Lines {
        if line.TaxTypeCode == "01" {
            ivaAmount += line.TaxAmount
        } else if line.TaxTypeCode == "04" {
            incAmount += line.TaxAmount
        } else if line.TaxTypeCode == "03" {
            icaAmount += line.TaxAmount
        }
    }
    
    // CÁLCULO DE CUFE (debería estar en invoice.CalculateCUFE())
    cufe := signature.CalculateCUFE(
        inv.Number,
        inv.IssueDate,
        issueTime,
        inv.Subtotal,
        ivaAmount,
        incAmount,
        icaAmount,
        inv.Total,
        inv.Company.NIT,
        inv.Customer.IdentificationNumber,
        technicalKey,
        getEnvironmentStr(inv.Software),
    )
    
    // ...
}
```

---

### **4. DOMAIN SIN COMPORTAMIENTO**

```go
// ❌ internal/domain/invoice.go
type Invoice struct {
    ID         int64
    CompanyID  int64
    Number     string
    Subtotal   float64
    TaxTotal   float64
    Total      float64
    Status     string
    // ... 40+ campos más
}

// NO HAY MÉTODOS:
// - invoice.CalculateTotals()
// - invoice.Validate()
// - invoice.CanBeSigned()
// - invoice.CanBeSent()
// - invoice.MarkAsSigned()
// - invoice.MarkAsSent()
// - invoice.GetTaxByType()
// - invoice.GenerateQRString()
```

**Resultado**: El dominio es un "modelo anémico" (anemic domain model), solo datos sin comportamiento.

---

### **5. DUPLICACIÓN DE CÓDIGO**

#### **Validaciones repetidas**
```go
// ❌ Repetido en CADA método del handler:
userID, err := utils.GetUserID(c)
if err != nil {
    return response.Unauthorized(c, "User not authenticated")
}

// ❌ Repetido en CADA método del handler:
if err.Error() == "invoice not found" {
    return response.NotFound(c, "Invoice not found")
}
if err.Error() == "unauthorized access to invoice" {
    return response.Unauthorized(c, "Unauthorized access to invoice")
}
```

#### **Lógica de conversión repetida**
```go
// ❌ Repetido en múltiples lugares:
idStr := c.Params("id")
id, err := strconv.ParseInt(idStr, 10, 64)
if err != nil {
    return response.BadRequest(c, "Invalid ID")
}
```

---

## ✅ SOLUCIÓN: CENTRALIZAR LÓGICA EN EL DOMINIO

### **ARQUITECTURA OBJETIVO (Hexagonal Pura)**

```
┌─────────────────────────────────────────────────────────────┐
│                    CAPA DE PRESENTACIÓN                     │
│  internal/adapters/input/http/handlers/                     │
│  - invoice_handler.go                                       │
│  - Responsabilidad: HTTP request/response SOLAMENTE         │
│  - NO lógica de negocio                                     │
│  - NO validaciones de negocio                               │
│  - NO cálculos                                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    CAPA DE APLICACIÓN                       │
│  internal/application/invoice/                              │
│  - create_invoice.go (Use Case)                             │
│  - sign_invoice.go (Use Case)                               │
│  - send_invoice.go (Use Case)                               │
│  - Responsabilidad: Orquestar el flujo                      │
│  - Usa: Repositories (interfaces)                           │
│  - Usa: Services (interfaces)                               │
│  - Delega lógica de negocio a entities                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    CAPA DE DOMINIO (NÚCLEO)                 │
│  internal/domain/entities/                                  │
│  - invoice.go (Entity con MÉTODOS)                          │
│  - Responsabilidad: TODA la lógica de negocio               │
│                                                              │
│  ✅ Métodos de negocio:                                     │
│  - invoice.CalculateTotals()                                │
│  - invoice.Validate()                                       │
│  - invoice.CanBeSigned() bool                               │
│  - invoice.CanBeSent() bool                                 │
│  - invoice.MarkAsSigned(cufe string)                        │
│  - invoice.MarkAsSent(trackID string)                       │
│  - invoice.GetTaxByType(taxType string) float64             │
│  - invoice.GenerateQRString() string                        │
│  - invoice.GetTotalTax() float64                            │
│                                                              │
│  internal/domain/valueobjects/                              │
│  - money.go (Amount + Currency)                             │
│  - tax.go (TaxType + Rate + Amount)                         │
│  - nit.go (NIT + DV con validación)                         │
│  - cufe.go (CUFE con cálculo)                               │
│                                                              │
│  internal/domain/ports/                                     │
│  - input/invoice_usecase.go (Interfaz)                      │
│  - output/invoice_repository.go (Interfaz)                  │
│  - output/dian_service.go (Interfaz)                        │
└─────────────────────────────────────────────────────────────┘
                            ↑
┌─────────────────────────────────────────────────────────────┐
│                    CAPA DE INFRAESTRUCTURA                  │
│  internal/adapters/output/                                  │
│  - postgres/invoice_repository.go (Implementación)          │
│  - dian/dian_adapter.go (Implementación)                    │
│  - Responsabilidad: Detalles técnicos                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 PLAN DE CENTRALIZACIÓN DE LÓGICA

### **FASE 1: MOVER LÓGICA AL DOMINIO**

#### **Paso 1.1: Crear Entity con métodos**

```go
// ✅ internal/domain/entities/invoice.go
package entities

import (
    "fmt"
    "time"
    "apidian-go/internal/domain/valueobjects"
)

type Invoice struct {
    ID              uint
    CompanyID       uint
    CustomerID      uint
    Number          string
    IssueDate       time.Time
    Lines           []InvoiceLine
    TaxTotals       []valueobjects.Tax
    LegalMonetaryTotal valueobjects.Money
    Status          InvoiceStatus
    TrackID         string
    CUFE            string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// ==================== MÉTODOS DE NEGOCIO ====================

// CalculateTotals calcula subtotal, impuestos y total
func (i *Invoice) CalculateTotals() {
    var lineTotal float64
    taxMap := make(map[string]*valueobjects.Tax)
    
    for _, line := range i.Lines {
        lineTotal += line.LineExtensionAmount.Amount
        
        if line.Tax != nil {
            key := line.Tax.TaxType
            if taxMap[key] == nil {
                taxMap[key] = &valueobjects.Tax{
                    TaxType: line.Tax.TaxType,
                    Rate:    line.Tax.Rate,
                    Amount:  0,
                }
            }
            taxMap[key].Amount += line.Tax.Amount
        }
    }
    
    i.TaxTotals = make([]valueobjects.Tax, 0, len(taxMap))
    for _, tax := range taxMap {
        i.TaxTotals = append(i.TaxTotals, *tax)
    }
    
    i.LegalMonetaryTotal = valueobjects.Money{
        Amount:   lineTotal + i.GetTotalTax(),
        Currency: "COP",
    }
}

// GetTotalTax retorna la suma de todos los impuestos
func (i *Invoice) GetTotalTax() float64 {
    var total float64
    for _, tax := range i.TaxTotals {
        total += tax.Amount
    }
    return total
}

// GetTaxByType retorna el monto de un tipo de impuesto específico
func (i *Invoice) GetTaxByType(taxType string) float64 {
    for _, tax := range i.TaxTotals {
        if tax.TaxType == taxType {
            return tax.Amount
        }
    }
    return 0
}

// Validate valida que la factura tenga todos los datos requeridos
func (i *Invoice) Validate() error {
    if i.Number == "" {
        return ErrInvoiceNumberRequired
    }
    if len(i.Lines) == 0 {
        return ErrInvoiceLinesRequired
    }
    if i.CompanyID == 0 {
        return ErrCompanyRequired
    }
    if i.CustomerID == 0 {
        return ErrCustomerRequired
    }
    return nil
}

// CanBeSigned verifica si la factura puede ser firmada
func (i *Invoice) CanBeSigned() bool {
    return i.Status == InvoiceStatusDraft
}

// CanBeSent verifica si la factura puede ser enviada a DIAN
func (i *Invoice) CanBeSent() bool {
    return i.Status == InvoiceStatusSigned
}

// CanBeUpdated verifica si la factura puede ser actualizada
func (i *Invoice) CanBeUpdated() bool {
    return i.Status == InvoiceStatusDraft
}

// CanBeDeleted verifica si la factura puede ser eliminada
func (i *Invoice) CanBeDeleted() bool {
    return i.Status == InvoiceStatusDraft
}

// MarkAsSigned marca la factura como firmada
func (i *Invoice) MarkAsSigned(cufe string) {
    i.Status = InvoiceStatusSigned
    i.CUFE = cufe
    i.UpdatedAt = time.Now()
}

// MarkAsSent marca la factura como enviada a DIAN
func (i *Invoice) MarkAsSent(trackID string) {
    i.Status = InvoiceStatusSent
    i.TrackID = trackID
    i.UpdatedAt = time.Now()
}

// MarkAsAccepted marca la factura como aceptada por DIAN
func (i *Invoice) MarkAsAccepted() {
    i.Status = InvoiceStatusAccepted
    i.UpdatedAt = time.Now()
}

// MarkAsRejected marca la factura como rechazada por DIAN
func (i *Invoice) MarkAsRejected() {
    i.Status = InvoiceStatusRejected
    i.UpdatedAt = time.Now()
}

// GenerateQRString genera el string del código QR
func (i *Invoice) GenerateQRString(company Company, customer Customer) string {
    qr := fmt.Sprintf("NumFac: %s\n", i.Number)
    qr += fmt.Sprintf("FecFac: %s\n", i.IssueDate.Format("2006-01-02"))
    qr += fmt.Sprintf("NitFac: %s\n", company.NIT)
    qr += fmt.Sprintf("DocAdq: %s\n", customer.IdentificationNumber)
    qr += fmt.Sprintf("ValFac: %.2f\n", i.LegalMonetaryTotal.Amount - i.GetTotalTax())
    qr += fmt.Sprintf("ValIva: %.2f\n", i.GetTaxByType("01"))
    qr += fmt.Sprintf("ValOtroIm: 0.00\n")
    qr += fmt.Sprintf("ValTotal: %.2f\n", i.LegalMonetaryTotal.Amount)
    qr += fmt.Sprintf("CUFE: %s\n", i.CUFE)
    qr += fmt.Sprintf("https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=%s", i.CUFE)
    return qr
}

// InvoiceStatus representa el estado de una factura
type InvoiceStatus string

const (
    InvoiceStatusDraft    InvoiceStatus = "draft"
    InvoiceStatusSigned   InvoiceStatus = "signed"
    InvoiceStatusSent     InvoiceStatus = "sent"
    InvoiceStatusAccepted InvoiceStatus = "accepted"
    InvoiceStatusRejected InvoiceStatus = "rejected"
)

// Domain Errors
var (
    ErrInvoiceNumberRequired = fmt.Errorf("invoice number is required")
    ErrInvoiceLinesRequired  = fmt.Errorf("invoice must have at least one line")
    ErrCompanyRequired       = fmt.Errorf("company is required")
    ErrCustomerRequired      = fmt.Errorf("customer is required")
)
```

---

#### **Paso 1.2: Crear Value Objects**

```go
// ✅ internal/domain/valueobjects/money.go
package valueobjects

type Money struct {
    Amount   float64
    Currency string
}

func NewMoney(amount float64, currency string) Money {
    return Money{
        Amount:   amount,
        Currency: currency,
    }
}

func (m Money) Add(other Money) Money {
    if m.Currency != other.Currency {
        panic("cannot add money with different currencies")
    }
    return Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }
}

func (m Money) Multiply(factor float64) Money {
    return Money{
        Amount:   m.Amount * factor,
        Currency: m.Currency,
    }
}
```

```go
// ✅ internal/domain/valueobjects/tax.go
package valueobjects

type Tax struct {
    TaxType string  // "01" = IVA, "04" = INC, "03" = ICA
    Rate    float64 // Porcentaje (ej: 19.0 para 19%)
    Amount  float64 // Monto calculado
}

func NewTax(taxType string, rate float64, baseAmount float64) Tax {
    return Tax{
        TaxType: taxType,
        Rate:    rate,
        Amount:  baseAmount * (rate / 100),
    }
}
```

```go
// ✅ internal/domain/valueobjects/nit.go
package valueobjects

import (
    "errors"
    "strconv"
)

type NIT struct {
    Number string
    DV     string
}

func NewNIT(number, dv string) (NIT, error) {
    nit := NIT{Number: number, DV: dv}
    if err := nit.Validate(); err != nil {
        return NIT{}, err
    }
    return nit, nil
}

func (n NIT) Validate() error {
    if n.Number == "" {
        return errors.New("NIT number is required")
    }
    
    calculatedDV := n.CalculateDV()
    if calculatedDV != n.DV {
        return errors.New("invalid DV")
    }
    
    return nil
}

func (n NIT) CalculateDV() string {
    primes := []int{71, 67, 59, 53, 47, 43, 41, 37, 29, 23, 19, 17, 13, 7, 3}
    sum := 0
    
    for i, digit := range n.Number {
        if i >= len(primes) {
            break
        }
        d, _ := strconv.Atoi(string(digit))
        sum += d * primes[len(primes)-len(n.Number)+i]
    }
    
    remainder := sum % 11
    if remainder == 0 || remainder == 1 {
        return strconv.Itoa(remainder)
    }
    return strconv.Itoa(11 - remainder)
}

func (n NIT) String() string {
    return n.Number + "-" + n.DV
}
```

---

### **FASE 2: REFACTORIZAR SERVICE (Casos de Uso)**

#### **Antes (Service con lógica)**
```go
// ❌ internal/service/invoice/invoice_service.go
func (s *InvoiceService) Create(req *domain.CreateInvoiceRequest, userID int64) (*domain.Invoice, error) {
    // Validaciones de negocio
    company, err := s.companyRepo.GetByID(req.CompanyID)
    if err != nil {
        return nil, fmt.Errorf("company not found")
    }
    if company.UserID != userID {
        return nil, fmt.Errorf("unauthorized access to company")
    }
    
    // Cálculos de totales (LÓGICA DE NEGOCIO)
    var subtotal, taxTotal float64
    for i, lineReq := range req.Lines {
        lineTotal := lineReq.Quantity * unitPrice
        taxAmount := lineTotal * (taxRate / 100)
        subtotal += lineTotal
        taxTotal += taxAmount
    }
    total := subtotal + taxTotal
    
    // Crear factura
    invoice := &domain.Invoice{
        CompanyID: req.CompanyID,
        Subtotal:  subtotal,
        TaxTotal:  taxTotal,
        Total:     total,
        Status:    "draft",
    }
    
    if err := s.invoiceRepo.Create(invoice, lines); err != nil {
        return nil, err
    }
    
    return invoice, nil
}
```

#### **Después (Use Case sin lógica)**
```go
// ✅ internal/application/invoice/create_invoice.go
package invoice

type CreateInvoiceUseCase struct {
    invoiceRepo    repositories.InvoiceRepository
    companyRepo    repositories.CompanyRepository
    customerRepo   repositories.CustomerRepository
    productRepo    repositories.ProductRepository
    resolutionRepo repositories.ResolutionRepository
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, req CreateInvoiceRequest) (*entities.Invoice, error) {
    // 1. Validar permisos (autorización)
    company, err := uc.companyRepo.GetByID(ctx, req.CompanyID)
    if err != nil {
        return nil, ErrCompanyNotFound
    }
    if company.UserID != req.UserID {
        return nil, ErrUnauthorized
    }
    
    // 2. Obtener datos necesarios
    customer, err := uc.customerRepo.GetByID(ctx, req.CustomerID)
    if err != nil {
        return nil, ErrCustomerNotFound
    }
    
    resolution, err := uc.resolutionRepo.GetByID(ctx, req.ResolutionID)
    if err != nil {
        return nil, ErrResolutionNotFound
    }
    
    // 3. Crear entity (sin lógica de negocio aquí)
    invoice := entities.NewInvoice(
        company.ID,
        customer.ID,
        resolution.ID,
        req.IssueDate,
    )
    
    // 4. Agregar líneas
    for _, lineReq := range req.Lines {
        product, err := uc.productRepo.GetByID(ctx, lineReq.ProductID)
        if err != nil {
            return nil, ErrProductNotFound
        }
        
        line := entities.NewInvoiceLine(
            product,
            lineReq.Quantity,
            lineReq.UnitPrice,
        )
        
        invoice.AddLine(line)
    }
    
    // 5. Calcular totales (DELEGADO A LA ENTITY)
    invoice.CalculateTotals()
    
    // 6. Validar (DELEGADO A LA ENTITY)
    if err := invoice.Validate(); err != nil {
        return nil, err
    }
    
    // 7. Persistir
    if err := uc.invoiceRepo.Create(ctx, invoice); err != nil {
        return nil, err
    }
    
    return invoice, nil
}
```

---

### **FASE 3: SIMPLIFICAR HANDLER**

#### **Antes (Handler con lógica)**
```go
// ❌ internal/handler/invoice_handler.go
func (h *InvoiceHandler) Sign(c *fiber.Ctx) error {
    userID, err := utils.GetUserID(c)
    if err != nil {
        return response.Unauthorized(c, "User not authenticated")
    }
    
    idStr := c.Params("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return response.BadRequest(c, "Invalid ID")
    }
    
    if err := h.service.Sign(id, userID); err != nil {
        if err.Error() == "invoice not found" {
            return response.NotFound(c, "Invoice not found")
        }
        // ... más ifs
        return response.InternalServerError(c, err.Error())
    }
    
    invoice, err := h.service.GetByID(id, userID)
    if err != nil {
        return response.InternalServerError(c, "Failed to retrieve signed invoice")
    }
    
    // LÓGICA DE NEGOCIO EN EL HANDLER
    data := &domain.DocumentData{
        InvoiceID:     invoice.ID,
        Number:        invoice.Number,
        URLInvoiceXML: "FES-" + invoice.Number + ".xml",
        URLInvoicePDF: "FES-" + invoice.Number + ".pdf",
    }
    
    if invoice.UUID != nil && *invoice.UUID != "" {
        data.CUFE = *invoice.UUID
        qrStr := "NumFac: " + invoice.Number + "\n"
        // ... más construcción de QR
        data.QRStr = qrStr
    }
    
    resp := domain.NewSuccessResponse("Factura firmada", data)
    return c.Status(fiber.StatusOK).JSON(resp)
}
```

#### **Después (Handler sin lógica)**
```go
// ✅ internal/adapters/input/http/handlers/invoice_handler.go
package handlers

type InvoiceHandler struct {
    invoiceUseCase input.InvoiceUseCase
}

func (h *InvoiceHandler) Sign(c *fiber.Ctx) error {
    // 1. Extraer parámetros
    userID := c.Locals("user_id").(uint)
    id, err := strconv.ParseUint(c.Params("id"), 10, 64)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
    }
    
    // 2. Ejecutar caso de uso
    invoice, err := h.invoiceUseCase.Sign(c.Context(), uint(id), userID)
    if err != nil {
        return handleError(c, err)
    }
    
    // 3. Mapear a DTO (sin lógica de negocio)
    response := mappers.ToInvoiceResponse(invoice)
    
    return c.Status(200).JSON(response)
}

// handleError centraliza el manejo de errores
func handleError(c *fiber.Ctx, err error) error {
    switch {
    case errors.Is(err, ErrNotFound):
        return c.Status(404).JSON(fiber.Map{"error": err.Error()})
    case errors.Is(err, ErrUnauthorized):
        return c.Status(401).JSON(fiber.Map{"error": err.Error()})
    case errors.Is(err, ErrValidation):
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    default:
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
}
```

---

## 📊 COMPARACIÓN: ANTES vs DESPUÉS

| Aspecto | ❌ Antes (Actual) | ✅ Después (Hexagonal) |
|---------|-------------------|------------------------|
| **Lógica de negocio** | Dispersa en 4 capas | Centralizada en domain |
| **Entity Invoice** | Solo datos (anemic) | Datos + 12 métodos |
| **Service** | 548 líneas | 150 líneas (orquestación) |
| **Handler** | 555 líneas con lógica | 200 líneas sin lógica |
| **Validaciones** | Repetidas en cada método | Centralizadas en entity |
| **Cálculos** | En service y builder | En entity |
| **Estados** | Strings mágicos | Enum tipado |
| **Testabilidad** | Difícil (acoplado) | Fácil (mocks) |
| **Reutilización** | Imposible | Alta |
| **Mantenibilidad** | Baja | Alta |

---

## 🎯 BENEFICIOS DE LA CENTRALIZACIÓN

### **1. Lógica de negocio en un solo lugar**
```go
// ✅ TODO en la entity:
invoice.CalculateTotals()
invoice.Validate()
invoice.CanBeSigned()
invoice.MarkAsSigned(cufe)
invoice.GenerateQRString(company, customer)
```

### **2. Reutilización**
```go
// ✅ Mismo código para Invoice, CreditNote, DebitNote:
type Document interface {
    CalculateTotals()
    Validate() error
    CanBeSigned() bool
    CanBeSent() bool
}
```

### **3. Testabilidad**
```go
// ✅ Test unitario simple:
func TestInvoiceCalculateTotals(t *testing.T) {
    invoice := &Invoice{
        Lines: []InvoiceLine{
            {Quantity: 2, UnitPrice: 100, TaxRate: 19},
        },
    }
    
    invoice.CalculateTotals()
    
    assert.Equal(t, 200.0, invoice.Subtotal)
    assert.Equal(t, 38.0, invoice.TaxTotal)
    assert.Equal(t, 238.0, invoice.Total)
}
```

### **4. Mantenibilidad**
```go
// ✅ Cambiar lógica de cálculo en UN solo lugar:
func (i *Invoice) CalculateTotals() {
    // Cambio aquí afecta a TODO el sistema
}
```

---

## 📋 CHECKLIST DE MIGRACIÓN

### **Fase 1: Domain (Semana 1)**
- [ ] Crear `internal/domain/entities/invoice.go` con métodos
- [ ] Crear `internal/domain/valueobjects/money.go`
- [ ] Crear `internal/domain/valueobjects/tax.go`
- [ ] Crear `internal/domain/valueobjects/nit.go`
- [ ] Crear `internal/domain/valueobjects/cufe.go`
- [ ] Definir interfaces en `internal/domain/ports/`

### **Fase 2: Application (Semana 2)**
- [ ] Crear `internal/application/invoice/create_invoice.go`
- [ ] Crear `internal/application/invoice/sign_invoice.go`
- [ ] Crear `internal/application/invoice/send_invoice.go`
- [ ] Mover lógica de service a use cases
- [ ] Eliminar lógica de negocio de service

### **Fase 3: Adapters (Semana 3)**
- [ ] Refactorizar handlers (eliminar lógica)
- [ ] Crear mappers (Entity ↔ DTO)
- [ ] Implementar repositories (interfaces)
- [ ] Implementar DIAN adapter

### **Fase 4: Testing (Semana 4)**
- [ ] Tests unitarios de entities
- [ ] Tests de use cases con mocks
- [ ] Tests de integración
- [ ] Tests E2E

---

## 🎓 CONCLUSIÓN

### **Problema Principal:**
La lógica de negocio está **dispersa** en múltiples capas (Handler, Service, Helpers, Builder), violando principios SOLID y dificultando el mantenimiento.

### **Solución:**
**Centralizar TODA la lógica de negocio en el dominio** (entities + value objects), dejando:
- **Handlers**: Solo HTTP request/response
- **Use Cases**: Solo orquestación
- **Repositories**: Solo persistencia
- **Adapters**: Solo detalles técnicos

### **Resultado:**
- ✅ Código más limpio y organizado
- ✅ Fácil de testear
- ✅ Fácil de mantener
- ✅ Fácil de extender (agregar CreditNote, DebitNote, etc.)
- ✅ Lógica de negocio reutilizable

---

**Fecha de creación**: 2026-02-10  
**Versión**: 1.0  
**Próximo paso**: Implementar Fase 1 (Domain con métodos)
