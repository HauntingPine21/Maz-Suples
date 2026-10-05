# Registro de cambios

Todos los cambios relevantes de Maz-Suplementos se documentan en este archivo.

## [Sin publicar]

### Cambiado

- Se retiraron objetivos, ingredientes y sus relaciones del formulario de productos, API, respaldos, Data Service y esquema TiDB.
- Se sustituyó la descarga SQL construida por la aplicación por tareas reales de TiDB Cloud Export.
- La baja lógica de usuarios se conserva y no elimina automáticamente su cuenta SQL.

### Añadido

- Canalización de CI con lint, cobertura, race detector, build, navegador y accesibilidad.
- Escaneos con govulncheck, Gitleaks, Semgrep y CodeQL.
- SBOM CycloneDX para Go y npm.
- Dependabot con enfriamiento de tres días y propiedad de código sensible.
- Presupuesto comprimido de JavaScript y política de licencias.
- Sincronización de cada nuevo usuario de aplicación con una cuenta creada mediante `CREATE USER` y registrada en `mysql.user`.
- Columna `users.db_username`, migración de datos y consulta de exports reales desde el panel administrativo.
- Prefijo SQL obligatorio de TiDB Cloud configurable mediante `TIDB_SQL_USER_PREFIX`.

## [1.0.0] - 2026-09-30

### Añadido

- Catálogo público, carrito y captura de pedidos.
- Administración por roles: administrador, capturista y auditor.
- Integración con TiDB Cloud Data Service.
- Respaldos SQL descargables.
- Despliegue productivo en Vercel.
