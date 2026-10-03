# Maz-Suplementos

Aplicación académica completa para catálogo, inventario y pedidos básicos de suplementos deportivos. El navegador utiliza HTML, CSS y JavaScript vanilla; un servidor Go aplica autenticación, autorización, validación y CSRF; todo el CRUD persistente se ejecuta mediante endpoints HTTP de TiDB Cloud Data Service.

## Arquitectura

```text
Navegador → servidor Go → TiDB Cloud Data Service → TiDB
```

Las credenciales de TiDB nunca llegan al navegador. Go usa Basic Authentication sobre HTTPS, mecanismo oficialmente soportado por Data Service. El cliente comprueba tanto el estado HTTP como `data.result.code`; un HTTP 200 con error SQL se considera fallo.

Roles estrictos:

- `ADMINISTRADOR`: usuarios, respaldos y lectura. No modifica catálogo.
- `CAPTURISTA`: suplementos, categorías y estado de pedidos. No administra usuarios ni respaldos.
- `AUDITOR`: lectura exclusivamente.

## Estructura

- `cmd/server`: servidor web.
- `cmd/bootstrap-admin`: alta segura del primer administrador.
- `internal/httpapi`: rutas, sesiones, CSRF, permisos y respaldos.
- `internal/tidb`: cliente central de Data Service.
- `database/schema.sql`, `database/seeds.sql`: esquema e información demo idempotente. El catálogo relaciona productos únicamente con categorías.
- `database/endpoints`: SQL revisable e inventario de endpoints.
- `web`: tienda, carrito, checkout y panel responsivo.
- `GET/POST /api/backups`: consulta y crea tareas reales de TiDB Cloud Export para administradores.

## Requisitos y configuración

TiDB Data Service está disponible para TiDB Cloud Starter en AWS. Crea primero el esquema con una conexión SQL temporal TLS, ejecuta los seeds y después configura/despliega los endpoints descritos en `database/endpoints/inventory.md`. Esto separa claramente la inicialización SQL del CRUD normal de la aplicación. Los comandos solicitan la contraseña de forma interactiva para no guardarla en el historial:

```bash
make db-init DB_HOST='host' DB_PORT='4000' DB_USER='usuario'
make db-seed DB_HOST='host' DB_PORT='4000' DB_USER='usuario' DB_NAME='maz_suplementos'
```

```bash
cp .env.example .env
```

Variables necesarias:

- `TIDB_DATA_SERVICE_BASE_URL`: dominio regional, por ejemplo `https://us-east-1.data.tidbcloud.com`.
- `TIDB_DATA_APP_ID`: identificador del Data App.
- `TIDB_DATA_API_PUBLIC_KEY` y `TIDB_DATA_API_PRIVATE_KEY`: clave `ReadAndWrite` del Data App.
- `TIDB_CLUSTER_ID`: identificador numérico del clúster que muestra TiDB Cloud.
- `TIDB_DATABASE`: base incluida en el export (por defecto `maz_suplementos`).
- `TIDB_CLOUD_PROFILE`: perfil autenticado de TiDB Cloud CLI (por defecto `default`).
- `TIDB_SQL_USER_PREFIX`: prefijo obligatorio de usuarios SQL mostrado por TiDB Cloud, incluido el punto final (por ejemplo `abc123.`).
- `COOKIE_SECURE=true` en HTTPS; en localhost HTTP se conserva `false`.
- `APP_PORT`, `APP_ENV`, `SESSION_TTL_HOURS` son configurables.

