# Maz-Suplementos

Aplicación académica completa para catálogo, inventario y pedidos básicos de suplementos deportivos. El navegador utiliza HTML, CSS y JavaScript vanilla; un servidor Go aplica autenticación, autorización, validación y CSRF; todo el CRUD persistente se ejecuta mediante endpoints HTTP de TiDB Cloud Data Service.

## Arquitectura

```text
Navegador → servidor Go → TiDB Cloud Data Service → TiDB
```

Las credenciales de TiDB nunca llegan al navegador. Go usa Basic Authentication sobre HTTPS, mecanismo oficialmente soportado por Data Service. El cliente comprueba tanto el estado HTTP como `data.result.code`; un HTTP 200 con error SQL se considera fallo.

Roles estrictos:

- `ADMINISTRADOR`: usuarios, respaldos y lectura. No modifica catálogo.
- `CAPTURISTA`: suplementos, catálogos y estado de pedidos. No administra usuarios ni respaldos.
- `AUDITOR`: lectura exclusivamente.

## Estructura

- `cmd/server`: servidor web.
- `cmd/bootstrap-admin`: alta segura del primer administrador.
- `internal/httpapi`: rutas, sesiones, CSRF, permisos y respaldos.
- `internal/tidb`: cliente central de Data Service.
- `database/schema.sql`, `database/seeds.sql`: esquema e información demo idempotente.
- `database/endpoints`: SQL revisable e inventario de endpoints.
- `web`: tienda, carrito, checkout y panel responsivo.
- `POST /api/backups`: genera una descarga SQL protegida para administradores.

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
- `COOKIE_SECURE=true` en HTTPS; en localhost HTTP se conserva `false`.
- `APP_PORT`, `APP_ENV`, `SESSION_TTL_HOURS` son configurables.

No se necesitan claves de la API administrativa para ejecutar la aplicación. La Data API Key no crea por sí sola el Data App ni sus endpoints.

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

## Respaldos

Un administrador usa el panel o `POST /api/backups`. El servidor pagina cada tabla y responde con `maz-suplementos-<fecha>.sql` como archivo adjunto; el navegador lo guarda en la carpeta de descargas configurada por el usuario. El archivo incluye el esquema y los datos, pero no sesiones ni secretos. Al consultar tablas en peticiones separadas no se puede afirmar un snapshot global consistente; esta limitación está documentada deliberadamente.

## Verificación

```bash
make fmt
make vet
make test
make build
```

Las pruebas cubren bcrypt/login, cookies, CSRF, matriz de roles, validación, pedidos y el contrato del cliente Data Service, incluido el error interno dentro de respuestas HTTP 200.

## Referencias técnicas verificadas

- [Get Started with Data Service](https://docs.pingcap.com/tidbcloud/data-service-get-started/)
- [Manage an Endpoint](https://docs.pingcap.com/tidbcloud/data-service-manage-endpoint/)
- [API Keys in Data Service](https://docs.pingcap.com/tidbcloud/data-service-api-key/)
- [Response and HTTP Status Codes](https://docs.pingcap.com/tidbcloud/data-service-response-and-status-code/)
- [Data App Configuration Files](https://docs.pingcap.com/tidbcloud/data-service-app-config-files/)
