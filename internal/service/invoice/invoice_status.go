package invoice

import (
	"apidian-go/pkg/crypto"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/diegofxm/ubl21-dian/signature"
	"github.com/diegofxm/ubl21-dian/soap"
	"github.com/diegofxm/ubl21-dian/soap/types"
)

// GetInvoiceStatus consulta el estado de una factura en DIAN usando TrackId
func (s *InvoiceService) GetInvoiceStatus(id int64, trackID string, userID int64) error {
	// 1. Obtener factura
	invoice, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 2. Validar que esté enviada
	if invoice.Status != "sent" {
		return fmt.Errorf("invoice must be sent to DIAN first")
	}

	// 3. Obtener certificado y configurar cliente SOAP
	cert, err := s.certificateRepo.GetByCompanyID(invoice.CompanyID)
	if err != nil {
		return fmt.Errorf("no certificate found: %w", err)
	}

	decryptedPassword, err := crypto.DecryptPassword(cert.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt certificate password: %w", err)
	}

	certPath := s.storage.CertificatePath(invoice.Company.NIT, cert.Name)
	clientPemPath, err := signature.ConvertP12ToClientPEM(certPath, decryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to convert certificate: %w", err)
	}

	// 4. Crear cliente SOAP
	var environment types.Environment
	if invoice.Software.Environment == "1" {
		environment = types.Produccion
	} else {
		environment = types.Habilitacion
	}

	config := &types.Config{
		Environment: environment,
		Certificate: clientPemPath,
		PrivateKey:  clientPemPath,
	}
	client, err := soap.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating SOAP client: %w", err)
	}

	// 5. Llamar GetStatus
	statusReq := &types.GetStatusRequest{
		TrackId: trackID,
	}
	statusResp, err := client.GetStatus(statusReq)
	if err != nil {
		return fmt.Errorf("error calling GetStatus: %w", err)
	}

	// 6. Guardar ApplicationResponse FINAL (firmado por DIAN)
	if statusResp.XmlBase64Bytes != "" {
		appResponseXML, err := base64.StdEncoding.DecodeString(statusResp.XmlBase64Bytes)
		if err == nil {
			appResponsePath := s.storage.InvoiceApplicationResponsePath(invoice.Company.NIT, invoice.Number)
			if err := os.WriteFile(appResponsePath, appResponseXML, 0644); err != nil {
				fmt.Printf("Warning: Failed to save ApplicationResponse: %v\n", err)
			}
		}
	}

	// 7. Actualizar estado en BD según respuesta
	status := "rejected"
	if statusResp.IsValid {
		status = "accepted"
	}

	if err := s.invoiceRepo.UpdateDIANStatus(
		id,
		status,
		statusResp.StatusMessage,
		statusResp.StatusCode,
		statusResp.StatusDescription,
	); err != nil {
		return err
	}

	// 8. Retornar error si fue rechazado
	if !statusResp.IsValid {
		return fmt.Errorf("DIAN rejected document: %s - %s",
			statusResp.StatusCode,
			statusResp.StatusDescription)
	}

	return nil
}
