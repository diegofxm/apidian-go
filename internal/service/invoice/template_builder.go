package invoice

import (
	"apidian-go/internal/domain"
	"fmt"

	"github.com/diegofxm/ubl21-dian/documents/invoice"
	"github.com/diegofxm/ubl21-dian/signature"
)

// BuildInvoiceWithTemplates genera una factura usando el nuevo sistema de templates
func (s *InvoiceService) BuildInvoiceWithTemplates(inv *domain.Invoice) ([]byte, string, error) {
	// 1. Crear builder
	builder := invoice.NewBuilder()

	// 2. Formatear fechas
	issueDate := inv.IssueDate.Format("2006-01-02")
	// Forzar timezone de Colombia (-05:00)
	issueTime := fmt.Sprintf("%02d:%02d:%02d-05:00", 
		inv.IssueTime.Hour(), 
		inv.IssueTime.Minute(), 
		inv.IssueTime.Second())
	dueDate := inv.IssueDate.AddDate(0, 0, 7).Format("2006-01-02")
	if inv.DueDate != nil {
		dueDate = inv.DueDate.Format("2006-01-02")
	}

	// 3. Calcular CUFE
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

	technicalKey := ""
	if inv.Resolution.TechnicalKey != nil {
		technicalKey = *inv.Resolution.TechnicalKey
	}

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

	// 4. Calcular Security Code
	securityCode := signature.CalculateSoftwareSecurityCode(
		inv.Software.Identifier,
		inv.Software.PIN,
		inv.Number,
	)

	// 5. Generar QR Code
	qrCode := signature.GenerateQRCode(
		inv.Number,
		inv.IssueDate,
		inv.Company.NIT,
		inv.Customer.IdentificationNumber,
		inv.Subtotal,
		ivaAmount,
		inv.Total,
		cufe,
		getEnvironmentStr(inv.Software),
	)

	// 6. Configurar datos básicos
	builder.SetInvoiceData(inv.Number, cufe, issueDate, issueTime, dueDate).
		SetProfileExecutionID(getEnvironmentStr(inv.Software)).
		SetNote(getInvoiceNote(inv)).
		SetDianExtensions(
			inv.Resolution.Resolution,
			inv.Resolution.DateFrom.Format("2006-01-02"),
			inv.Resolution.DateTo.Format("2006-01-02"),
			inv.Resolution.Prefix,
			fmt.Sprintf("%d", inv.Resolution.FromNumber),
			fmt.Sprintf("%d", inv.Resolution.ToNumber),
			inv.Company.NIT,
			getProviderSchemeID(inv.Company.TypeOrganizationCode, inv.Company.DV),
			getProviderSchemeName(inv.Company.TypeOrganizationCode),
			inv.Software.Identifier,
			securityCode,
			qrCode,
		)

	// 7. Configurar Supplier
	supplier := invoice.PartyTemplateData{
		AdditionalAccountID:        inv.Company.TypeOrganizationCode,
		PartyName:                  inv.Company.Name,
		IndustryClassificationCode: formatIndustryCodes(getStringValue(inv.Company.IndustryCodes)),
		Address: invoice.AddressTemplateData{
			ID:                   inv.Company.MunicipalityCode,
			CityName:             inv.Company.Municipality,
			PostalZone:           getStringValue(inv.Company.PostalZone),
			CountrySubentity:     inv.Company.Department,
			CountrySubentityCode: inv.Company.DepartmentCode,
			Line:                 inv.Company.AddressLine,
			CountryCode:          inv.Company.CountryCode,
			CountryName:          inv.Company.CountryName,
		},
		TaxScheme: invoice.TaxSchemeTemplateData{
			RegistrationName:    inv.Company.RegistrationName,
			CompanyID:           inv.Company.NIT,
			CompanyIDSchemeID:   getDocumentTypeSchemeID(inv.Company.DocumentTypeCode),
			CompanyIDSchemeName: getDocumentTypeSchemeName(inv.Company.DocumentTypeCode),
			TaxLevelCode:        inv.Company.TaxLevelCode,
			ID:                  inv.Company.TaxSchemeID,
			Name:                inv.Company.TaxSchemeName,
		},
		LegalEntity: invoice.LegalEntityTemplateData{
			RegistrationName:            inv.Company.RegistrationName,
			CompanyID:                   inv.Company.NIT,
			CompanyIDSchemeID:           getDocumentTypeSchemeID(inv.Company.DocumentTypeCode),
			CompanyIDSchemeName:         getDocumentTypeSchemeName(inv.Company.DocumentTypeCode),
			CorporateRegistrationScheme: inv.Resolution.Prefix,
		},
		Contact: invoice.ContactTemplateData{
			Telephone: getStringValue(inv.Company.Phone),
			Email:     getStringValue(inv.Company.Email),
		},
	}
	builder.SetSupplier(supplier)

	// 8. Configurar Customer
	customer := invoice.PartyTemplateData{
		AdditionalAccountID: inv.Customer.TypeOrganizationCode,
		PartyName:           inv.Customer.Name,
		Address: invoice.AddressTemplateData{
			ID:                   inv.Customer.MunicipalityCode,
			CityName:             inv.Customer.Municipality,
			PostalZone:           getStringValue(inv.Customer.PostalZone),
			CountrySubentity:     inv.Customer.Department,
			CountrySubentityCode: inv.Customer.DepartmentCode,
			Line:                 inv.Customer.AddressLine,
			CountryCode:          inv.Customer.CountryCode,
			CountryName:          inv.Customer.CountryName,
		},
		TaxScheme: invoice.TaxSchemeTemplateData{
			RegistrationName:    inv.Customer.Name,
			CompanyID:           inv.Customer.IdentificationNumber,
			CompanyIDSchemeID:   getDocumentTypeSchemeID(inv.Customer.DocumentTypeCode),
			CompanyIDSchemeName: getDocumentTypeSchemeName(inv.Customer.DocumentTypeCode),
			TaxLevelCode:        inv.Customer.TaxLevelCode,
			ID:                  inv.Customer.TaxSchemeID,
			Name:                inv.Customer.TaxSchemeName,
		},
		LegalEntity: invoice.LegalEntityTemplateData{
			RegistrationName:    inv.Customer.Name,
			CompanyID:           inv.Customer.IdentificationNumber,
			CompanyIDSchemeID:   getDocumentTypeSchemeID(inv.Customer.DocumentTypeCode),
			CompanyIDSchemeName: getDocumentTypeSchemeName(inv.Customer.DocumentTypeCode),
		},
		Contact: invoice.ContactTemplateData{
			Telephone: getStringValue(inv.Customer.Phone),
			Email:     getStringValue(inv.Customer.Email),
		},
	}
	builder.SetCustomer(customer)

	// 8.5. Configurar Delivery
	delivery := &invoice.DeliveryTemplateData{
		ActualDeliveryDate: issueDate,
		Address: invoice.AddressTemplateData{
			ID:                   inv.Customer.MunicipalityCode,
			CityName:             inv.Customer.Municipality,
			PostalZone:           getStringValue(inv.Customer.PostalZone),
			CountrySubentity:     inv.Customer.Department,
			CountrySubentityCode: inv.Customer.DepartmentCode,
			Line:                 inv.Customer.AddressLine,
			CountryCode:          inv.Customer.CountryCode,
			CountryName:          inv.Customer.CountryName,
		},
	}
	builder.SetDelivery(delivery)

	// 9. Configurar Payment Means
	paymentMethodID := int64(0)
	if inv.PaymentMethodID != nil {
		paymentMethodID = int64(*inv.PaymentMethodID)
	}
	builder.SetPaymentMeans("1", getPaymentMethodCode(&paymentMethodID), dueDate)

	// 10. Configurar totales
	builder.SetMonetaryTotals(
		fmt.Sprintf("%.2f", inv.Subtotal),
		fmt.Sprintf("%.2f", inv.Subtotal),
		fmt.Sprintf("%.2f", inv.Total),
		"0.00",
		fmt.Sprintf("%.2f", inv.Total),
	)

	// 11. Agregar líneas
	for i, line := range inv.Lines {
		invoiceLine := invoice.InvoiceLineTemplateData{
			ID:                    fmt.Sprintf("%d", i+1),
			UnitCode:              line.UnitCode,
			Quantity:              fmt.Sprintf("%.6f", line.Quantity),
			LineExtensionAmount:   fmt.Sprintf("%.2f", line.LineTotal),
			FreeOfChargeIndicator: "false",
			CurrencyID:            "COP",
			Item: invoice.ItemTemplateData{
				Description: line.Description,
				StandardItemID: invoice.ItemIDTemplateData{
					ID:       line.ProductCode,
					SchemeID: "999",
				},
				AdditionalItemID: invoice.ItemIDTemplateData{
					ID:         line.ProductCode,
					SchemeID:   "999",
					SchemeName: "EAN13",
				},
			},
			Price: invoice.PriceTemplateData{
				Amount:       fmt.Sprintf("%.2f", line.UnitPrice),
				BaseQuantity: "1.000000",
			},
		}
		builder.AddInvoiceLine(invoiceLine)
	}

	// 12. Generar XML
	xmlBytes, err := builder.Build()
	if err != nil {
		return nil, "", fmt.Errorf("error building invoice XML: %w", err)
	}

	return xmlBytes, cufe, nil
}


