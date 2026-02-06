# 📋 Tareas Pendientes

## 🔴 Prioridad Alta

### 1. Integración con DIAN
**Ubicación:** `internal/service/invoice_service.go:313`

**Descripción:**  
Implementar la integración completa con el sistema DIAN para el envío de facturas electrónicas.

**Tareas:**
- [ ] Investigar API oficial de DIAN
- [ ] Implementar cliente HTTP para comunicación con DIAN
- [ ] Manejar autenticación y certificados digitales
- [ ] Implementar retry logic y manejo de errores
- [ ] Validar respuestas de DIAN (CUFE, estado, etc.)
- [ ] Actualizar estado de factura según respuesta DIAN

**Código actual:**
```go
// TODO: Implementar integración con DIAN
// Por ahora solo cambiamos el estado
return s.invoiceRepo.UpdateStatus(id, "sent")
```

**Referencias:**
- Documentación DIAN: https://www.dian.gov.co/
- Validaciones DIAN ya implementadas en `pkg/validator/dian.go`

---

### 2. Almacenamiento de Certificados Digitales
**Ubicación:** `internal/handler/company_handler.go:276`

**Descripción:**  
Implementar sistema de almacenamiento seguro para certificados digitales (.p12) de las empresas.

**Tareas:**
- [ ] Definir estrategia de almacenamiento (S3, filesystem, etc.)
- [ ] Implementar encriptación de certificados
- [ ] Crear servicio de storage en `internal/infrastructure/storage/`
- [ ] Actualizar BD con path del certificado
- [ ] Implementar rotación y renovación de certificados
- [ ] Agregar validación de expiración de certificados

**Código actual:**
```go
// TODO: Guardar certificado en storage y actualizar BD
// Por ahora retorno éxito
return response.Success(c, "Certificado creado con éxito", fiber.Map{
    "company_id": id,
    "message": "Certificado subido correctamente",
})
```

**Consideraciones de seguridad:**
- Certificados deben estar encriptados en reposo
- Acceso restringido solo a procesos autorizados
- Logs de acceso a certificados
- Backup automático de certificados

---

## 🟢 Completadas Recientemente

- ✅ Refactorización de endpoints a NIT/DV
- ✅ Eliminación de código duplicado (validateNIT)
- ✅ Creación de helper de paginación
- ✅ Paginación unificada en todo el sistema
- ✅ Eliminación de métodos sin uso (GetByUserID)

---

## 📝 Notas

**Última actualización:** 2026-01-13  
**Responsable:** Equipo de desarrollo
