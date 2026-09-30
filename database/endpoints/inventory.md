# Inventario para el Data App

Configura cada endpoint con el método y la ruta indicados. Los nombres de parámetros, tipo y ubicación se derivan de los `${placeholders}`; los parámetros de path no se utilizan en este diseño para que el servidor Go pueda mantener un mapa de rutas fijo.

| Método | Ruta Data Service | Archivo | Paginación |
|---|---|---|---|
| GET | `/catalog` | `catalog/get-catalog.sql` | sí |
| GET | `/catalog/categories` | `catalog/get-catalog-categories.sql` | sí |
| GET | `/catalog/item` | `catalog/get-catalog-item.sql` | no |
| GET | `/auth/user` | `auth/get-auth-user.sql` | no |
| POST/GET/DELETE | `/sessions`, `/sessions/current` | `sessions/*` | no |
| GET/POST/PUT/DELETE | `/users`, `/users/item` | `users/*` | GET lista |
| POST | `/users/bootstrap` | `users/post-users-bootstrap.sql` | no |
| GET/POST/PUT/DELETE | `/supplements`, `/supplements/item` | `supplements/*` | GET lista |
| GET/POST/PUT/DELETE | `/categories`, `/categories/item` | `categories/*` | GET lista |
| GET/POST/PUT/DELETE | `/goals`, `/goals/item` | `goals/*` | GET lista |
| GET/POST/PUT/DELETE | `/ingredients`, `/ingredients/item` | `ingredients/*` | GET lista |
| GET/POST | `/orders` | `orders/get-orders.sql`, `orders/post-orders.sql` | GET |
| GET | `/orders/item` | `orders/get-orders-item.sql` | no |
| PUT | `/orders/status` | `orders/put-orders-status.sql` | no |
| GET | `/backup/{table-name}` | `backup/*` (una ruta fija por tabla) | sí, 2000 |

Usa una Data API Key `ReadAndWrite`. La clave es credencial de ejecución del Data App, no una clave administrativa de TiDB Cloud. Una vez creado el primer administrador, retira o desactiva `/users/bootstrap`.
