package pdf

import (
	"apidian-go/internal/domain"
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	marotocfg "github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// DefaultTemplate implementa el diseño por defecto de facturas
type DefaultTemplate struct{}

// NewDefaultTemplate crea una nueva instancia del template por defecto
func NewDefaultTemplate() *DefaultTemplate {
	return &DefaultTemplate{}
}

// BuildPDF construye el documento PDF completo para una factura
func (t *DefaultTemplate) BuildPDF(invoice *domain.Invoice, logoPath string) core.Maroto {
	cfg := marotocfg.NewBuilder().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithRightMargin(10).
		WithPageNumber(props.PageNumber{
			Pattern: "Página {current} de {total}",
			Place:   props.Bottom,
			Family:  "",
			Style:   fontstyle.Normal,
			Size:    7,
		}).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	t.addHeader(m, invoice, logoPath)
	m.AddRows(text.NewRow(10, "", props.Text{}))
	t.addCustomerInfo(m, invoice)
	m.AddRows(text.NewRow(5, "", props.Text{}))
	t.addItemsTable(m, invoice)
	m.AddRows(text.NewRow(3, "", props.Text{}))
	m.AddRows(text.NewRow(5, "", props.Text{}))
	t.addNotesSection(m, invoice)
	t.addFooter(m, invoice)

	return m
}

func (t *DefaultTemplate) addHeader(m core.Maroto, invoice *domain.Invoice, logoPath string) {
	company := invoice.Company
	resolution := invoice.Resolution

	companyName := "EMPRESA SAS"
	nit := "000000000-0"
	taxLevel := "No Responsable de IVA"
	regime := "Régimen Simplificado"
	resolutionText := "Resolución de Facturación Electronica"
	prefixRange := "Prefijo y Rango"
	address := "Dirección"
	phone := "Teléfono"
	email := "email@empresa.com"

	if company != nil {
		companyName = company.Name
		nit = company.NIT
		if company.DV != nil {
			nit += "-" + *company.DV
		}
		taxLevel = company.TaxLevelName
		regime = company.TypeRegimeName
		address = company.AddressLine + " - " + company.Municipality + " - " + company.Department + " - " + company.CountryName
		if company.Phone != nil {
			phone = "Telefono - " + *company.Phone
		}
		if company.Email != nil {
			email = "E-mail: " + *company.Email
		}
	}

	if resolution != nil {
		resolutionText = fmt.Sprintf("Resolución de Facturación Electronica No. %s de %s",
			resolution.Resolution,
			resolution.DateFrom.In(time.Local).Format("2006-01-02"))
		prefixRange = fmt.Sprintf("Prefijo: %s - Rango %d al %d - Vigencia Desde: %s Hasta: %s",
			resolution.Prefix,
			resolution.FromNumber,
			resolution.ToNumber,
			resolution.DateFrom.In(time.Local).Format("2006-01-02"),
			resolution.DateTo.In(time.Local).Format("2006-01-02"))
	}

	m.AddRow(25,
		col.New(2).Add(
			image.NewFromFile(logoPath, props.Rect{
				Center:  false,
				Percent: 90,
			}),
		),
		col.New(7).Add(
			text.New(companyName, props.Text{
				Top:   0,
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 102, Blue: 204},
			}),
			text.New(fmt.Sprintf("NIT: %s - Regimen: %s - Obligacion: %s", nit, taxLevel, regime), props.Text{
				Top:   6.5,
				Size:  6,
				Align: align.Center,
			}),
			text.New(resolutionText, props.Text{
				Top:   9.1,
				Size:  6,
				Align: align.Center,
			}),
			text.New(prefixRange, props.Text{
				Top:   11.7,
				Size:  6,
				Align: align.Center,
			}),
			text.New("REPRESENTACIÓN GRÁFICA DE FACTURA ELECTRÓNICA", props.Text{
				Top:   14.3,
				Size:  6,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
			text.New(address, props.Text{
				Top:   16.9,
				Size:  6,
				Align: align.Center,
			}),
			text.New(phone, props.Text{
				Top:   19.5,
				Size:  6,
				Align: align.Center,
			}),
			text.New(email, props.Text{
				Top:   22.1,
				Size:  6,
				Align: align.Center,
			}),
		),
		col.New(3).Add(
			text.New("FACTURA ELECTRÓNICA DE VENTA", props.Text{
				Top:   1,
				Size:  6,
				Style: fontstyle.Bold,
				Align: align.Right,
			}),
			text.New(invoice.Number, props.Text{
				Top:   3,
				Size:  12,
				Style: fontstyle.Bold,
				Align: align.Right,
				Color: &props.Color{Red: 255, Green: 0, Blue: 0},
			}),
			text.New(fmt.Sprintf("Fecha Emisión: %s", invoice.IssueDate.In(time.Local).Format("2006-01-02")), props.Text{
				Top:   8,
				Size:  8,
				Align: align.Right,
				Color: &props.Color{Red: 255, Green: 0, Blue: 0},
			}),
			text.New(fmt.Sprintf("Fecha Validación DIAN: %s", invoice.IssueDate.In(time.Local).Format("2006-01-02")), props.Text{
				Top:   11.5,
				Size:  6,
				Align: align.Right,
			}),
			text.New(fmt.Sprintf("Hora Validación DIAN: %s", invoice.IssueTime.In(time.Local).Format("15:04:05")), props.Text{
				Top:   14.2,
				Size:  6,
				Align: align.Right,
			}),
		),
	)
}

func (t *DefaultTemplate) addCustomerInfo(m core.Maroto, invoice *domain.Invoice) {
	customer := invoice.Customer

	identification := "000000000"
	customerName := "CLIENTE"
	regime := "Régimen"
	obligation := "Obligación"
	customerAddress := "Dirección"
	city := "Ciudad"
	customerPhone := "Teléfono"
	customerEmail := "email@cliente.com"

	if customer != nil {
		identification = customer.IdentificationNumber
		if customer.DV != nil {
			identification += "-" + *customer.DV
		}
		customerName = customer.Name
		regime = customer.TaxLevelName
		obligation = customer.TypeRegimeName
		customerAddress = customer.AddressLine
		city = customer.Municipality + " - " + customer.CountryName
		if customer.Phone != nil {
			customerPhone = *customer.Phone
		}
		if customer.Email != nil {
			customerEmail = *customer.Email
		}
	}

	paymentForm := "Contado"
	if invoice.PaymentFormName != nil {
		paymentForm = *invoice.PaymentFormName
	}

	paymentMeans := "Efectivo"
	if invoice.PaymentMethodName != nil {
		paymentMeans = *invoice.PaymentMethodName
	}

	paymentTerm := "0 Dias"
	dueDate := invoice.IssueDate.In(time.Local).Format("2006-01-02")
	if invoice.DueDate != nil && !invoice.DueDate.Equal(invoice.IssueDate) {
		dueDate = invoice.DueDate.In(time.Local).Format("2006-01-02")
		days := int(invoice.DueDate.Sub(invoice.IssueDate).Hours() / 24)
		paymentTerm = fmt.Sprintf("%d Dias", days)
	}

	qrURL := "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=no-disponible"
	if invoice.QRCodeURL != nil && *invoice.QRCodeURL != "" {
		qrURL = *invoice.QRCodeURL
	} else if invoice.UUID != nil && *invoice.UUID != "" {
		qrURL = "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=" + *invoice.UUID
	}

	m.AddRow(30,
		col.New(1).Add(
			text.New("CC o NIT:", props.Text{Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Cliente:", props.Text{Top: 3, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Regimen:", props.Text{Top: 6, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Obligación:", props.Text{Top: 9, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Dirección:", props.Text{Top: 12, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Ciudad:", props.Text{Top: 15, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Telefono:", props.Text{Top: 18, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Email:", props.Text{Top: 21, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
		),
		col.New(3).Add(
			text.New(identification, props.Text{Size: 7, Align: align.Left}),
			text.New(customerName, props.Text{Top: 3, Size: 7, Align: align.Left}),
			text.New(regime, props.Text{Top: 6, Size: 7, Align: align.Left}),
			text.New(obligation, props.Text{Top: 9, Size: 7, Align: align.Left}),
			text.New(customerAddress, props.Text{Top: 12, Size: 7, Align: align.Left}),
			text.New(city, props.Text{Top: 15, Size: 7, Align: align.Left}),
			text.New(customerPhone, props.Text{Top: 18, Size: 7, Align: align.Left}),
			text.New(customerEmail, props.Text{Top: 21, Size: 7, Align: align.Left}),
		),
		col.New(2).Add(
			text.New("Forma de Pago:", props.Text{Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Medio de Pago:", props.Text{Top: 3, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Plazo:", props.Text{Top: 6, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
			text.New("Fecha Vencimiento:", props.Text{Top: 9, Size: 7, Style: fontstyle.Bold, Align: align.Left}),
		),
		col.New(4).Add(
			text.New(paymentForm, props.Text{Size: 7, Align: align.Left}),
			text.New(paymentMeans, props.Text{Top: 3, Size: 7, Align: align.Left}),
			text.New(paymentTerm, props.Text{Top: 6, Size: 7, Align: align.Left}),
			text.New(dueDate, props.Text{Top: 9, Size: 7, Align: align.Left}),
		),
		col.New(3).Add(
			code.NewQr(qrURL, props.Rect{
				Left:    3.7,
				Percent: 90,
			}),
		),
	)
}

func (t *DefaultTemplate) addItemsTable(m core.Maroto, invoice *domain.Invoice) {
	header := []string{"#", "Código", "Descripción", "Cantidad", "UM", "Val. Unit", "IVA/IC", "Dcto", "%", "Val. Item"}
	widths := []uint{1, 1, 3, 1, 1, 1, 1, 1, 1, 1}

	headerRow := row.New(6)
	for i, h := range header {
		colWidth := int(widths[i])
		headerRow.Add(text.NewCol(colWidth, h, props.Text{
			Size:  7,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   1.5,
		}))
	}
	headerRow.WithStyle(&props.Cell{
		BackgroundColor: &props.Color{Red: 200, Green: 200, Blue: 200},
		BorderColor:     &props.Color{Red: 180, Green: 180, Blue: 180},
		BorderType:      border.Full,
		BorderThickness: 0.1,
	})
	m.AddRows(headerRow)

	for i, line := range invoice.Lines {
		bgColor := &props.Color{Red: 245, Green: 245, Blue: 245}
		if i%2 == 0 {
			bgColor = nil
		}

		rowHeight := 7.0
		if len(line.Description) > 50 {
			rowHeight = 10.0
		}

		discount := 0.0
		discountPercent := "0.00"

		dataRow := row.New(rowHeight).Add(
			text.NewCol(1, fmt.Sprintf("%d", line.LineNumber), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, line.ProductCode, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			col.New(3).Add(
				text.New(line.Description, props.Text{Size: 7, Top: 1.5}),
			),
			text.NewCol(1, fmt.Sprintf("%.2f", line.Quantity), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, line.UnitName, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, fmt.Sprintf("%.2f", line.UnitPrice), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, fmt.Sprintf("%.2f", line.TaxAmount), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, fmt.Sprintf("%.2f", discount), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, discountPercent, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, fmt.Sprintf("%.2f", line.LineTotal), props.Text{Size: 7, Align: align.Center, Top: 1.5}),
		)

		if bgColor != nil {
			dataRow.WithStyle(&props.Cell{
				BackgroundColor: bgColor,
				BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
				BorderType:      border.Full,
				BorderThickness: 0.1,
			})
		} else {
			dataRow.WithStyle(&props.Cell{
				BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
				BorderType:      border.Full,
				BorderThickness: 0.1,
			})
		}

		m.AddRows(dataRow)
	}

	t.addTotalsSection(m, invoice)
}

func (t *DefaultTemplate) addTotalsSection(m core.Maroto, invoice *domain.Invoice) {
	m.AddRows(text.NewRow(3, "", props.Text{}))

	m.AddRow(6,
		col.New(3).Add(
			text.New("Impuestos", props.Text{
				Top:   1.5,
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
		col.New(1),
		col.New(3).Add(
			text.New("Retenciones", props.Text{
				Top:   1.5,
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
		col.New(1),
		col.New(4).Add(
			text.New("Totales", props.Text{
				Top:   1.5,
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)

	headerRowTotales := row.New(6)
	headerRowTotales.Add(
		text.NewCol(1, "Tipo", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		text.NewCol(1, "Base", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		text.NewCol(1, "Porcentaje", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		col.New(1),
		text.NewCol(1, "Tipo", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		text.NewCol(1, "Base", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		text.NewCol(1, "Porcentaje", props.Text{Size: 6, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		col.New(1),
		text.NewCol(2, "Concepto", props.Text{Size: 7, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
		text.NewCol(2, "Valor", props.Text{Size: 7, Style: fontstyle.Bold, Align: align.Center, Top: 1.5}),
	)
	headerRowTotales.WithStyle(&props.Cell{
		BackgroundColor: &props.Color{Red: 200, Green: 200, Blue: 200},
		BorderColor:     &props.Color{Red: 180, Green: 180, Blue: 180},
		BorderType:      border.Full,
		BorderThickness: 0.1,
	})
	m.AddRows(headerRowTotales)

	taxRate := 0.0
	for _, line := range invoice.Lines {
		if line.TaxRate > 0 {
			taxRate = line.TaxRate
			break
		}
	}

	rows := []struct {
		taxType    string
		taxBase    string
		taxPercent string
		concept    string
		value      string
		bgColor    *props.Color
	}{
		{"IVA", fmt.Sprintf("%.2f", invoice.Subtotal), fmt.Sprintf("%.0f%%", taxRate), "Nro Lineas:", fmt.Sprintf("%d", len(invoice.Lines)), nil},
		{"", "", "", "Base:", fmt.Sprintf("%.2f", invoice.Subtotal), &props.Color{Red: 245, Green: 245, Blue: 245}},
		{"", "", "", "Impuestos:", fmt.Sprintf("%.2f", invoice.TaxTotal), nil},
		{"", "", "", "Retenciones:", "0.00", &props.Color{Red: 245, Green: 245, Blue: 245}},
		{"", "", "", "Descuentos:", "0.00", nil},
	}

	for _, r := range rows {
		fila := row.New(7)
		fila.Add(
			text.NewCol(1, r.taxType, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, r.taxBase, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			text.NewCol(1, r.taxPercent, props.Text{Size: 7, Align: align.Center, Top: 1.5}),
			col.New(1),
			text.NewCol(1, "", props.Text{Size: 7}),
			text.NewCol(1, "", props.Text{Size: 7}),
			text.NewCol(1, "", props.Text{Size: 7}),
			col.New(1),
			text.NewCol(2, r.concept, props.Text{Size: 7, Align: align.Left, Top: 1.5, Left: 1}),
			text.NewCol(2, r.value, props.Text{Size: 7, Align: align.Right, Top: 1.5, Right: 1}),
		)

		if r.bgColor != nil {
			fila.WithStyle(&props.Cell{
				BackgroundColor: r.bgColor,
				BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
				BorderType:      border.Full,
				BorderThickness: 0.1,
			})
		} else {
			fila.WithStyle(&props.Cell{
				BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
				BorderType:      border.Full,
				BorderThickness: 0.1,
			})
		}

		m.AddRows(fila)
	}

	filaTotal := row.New(8)
	filaTotal.Add(
		text.NewCol(1, "", props.Text{Size: 7}),
		text.NewCol(1, "", props.Text{Size: 7}),
		text.NewCol(1, "", props.Text{Size: 7}),
		col.New(1),
		text.NewCol(1, "", props.Text{Size: 7}),
		text.NewCol(1, "", props.Text{Size: 7}),
		text.NewCol(1, "", props.Text{Size: 7}),
		col.New(1),
		text.NewCol(2, "Total Factura:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left, Top: 2, Left: 1}),
		text.NewCol(2, fmt.Sprintf("%.2f", invoice.Total), props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right, Top: 2, Right: 1, Color: &props.Color{Red: 0, Green: 100, Blue: 0}}),
	)
	filaTotal.WithStyle(&props.Cell{
		BackgroundColor: &props.Color{Red: 245, Green: 245, Blue: 245},
		BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
		BorderType:      border.Full,
		BorderThickness: 0.1,
	})
	m.AddRows(filaTotal)
}

func (t *DefaultTemplate) addNotesSection(m core.Maroto, invoice *domain.Invoice) {
	notes := "Sin notas adicionales."
	if invoice.Notes != nil && *invoice.Notes != "" {
		notes = *invoice.Notes
	}

	notasRow := row.New(25)
	notasRow.Add(
		col.New(12).Add(
			text.New("NOTAS:", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Top:   2,
				Left:  2,
			}),
			text.New(notes, props.Text{
				Size:  7,
				Style: fontstyle.Italic,
				Top:   6,
				Left:  2,
			}),
		),
	)
	notasRow.WithStyle(&props.Cell{
		BorderColor:     &props.Color{Red: 220, Green: 220, Blue: 220},
		BorderType:      border.Full,
		BorderThickness: 0.1,
	})
	m.AddRows(notasRow)
}

func (t *DefaultTemplate) addFooter(m core.Maroto, invoice *domain.Invoice) {
	footerHeight := 12.0
	espacioMinimo := 10.0

	espacioMaximo := 250.0
	espacioOptimo := espacioMinimo

	for espacio := espacioMinimo; espacio <= espacioMaximo; espacio += 1.0 {
		if m.FitlnCurrentPage(espacio + footerHeight) {
			espacioOptimo = espacio
		} else {
			break
		}
	}

	m.AddRows(text.NewRow(espacioOptimo, "", props.Text{}))

	lineaFooter := row.New(3)
	lineaFooter.Add(col.New(12))
	lineaFooter.WithStyle(&props.Cell{
		BorderColor:     &props.Color{Red: 180, Green: 180, Blue: 180},
		BorderType:      border.Top,
		BorderThickness: 0.1,
	})
	m.AddRows(lineaFooter)

	issueDateTime := invoice.IssueDate.In(time.Local).Format("2006-01-02") + " - " + invoice.IssueTime.In(time.Local).Format("15:04:05")
	m.AddRows(
		text.NewRow(3, fmt.Sprintf("Factura No: %s - Fecha y Hora de Generación: %s", invoice.Number, issueDateTime), props.Text{
			Size:  6,
			Align: align.Center,
		}),
	)

	cufe := "CUFE: no-disponible-factura-borrador-no-disponible-factura-borrador-no-disponible-factura-borrador-no-disponible-factura-borrador"
	if invoice.UUID != nil && *invoice.UUID != "" {
		cufe = "CUFE: " + *invoice.UUID
	}
	m.AddRows(
		text.NewRow(3, cufe, props.Text{
			Size:  6,
			Align: align.Center,
		}),
	)

	dianText := "DOCUMENTO ELECTRÓNICO TRANSMITIDO Y ACEPTADO POR LA DIAN"
	textColor := &props.Color{Red: 0, Green: 100, Blue: 0}
	textSize := 6.0

	if invoice.Status == "draft" || invoice.UUID == nil || *invoice.UUID == "" {
		dianText = "*** PREVIEW - DOCUMENTO SIN FIRMAR - NO VALIDO ANTE LA DIAN ***"
		textColor = &props.Color{Red: 255, Green: 0, Blue: 0}
		textSize = 8.0
	}

	m.AddRows(
		text.NewRow(3, dianText, props.Text{
			Size:  textSize,
			Align: align.Center,
			Style: fontstyle.Bold,
			Color: textColor,
		}),
	)
}
