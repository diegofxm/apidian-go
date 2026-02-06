# 📁 Storage Structure

## Overview

The APIDIAN-GO system organizes files in a structured directory hierarchy based on company NIT (Tax ID).

---

## Directory Structure

```
/var/www/apidian-go/storage/
├── {NIT}/                          # Company folder (NIT without DV)
│   ├── certificates/               # Digital certificates
│   │   ├── {NIT}_1736870400.p12   # Current active certificate
│   │   └── {NIT}_1736780000.p12   # Previous certificate (inactive)
│   ├── invoices/                   # Electronic invoices (future)
│   │   ├── xml/                   # XML files
│   │   └── pdf/                   # PDF files
│   ├── credit-notes/               # Credit notes (future)
│   │   ├── xml/
│   │   └── pdf/
│   ├── debit-notes/                # Debit notes (future)
│   │   ├── xml/
│   │   └── pdf/
│   └── support-documents/          # Support documents (future)
│       ├── xml/
│       └── pdf/
```

---

## Examples

### Example 1: Company with NIT 900123456

```
/var/www/apidian-go/storage/
└── 900123456/
    ├── certificates/
    │   ├── 900123456_1736870400.p12  # Active
    │   └── 900123456_1736780000.p12  # Inactive (historical)
    ├── invoices/
    │   ├── xml/
    │   │   ├── SETP9901234560000000001.xml
    │   │   └── SETP9901234560000000002.xml
    │   └── pdf/
    │       ├── SETP9901234560000000001.pdf
    │       └── SETP9901234560000000002.pdf
    └── credit-notes/
        ├── xml/
        └── pdf/
```

### Example 2: Company with NIT 800987654

```
/var/www/apidian-go/storage/
└── 800987654/
    ├── certificates/
    │   └── 800987654_1736870500.p12
    └── invoices/
        ├── xml/
        └── pdf/
```

---

## File Naming Conventions

### Certificates
- **Format:** `{NIT}_{timestamp}.p12`
- **Example:** `900123456_1736870400.p12`
- **Location:** `/storage/{NIT}/certificates/{NIT}_{timestamp}.p12`
- **Timestamp:** Unix timestamp (seconds since epoch) for historical tracking

### Invoices (Future Implementation)
- **XML Format:** `SETP{NIT}{consecutive}.xml`
- **PDF Format:** `SETP{NIT}{consecutive}.pdf`
- **Example:** `SETP9901234560000000001.xml`

### Credit Notes (Future Implementation)
- **XML Format:** `NC{NIT}{consecutive}.xml`
- **PDF Format:** `NC{NIT}{consecutive}.pdf`

### Debit Notes (Future Implementation)
- **XML Format:** `ND{NIT}{consecutive}.xml`
- **PDF Format:** `ND{NIT}{consecutive}.pdf`

---

## Directory Permissions

- **Company folders:** `0755` (rwxr-xr-x)
- **Subdirectories:** `0755` (rwxr-xr-x)
- **Certificate files:** `0600` (rw-------)
- **Document files:** `0644` (rw-r--r--)

---

## Implementation Details

### Certificate Storage

When a certificate is uploaded:

1. **Decode** base64 certificate to binary
2. **Create** directory structure: `/storage/{NIT}/certificates/`
3. **Generate** filename with timestamp: `{NIT}_{timestamp}.p12`
4. **Save** file to filesystem
5. **Deactivate** previous certificates in database
6. **Store** new certificate metadata (name, encrypted password, active status)

### Path Generation

```go
// GetCertificatePath returns the full path to a certificate file
// Path structure: /storage/{NIT}/certificates/{filename}
func (s *CertificateService) GetCertificatePath(filename string, companyNIT string) string {
    return filepath.Join(s.storagePath, companyNIT, "certificates", filename)
}
```

---

## Benefits

1. ✅ **Organized by Company:** Easy to locate all files for a specific company
2. ✅ **Type Separation:** Documents organized by type (certificates, invoices, etc.)
3. ✅ **Format Separation:** XML and PDF files in separate folders
4. ✅ **Scalable:** Easy to add new document types
5. ✅ **Predictable:** Consistent naming and structure
6. ✅ **Secure:** Proper file permissions per document type

---

## Historical Certificate Management

### Active vs Inactive Certificates

- **Only one certificate is active** per company at any time (`is_active = true` in database)
- **Previous certificates remain on filesystem** for historical/audit purposes
- **Filename includes timestamp** to differentiate versions: `{NIT}_{timestamp}.p12`
- **Database tracks all certificates** with creation dates and active status

### Example Timeline

```
2024-01-15: Upload certificate → 900123456_1705334400.p12 (active)
2024-06-20: Upload new cert   → 900123456_1718870400.p12 (active)
                                 900123456_1705334400.p12 (inactive)
2024-12-10: Upload new cert   → 900123456_1733875200.p12 (active)
                                 900123456_1718870400.p12 (inactive)
                                 900123456_1705334400.p12 (inactive)
```

### Benefits

- ✅ **Audit trail:** All historical certificates preserved
- ✅ **Rollback capability:** Can reactivate previous certificate if needed
- ✅ **Compliance:** Meets regulatory requirements for record keeping
- ✅ **Debugging:** Can investigate issues with previous certificates

---

## Environment Variables

- **CERTIFICATE_STORAGE_PATH:** Base path for storage (default: `/var/www/apidian-go/storage`)

---

## Future Enhancements

- [ ] Invoice XML/PDF storage
- [ ] Credit note XML/PDF storage
- [ ] Debit note XML/PDF storage
- [ ] Support document storage
- [ ] Automatic cleanup of old files
- [ ] File compression for archived documents
- [ ] Cloud storage integration (S3, etc.)