La Data API Key no crea por sí sola el Data App ni sus endpoints. La funcionalidad de respaldos también requiere [TiDB Cloud CLI](https://docs.pingcap.com/tidbcloud/get-started-with-cli/) en el servidor. En Windows descarga el binario oficial, coloca `ticloud.exe` en una carpeta incluida en `PATH`; en macOS/Linux puede usarse el instalador oficial indicado en esa guía. Autentica el perfil y comprueba la instalación con:

```bash
ticloud version
ticloud -P default auth login
ticloud serverless export list -c "$TIDB_CLUSTER_ID" -o json
```

## Ejecución

Con Go 1.24 o posterior:

```bash
make run
```

En NixOS o con Nix instalado:

```bash
nix develop
make run
```

Abre `http://localhost:8080`. El servidor se detiene de forma explícita si faltan credenciales; no cambia silenciosamente a datos falsos.

## Primer administrador

Despliega temporalmente `/users/bootstrap`, configura las credenciales en el entorno y ejecuta:

```bash
export BOOTSTRAP_ADMIN_USERNAME='administrador'
export BOOTSTRAP_ADMIN_FULL_NAME='Nombre del administrador'
export BOOTSTRAP_ADMIN_PASSWORD='una contraseña robusta'
make bootstrap-admin
```

La contraseña se convierte a bcrypt en Go y nunca se imprime. Después del bootstrap, retira el endpoint `/users/bootstrap` del Data App.

## Pedidos y consistencia

El carrito solo guarda en `localStorage` IDs, nombres, precios visibles, cantidades e imagen; no guarda sesión, claves ni datos del cliente. Al confirmar, el navegador genera una clave de idempotencia y el servidor deriva de ella el ID del pedido: un reintento no puede descontar stock por segunda vez. Además, los pedidos públicos están limitados por IP. El servidor envía el pedido al endpoint `POST /orders`. Su SQL inicia una transacción pesimista, bloquea productos, valida todos los stocks, crea snapshots de precios, descuenta inventario y confirma.

La documentación pública de Data Service confirma que varias sentencias se ejecutan secuencialmente y que solo se devuelve la última respuesta, pero no documenta de forma suficientemente explícita la conservación transaccional de este caso. Por ello el endpoint está diseñado de forma atómica, pero su garantía debe validarse contra la instancia real antes de producción.

## Seguridad

Sesiones opacas y CSRF se generan con `crypto/rand`; en TiDB solo se guardan hashes SHA-256. Cookies `HttpOnly`, `SameSite=Lax` y `Secure` bajo HTTPS; producción falla de forma segura si `COOKIE_SECURE` no es `true`. El token CSRF usa cookie `SameSite=Strict` y cabecera. Hay CSP sin scripts inline, límites de cuerpo y cabeceras, timeouts, rate limit de login por IP y cuenta, validaciones cerradas, bcrypt y autorización centralizada. Cambiar una contraseña revoca las sesiones activas. `.env` y los respaldos locales antiguos están ignorados.

## Respaldos con TiDB Cloud Export

Un administrador usa el panel o `POST /api/backups`. El backend ejecuta `ticloud` sin shell y con argumentos separados:

```text
ticloud -P <perfil> serverless export create -c <cluster> --target-type LOCAL --file-type SQL --compression GZIP --filter <base>.* --force --no-color
```

`GET /api/backups` obtiene la lista real con `ticloud serverless export list`; no utiliza registros simulados ni una tabla local. El Export ID, estado, fecha, formato y destino proceden de TiDB Cloud y la misma tarea aparece en **Data > Export**. El destino `LOCAL` de TiDB Cloud conserva temporalmente el archivo para descarga desde su panel.

El proceso del servidor necesita el binario `ticloud` en `PATH` y acceso al perfil indicado. Una función estándar de Vercel no incluye automáticamente ese binario ni el perfil del equipo local: para activar respaldos allí se debe provisionar ambos en el runtime o ejecutar el backend Go en un host persistente. Si faltan, la API responde `TIDB_CLI_UNAVAILABLE` en vez de simular éxito.

## Usuarios de aplicación y cuentas SQL

Al crear un usuario desde el panel, Go mantiene el hash bcrypt en `maz_suplementos.users` y también solicita `CREATE USER IF NOT EXISTS '<usuario_sql>'@'%' IDENTIFIED BY ...` mediante el endpoint protegido de Data Service. La contraseña en texto solo vive durante esa solicitud; no se almacena ni se devuelve. Las cuentas SQL no reciben `GRANT`, por lo que quedan con privilegios mínimos.

TiDB Cloud exige un prefijo propio del clúster para estas cuentas. El backend lo toma de `TIDB_SQL_USER_PREFIX`: conserva el nombre de aplicación cuando cabe dentro del límite SQL y guarda el resultado completo (`<prefijo><username>`) en `users.db_username`; si no cabe o contiene caracteres incompatibles, usa `<prefijo>app_<hash>` de forma determinista. La eliminación actual es lógica (`active=FALSE`), así que deliberadamente no ejecuta `DROP USER`: la cuenta continúa en `mysql.user`. Cambiar el nombre o contraseña de aplicación tampoco renombra ni altera automáticamente la cuenta SQL.

Verificación académica:

```sql
SELECT User, Host FROM mysql.user ORDER BY User;
SELECT id, username, db_username FROM maz_suplementos.users ORDER BY id;
```

## Verificación local

El repositorio fija Node.js 24.7.0 y npm 11.5.1 para las herramientas de calidad. Instala las dependencias reproducibles y ejecuta la verificación principal:

```bash
npm ci
make verify
```

Para las pruebas del navegador, inicia la aplicación en `http://127.0.0.1:8181` y ejecuta:

```bash
npx playwright install chromium
make test-e2e
```

Las pruebas cubren bcrypt/login, cookies, CSRF, matriz de roles, validación, pedidos, creación de suplementos con sus relaciones y el contrato del cliente Data Service, incluido el error interno dentro de respuestas HTTP 200. La cobertura Go tiene un mínimo obligatorio de 50%; Axe bloquea violaciones de accesibilidad serias o críticas y el frontend mantiene un presupuesto comprimido de 170 KiB.

## Integración continua y mantenimiento

Cada cambio enviado a `master` o propuesto mediante pull request ejecuta:

- formato, análisis estático, `go vet`, pruebas con detector de carreras y cobertura;
- compilación, presupuesto de tamaño y política de licencias;
- recorridos Playwright y cinco revisiones automáticas WCAG con Axe;
- detección de secretos, vulnerabilidades Go alcanzables y reglas Semgrep;
- CodeQL para Go y JavaScript, además de SBOM CycloneDX para ambos ecosistemas.

Dependabot revisa semanalmente módulos Go, herramientas npm y GitHub Actions con un periodo de espera de tres días. Las acciones de CI están fijadas por SHA. Los cambios en autenticación, permisos, configuración de despliegue, esquema, workflows y archivos de dependencias tienen propietario explícito en `.github/CODEOWNERS`.

Antes de publicar producción, verifica que Vercel conserve todas las variables indicadas arriba y que `COOKIE_SECURE=true`. Los flujos E2E de CI usan respuestas controladas y no escriben en TiDB; las comprobaciones contra datos reales deben realizarse deliberadamente con una cuenta de prueba.

## Referencias técnicas verificadas

- [Get Started with Data Service](https://docs.pingcap.com/tidbcloud/data-service-get-started/)
- [Manage an Endpoint](https://docs.pingcap.com/tidbcloud/data-service-manage-endpoint/)
- [API Keys in Data Service](https://docs.pingcap.com/tidbcloud/data-service-api-key/)
- [Response and HTTP Status Codes](https://docs.pingcap.com/tidbcloud/data-service-response-and-status-code/)
- [Data App Configuration Files](https://docs.pingcap.com/tidbcloud/data-service-app-config-files/)
