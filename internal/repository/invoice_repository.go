package repository

import (
	"apidian-go/internal/domain"
	"apidian-go/internal/infrastructure/database"
	"database/sql"
	"fmt"
	"time"
)

type InvoiceRepository struct {
	db *database.Database
}

func NewInvoiceRepository(db *database.Database) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// Create crea una nueva factura con sus líneas
func (r *InvoiceRepository) Create(invoice *domain.Invoice, lines []domain.InvoiceLine) error {
	tx, err := r.db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Insertar documento (factura) - UUID se generará al firmar (CUFE)
	query := `
		INSERT INTO documents (
			company_id, customer_id, resolution_id, number, consecutive,
			issue_date, issue_time, due_date, type_document_id, currency_code_id,
			notes, payment_method_id, payment_form_id,
			subtotal, tax_total, total, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		query,
		invoice.CompanyID,
		invoice.CustomerID,
		invoice.ResolutionID,
		invoice.Number,
		invoice.Consecutive,
		invoice.IssueDate,
		invoice.IssueTime,
		invoice.DueDate,
		invoice.TypeDocumentID,
		invoice.CurrencyCodeID,
		invoice.Notes,
		invoice.PaymentMethodID,
		invoice.PaymentFormID,
		invoice.Subtotal,
		invoice.TaxTotal,
		invoice.Total,
		invoice.Status,
	).Scan(&invoice.ID, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating invoice: %w", err)
	}

	// Insertar líneas
	for i, line := range lines {
		lineQuery := `
			INSERT INTO document_lines (
				document_id, product_id, line_number, description,
				quantity, unit_price, line_total, tax_rate, tax_amount,
				brand_name, model_name, standard_item_code, classification_code,
				created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
			RETURNING id, created_at
		`

		err = tx.QueryRow(
			lineQuery,
			invoice.ID,
			line.ProductID,
			i+1, // line_number
			line.Description,
			line.Quantity,
			line.UnitPrice,
			line.LineTotal,
			line.TaxRate,
			line.TaxAmount,
			line.BrandName,
			line.ModelName,
			line.StandardItemCode,
			line.ClassificationCode,
		).Scan(&lines[i].ID, &lines[i].CreatedAt)

		if err != nil {
			return fmt.Errorf("error creating invoice line %d: %w", i+1, err)
		}
		lines[i].DocumentID = invoice.ID
		lines[i].LineNumber = int64(i + 1)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Note: No asignamos lines a invoice.Lines porque son tipos diferentes
	// invoice.Lines es []InvoiceLineDetail (con JOINs) y lines es []InvoiceLine (sin JOINs)
	// Si se necesita retornar la factura completa, usar GetByID() después de crear
	return nil
}

// GetByID obtiene una factura por ID con todos los datos necesarios para DIAN (JOINs completos)
func (r *InvoiceRepository) GetByID(id int64) (*domain.Invoice, error) {
	query := `
		SELECT 
			-- Documento base
			d.id, d.company_id, d.customer_id, d.resolution_id, d.number, d.consecutive,
			d.uuid, d.issue_date, d.issue_time, d.due_date, d.type_document_id, d.currency_code_id,
			d.notes, d.payment_method_id, d.payment_form_id,
			d.subtotal, d.tax_total, d.total,
			d.xml_path, d.pdf_path, d.zip_path, d.qr_code_url, d.track_id,
			d.status, d.dian_status, d.dian_response, d.dian_status_code, d.dian_status_description,
			d.sent_to_dian_at, d.accepted_by_dian_at,
			d.created_at, d.updated_at,
			
			-- Códigos DIAN
			itc.code AS invoice_type_code,
			cc.code AS currency_code,
			pm.code AS payment_method_code,
			pm.name AS payment_method_name,
			pf.code AS payment_form_code,
			pf.name AS payment_form_name,
			
			-- Company (Emisor)
			c.id AS company_id_detail,
			c.nit AS company_nit,
			c.dv AS company_dv,
			c.name AS company_name,
			c.trade_name AS company_trade_name,
			c.registration_name AS company_registration_name,
			dt_c.code AS company_document_type_code,
			dt_c.name AS company_document_type_name,
			tlc_c.code AS company_tax_level_code,
			tlc_c.name AS company_tax_level_name,
			to_c.code AS company_type_organization_code,
			tr_c.code AS company_type_regime_code,
			tr_c.name AS company_type_regime_name,
			c.industry_codes AS company_industry_codes,
			c.address_line AS company_address_line,
			c.postal_zone AS company_postal_zone,
			c.phone AS company_phone,
			c.email AS company_email,
			c.website AS company_website,
			c.logo_path AS company_logo_path,
			mun_c.name AS company_municipality,
			mun_c.code AS company_municipality_code,
			dep_c.name AS company_department,
			dep_c.code AS company_department_code,
			country_c.code AS company_country_code,
			country_c.name AS company_country_name,
			tt_c.code AS company_tax_scheme_id,
			tt_c.name AS company_tax_scheme_name,
			
			-- Customer (Adquiriente)
			cust.id AS customer_id_detail,
			cust.identification_number AS customer_identification_number,
			cust.dv AS customer_dv,
			cust.name AS customer_name,
			cust.trade_name AS customer_trade_name,
			dt_cust.code AS customer_document_type_code,
			dt_cust.name AS customer_document_type_name,
			tlc_cust.code AS customer_tax_level_code,
			tlc_cust.name AS customer_tax_level_name,
			to_cust.code AS customer_type_organization_code,
			tr_cust.code AS customer_type_regime_code,
			tr_cust.name AS customer_type_regime_name,
			cust.address_line AS customer_address_line,
			cust.postal_zone AS customer_postal_zone,
			cust.phone AS customer_phone,
			cust.email AS customer_email,
			mun_cust.name AS customer_municipality,
			mun_cust.code AS customer_municipality_code,
			dep_cust.name AS customer_department,
			dep_cust.code AS customer_department_code,
			country_cust.code AS customer_country_code,
			country_cust.name AS customer_country_name,
			tt_cust.code AS customer_tax_scheme_id,
			tt_cust.name AS customer_tax_scheme_name,
			
			-- Resolution
			r.id AS resolution_id_detail,
			r.prefix AS resolution_prefix,
			r.resolution AS resolution_resolution,
			r.technical_key AS resolution_technical_key,
			r.from_number AS resolution_from_number,
			r.to_number AS resolution_to_number,
			r.date_from AS resolution_date_from,
			r.date_to AS resolution_date_to,
			
			-- Software
			s.id AS software_id,
			s.identifier AS software_identifier,
			s.pin AS software_pin,
			s.environment AS software_environment,
			s.test_set_id AS software_test_set_id
			
		FROM documents d
		
		-- JOINs EMISOR
		INNER JOIN companies c ON d.company_id = c.id
		INNER JOIN document_types dt_c ON c.document_type_id = dt_c.id
		INNER JOIN tax_level_codes tlc_c ON c.tax_level_code_id = tlc_c.id
		INNER JOIN organization_types to_c ON c.type_organization_id = to_c.id
		INNER JOIN regime_types tr_c ON c.type_regime_id = tr_c.id
		INNER JOIN municipalities mun_c ON c.municipality_id = mun_c.id
		INNER JOIN departments dep_c ON c.department_id = dep_c.id
		INNER JOIN countries country_c ON c.country_id = country_c.id
		LEFT JOIN tax_types tt_c ON c.tax_type_id = tt_c.id
		
		-- JOINs ADQUIRIENTE
		INNER JOIN customers cust ON d.customer_id = cust.id
		INNER JOIN document_types dt_cust ON cust.document_type_id = dt_cust.id
		INNER JOIN tax_level_codes tlc_cust ON cust.tax_level_code_id = tlc_cust.id
		INNER JOIN organization_types to_cust ON cust.type_organization_id = to_cust.id
		INNER JOIN regime_types tr_cust ON cust.type_regime_id = tr_cust.id
		INNER JOIN municipalities mun_cust ON cust.municipality_id = mun_cust.id
		INNER JOIN departments dep_cust ON cust.department_id = dep_cust.id
		INNER JOIN countries country_cust ON cust.country_id = country_cust.id
		LEFT JOIN tax_types tt_cust ON cust.tax_type_id = tt_cust.id
		
		-- JOINs RESOLUCIÓN Y SOFTWARE
		INNER JOIN resolutions r ON d.resolution_id = r.id
		INNER JOIN software s ON c.id = s.company_id
		
		-- JOINs CÓDIGOS DIAN
		INNER JOIN invoice_type_codes itc ON d.type_document_id = itc.id
		INNER JOIN currency_codes cc ON d.currency_code_id = cc.id
		LEFT JOIN payment_methods pm ON d.payment_method_id = pm.id
		LEFT JOIN payment_forms pf ON d.payment_form_id = pf.id
		
		WHERE d.id = $1 AND d.type_document_id = 1
	`

	invoice := &domain.Invoice{}
	company := &domain.CompanyDetail{}
	customer := &domain.CustomerDetail{}
	resolution := &domain.ResolutionDetail{}
	software := &domain.SoftwareDetail{}
	
	err := r.db.DB.QueryRow(query, id).Scan(
		// Documento base
		&invoice.ID,
		&invoice.CompanyID,
		&invoice.CustomerID,
		&invoice.ResolutionID,
		&invoice.Number,
		&invoice.Consecutive,
		&invoice.UUID,
		&invoice.IssueDate,
		&invoice.IssueTime,
		&invoice.DueDate,
		&invoice.TypeDocumentID,
		&invoice.CurrencyCodeID,
		&invoice.Notes,
		&invoice.PaymentMethodID,
		&invoice.PaymentFormID,
		&invoice.Subtotal,
		&invoice.TaxTotal,
		&invoice.Total,
		&invoice.XMLPath,
		&invoice.PDFPath,
		&invoice.ZipPath,
		&invoice.QRCodeURL,
		&invoice.TrackID,
		&invoice.Status,
		&invoice.DIANStatus,
		&invoice.DIANResponse,
		&invoice.DIANStatusCode,
		&invoice.DIANStatusDescription,
		&invoice.SentToDIANAt,
		&invoice.AcceptedByDIANAt,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
		
		// Códigos DIAN
		&invoice.InvoiceTypeCode,
		&invoice.CurrencyCode,
		&invoice.PaymentMethodCode,
		&invoice.PaymentMethodName,
		&invoice.PaymentFormCode,
		&invoice.PaymentFormName,
		
		// Company
		&company.ID,
		&company.NIT,
		&company.DV,
		&company.Name,
		&company.TradeName,
		&company.RegistrationName,
		&company.DocumentTypeCode,
		&company.DocumentTypeName,
		&company.TaxLevelCode,
		&company.TaxLevelName,
		&company.TypeOrganizationCode,
		&company.TypeRegimeCode,
		&company.TypeRegimeName,
		&company.IndustryCodes,
		&company.AddressLine,
		&company.PostalZone,
		&company.Phone,
		&company.Email,
		&company.Website,
		&company.LogoPath,
		&company.Municipality,
		&company.MunicipalityCode,
		&company.Department,
		&company.DepartmentCode,
		&company.CountryCode,
		&company.CountryName,
		&company.TaxSchemeID,
		&company.TaxSchemeName,
		
		// Customer
		&customer.ID,
		&customer.IdentificationNumber,
		&customer.DV,
		&customer.Name,
		&customer.TradeName,
		&customer.DocumentTypeCode,
		&customer.DocumentTypeName,
		&customer.TaxLevelCode,
		&customer.TaxLevelName,
		&customer.TypeOrganizationCode,
		&customer.TypeRegimeCode,
		&customer.TypeRegimeName,
		&customer.AddressLine,
		&customer.PostalZone,
		&customer.Phone,
		&customer.Email,
		&customer.Municipality,
		&customer.MunicipalityCode,
		&customer.Department,
		&customer.DepartmentCode,
		&customer.CountryCode,
		&customer.CountryName,
		&customer.TaxSchemeID,
		&customer.TaxSchemeName,
		
		// Resolution
		&resolution.ID,
		&resolution.Prefix,
		&resolution.Resolution,
		&resolution.TechnicalKey,
		&resolution.FromNumber,
		&resolution.ToNumber,
		&resolution.DateFrom,
		&resolution.DateTo,
		
		// Software
		&software.ID,
		&software.Identifier,
		&software.PIN,
		&software.Environment,
		&software.TestSetID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invoice not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting invoice: %w", err)
	}
	
	// Asignar datos anidados
	invoice.Company = company
	invoice.Customer = customer
	invoice.Resolution = resolution
	invoice.Software = software

	// Obtener líneas con JOINs
	lines, err := r.GetLinesDetailByDocumentID(invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Lines = lines

	return invoice, nil
}

// GetLinesByDocumentID obtiene las líneas de un documento (sin JOINs, para compatibilidad)
// DEPRECATED: Usar GetLinesDetailByDocumentID para datos completos
func (r *InvoiceRepository) GetLinesByDocumentID(documentID int64) ([]domain.InvoiceLine, error) {
	query := `
		SELECT 
			id, document_id, product_id, line_number, description,
			quantity, unit_price, line_total, tax_rate, tax_amount,
			brand_name, model_name, standard_item_code, classification_code,
			created_at
		FROM document_lines
		WHERE document_id = $1
		ORDER BY line_number ASC
	`

	rows, err := r.db.DB.Query(query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []domain.InvoiceLine
	for rows.Next() {
		var line domain.InvoiceLine
		err := rows.Scan(
			&line.ID,
			&line.DocumentID,
			&line.ProductID,
			&line.LineNumber,
			&line.Description,
			&line.Quantity,
			&line.UnitPrice,
			&line.LineTotal,
			&line.TaxRate,
			&line.TaxAmount,
			&line.BrandName,
			&line.ModelName,
			&line.StandardItemCode,
			&line.ClassificationCode,
			&line.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// GetLinesDetailByDocumentID obtiene las líneas de un documento con JOINs completos
func (r *InvoiceRepository) GetLinesDetailByDocumentID(documentID int64) ([]domain.InvoiceLineDetail, error) {
	query := `
		SELECT 
			-- Campos base (document_lines)
			dl.id, dl.document_id, dl.product_id, dl.line_number, dl.description,
			dl.quantity, dl.unit_price, dl.line_total, dl.tax_rate, dl.tax_amount,
			dl.brand_name, dl.model_name, dl.standard_item_code, dl.classification_code,
			dl.created_at,
			
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
		ORDER BY dl.line_number ASC
	`

	rows, err := r.db.DB.Query(query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []domain.InvoiceLineDetail
	for rows.Next() {
		var line domain.InvoiceLineDetail
		err := rows.Scan(
			// Campos base
			&line.ID,
			&line.DocumentID,
			&line.ProductID,
			&line.LineNumber,
			&line.Description,
			&line.Quantity,
			&line.UnitPrice,
			&line.LineTotal,
			&line.TaxRate,
			&line.TaxAmount,
			&line.BrandName,
			&line.ModelName,
			&line.StandardItemCode,
			&line.ClassificationCode,
			&line.CreatedAt,
			
			// Producto
			&line.ProductCode,
			&line.ProductName,
			&line.ProductStandardCode,
			&line.UNSPSCCode,
			
			// Unidad
			&line.UnitCode,
			&line.UnitName,
			
			// Impuesto
			&line.TaxTypeCode,
			&line.TaxTypeName,
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// GetByCompanyID obtiene todas las facturas de una empresa
func (r *InvoiceRepository) GetByCompanyID(companyID int64, limit, offset int) ([]domain.Invoice, int64, error) {
	// Contar total
	var total int64
	countQuery := `SELECT COUNT(*) FROM documents WHERE company_id = $1 AND type_document_id = 1`
	err := r.db.DB.QueryRow(countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener facturas
	query := `
		SELECT 
			id, company_id, customer_id, resolution_id, number, consecutive,
			uuid, issue_date, issue_time, due_date, type_document_id, currency_code_id,
			notes, payment_method_id, payment_form_id,
			subtotal, tax_total, total,
			xml_path, pdf_path, zip_path, qr_code_url,
			status, dian_status, dian_response, dian_status_code, dian_status_description,
			sent_to_dian_at, accepted_by_dian_at,
			created_at, updated_at
		FROM documents
		WHERE company_id = $1 AND type_document_id = 1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.DB.Query(query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.CompanyID,
			&invoice.CustomerID,
			&invoice.ResolutionID,
			&invoice.Number,
			&invoice.Consecutive,
			&invoice.UUID,
			&invoice.IssueDate,
			&invoice.IssueTime,
			&invoice.DueDate,
			&invoice.TypeDocumentID,
			&invoice.CurrencyCodeID,
			&invoice.Notes,
			&invoice.PaymentMethodID,
			&invoice.PaymentFormID,
			&invoice.Subtotal,
			&invoice.TaxTotal,
			&invoice.Total,
			&invoice.XMLPath,
			&invoice.PDFPath,
			&invoice.ZipPath,
			&invoice.QRCodeURL,
			&invoice.Status,
			&invoice.DIANStatus,
			&invoice.DIANResponse,
			&invoice.DIANStatusCode,
			&invoice.DIANStatusDescription,
			&invoice.SentToDIANAt,
			&invoice.AcceptedByDIANAt,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, total, nil
}

// Update actualiza una factura (solo campos editables)
func (r *InvoiceRepository) Update(invoice *domain.Invoice) error {
	query := `
		UPDATE documents
		SET 
			due_date = $1,
			notes = $2,
			payment_method_id = $3,
			payment_form_id = $4,
			updated_at = NOW()
		WHERE id = $5 AND type_document_id = 1
		RETURNING updated_at
	`

	err := r.db.DB.QueryRow(
		query,
		invoice.DueDate,
		invoice.Notes,
		invoice.PaymentMethodID,
		invoice.PaymentFormID,
		invoice.ID,
	).Scan(&invoice.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("invoice not found")
	}
	return err
}

// UpdateStatus actualiza el estado de una factura
func (r *InvoiceRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE documents
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, status, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// Delete elimina una factura (solo si está en draft)
func (r *InvoiceRepository) Delete(id int64) error {
	query := `
		DELETE FROM documents
		WHERE id = $1 AND type_document_id = 1 AND status = 'draft'
	`

	result, err := r.db.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found or cannot be deleted (only draft invoices can be deleted)")
	}

	return nil
}

// UpdateDIANStatus actualiza el estado DIAN de una factura
func (r *InvoiceRepository) UpdateDIANStatus(id int64, dianStatus, dianResponse, dianStatusCode, dianStatusDescription string) error {
	query := `
		UPDATE documents
		SET 
			dian_status = $1,
			dian_response = $2,
			dian_status_code = $3,
			dian_status_description = $4,
			sent_to_dian_at = CASE WHEN sent_to_dian_at IS NULL THEN NOW() ELSE sent_to_dian_at END,
			accepted_by_dian_at = CASE WHEN $1 = 'accepted' THEN NOW() ELSE accepted_by_dian_at END,
			updated_at = NOW()
		WHERE id = $5 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, dianStatus, dianResponse, dianStatusCode, dianStatusDescription, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdateIssueDateAndTime actualiza la fecha y hora de emisión de una factura
func (r *InvoiceRepository) UpdateIssueDateAndTime(id int64, issueDate time.Time, issueTime time.Time) error {
	query := `
		UPDATE documents
		SET issue_date = $1, issue_time = $2, updated_at = NOW()
		WHERE id = $3 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, issueDate, issueTime, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdateUUID actualiza el UUID (CUFE) de una factura
func (r *InvoiceRepository) UpdateUUID(id int64, uuid string) error {
	query := `
		UPDATE documents
		SET uuid = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, uuid, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdateXMLPath actualiza la ruta del XML firmado
func (r *InvoiceRepository) UpdateXMLPath(id int64, xmlPath string) error {
	query := `
		UPDATE documents
		SET xml_path = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, xmlPath, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdatePDFPath actualiza la ruta del PDF
func (r *InvoiceRepository) UpdatePDFPath(id int64, pdfPath string) error {
	query := `
		UPDATE documents
		SET pdf_path = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, pdfPath, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// GetByNumber busca una factura por su número completo (con prefijo)
func (r *InvoiceRepository) GetByNumber(number string) (*domain.Invoice, error) {
	// Buscar por número completo (ej: SETP990000003)
	var id int64
	query := `SELECT id FROM documents WHERE number = $1 AND type_document_id = 1`
	err := r.db.DB.QueryRow(query, number).Scan(&id)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	// Luego obtener la factura completa con todos sus datos relacionados
	return r.GetByID(id)
}

// UpdateZIPPath actualiza la ruta del ZIP final
func (r *InvoiceRepository) UpdateZIPPath(id int64, zipPath string) error {
	query := `
		UPDATE documents
		SET zip_path = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, zipPath, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdateTrackId actualiza el TrackId retornado por DIAN
func (r *InvoiceRepository) UpdateTrackId(id int64, trackId string) error {
	query := `
		UPDATE documents
		SET track_id = $1, updated_at = NOW()
		WHERE id = $2 AND type_document_id = 1
	`

	result, err := r.db.DB.Exec(query, trackId, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}