// ValidateInvoiceForDIAN valida que una factura tenga todos los datos necesarios para DIAN
func ValidateInvoiceForDIAN(inv *domain.Invoice) error {
	if inv == nil {
		return fmt.Errorf("invoice cannot be nil")
	}

	if inv.Number == "" {
		return fmt.Errorf("invoice number is required")
	}
	if inv.InvoiceTypeCode == "" {
		return fmt.Errorf("invoice type code is required")
	}
	if inv.CurrencyCode == "" {
		return fmt.Errorf("currency code is required")
	}

	if inv.Company == nil {
		return fmt.Errorf("company data is required")
	}
	if inv.Company.NIT == "" {
		return fmt.Errorf("company NIT is required")
	}
	if inv.Company.Name == "" {
		return fmt.Errorf("company name is required")
	}

	if inv.Customer == nil {
		return fmt.Errorf("customer data is required")
	}
	if inv.Customer.IdentificationNumber == "" {
		return fmt.Errorf("customer identification number is required")
	}
	if inv.Customer.Name == "" {
		return fmt.Errorf("customer name is required")
	}

	if inv.Resolution == nil {
		return fmt.Errorf("resolution data is required")
	}
	if inv.Resolution.Resolution == "" {
		return fmt.Errorf("resolution number is required")
	}

	if inv.Software == nil {
		return fmt.Errorf("software data is required")
	}
	if inv.Software.Identifier == "" {
		return fmt.Errorf("software identifier is required")
	}
	if inv.Software.PIN == "" {
		return fmt.Errorf("software PIN is required")
	}

	if len(inv.Lines) == 0 {
		return fmt.Errorf("invoice must have at least one line")
	}

	for i, line := range inv.Lines {
		if line.Description == "" {
			return fmt.Errorf("line %d: description is required", i+1)
		}
		if line.Quantity <= 0 {
			return fmt.Errorf("line %d: quantity must be greater than 0", i+1)
		}
		if line.UnitPrice < 0 {
			return fmt.Errorf("line %d: unit price cannot be negative", i+1)
		}
		if line.UnitCode == "" {
			return fmt.Errorf("line %d: unit code is required", i+1)
		}
		if line.TaxTypeCode == "" {
			return fmt.Errorf("line %d: tax type code is required", i+1)
		}
	}

	return nil
}
