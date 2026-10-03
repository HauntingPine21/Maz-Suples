package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"maz-suplementos/internal/cloudexport"
	"maz-suplementos/internal/identity"
	"maz-suplementos/internal/models"
	"maz-suplementos/internal/tidb"
	"maz-suplementos/internal/validation"
)

const sessionCookie = "maz_session"
const csrfCookie = "maz_csrf"

type Server struct {
	db            tidb.Caller
	exporter      cloudexport.Exporter
	web           http.Handler
	secureCookie  bool
	sessionTTL    time.Duration
	logger        *slog.Logger
	sqlUserPrefix string
	attemptMu     sync.Mutex
	loginAttempts map[string][]time.Time
	orderAttempts map[string][]time.Time
}

type principalKey struct{}

func New(db tidb.Caller, exporter cloudexport.Exporter, web http.Handler, secureCookie bool, ttl time.Duration, logger *slog.Logger, sqlUserPrefix ...string) *Server {
	prefix := ""
	if len(sqlUserPrefix) > 0 {
		prefix = sqlUserPrefix[0]
	}
	return &Server{db: db, exporter: exporter, web: web, secureCookie: secureCookie, sessionTTL: ttl, logger: logger, sqlUserPrefix: prefix, loginAttempts: map[string][]time.Time{}, orderAttempts: map[string][]time.Time{}}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.Handle("POST /api/auth/logout", s.requireAuth(http.HandlerFunc(s.logout)))
	mux.Handle("GET /api/auth/me", s.requireAuth(http.HandlerFunc(s.me)))
	mux.HandleFunc("GET /api/catalog", s.catalog)
	mux.HandleFunc("GET /api/catalog/categories", s.catalogCategories)
	mux.HandleFunc("GET /api/catalog/{id}", s.catalogDetail)
	mux.HandleFunc("POST /api/orders", s.createOrder)

	mux.Handle("GET /api/users", s.roles(models.RoleAdmin, models.RoleAuditor)(http.HandlerFunc(s.listUsers)))
	mux.Handle("GET /api/users/{id}", s.roles(models.RoleAdmin, models.RoleAuditor)(http.HandlerFunc(s.getUser)))
	mux.Handle("POST /api/users", s.roles(models.RoleAdmin)(http.HandlerFunc(s.createUser)))
	mux.Handle("PUT /api/users/{id}", s.roles(models.RoleAdmin)(http.HandlerFunc(s.updateUser)))
	mux.Handle("DELETE /api/users/{id}", s.roles(models.RoleAdmin)(http.HandlerFunc(s.disableUser)))

	for _, r := range []string{"supplements", "categories"} {
		resource := r
		mux.Handle("GET /api/"+resource, s.roles(models.RoleAdmin, models.RoleCapturer, models.RoleAuditor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.listResource(w, r, resource) })))
		if resource == "supplements" {
			mux.Handle("GET /api/"+resource+"/{id}", s.roles(models.RoleAdmin, models.RoleCapturer, models.RoleAuditor)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.getResource(w, r, resource) })))
		}
		mux.Handle("POST /api/"+resource, s.roles(models.RoleCapturer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.createResource(w, r, resource) })))
		mux.Handle("PUT /api/"+resource+"/{id}", s.roles(models.RoleCapturer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.updateResource(w, r, resource) })))
		mux.Handle("DELETE /api/"+resource+"/{id}", s.roles(models.RoleCapturer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.deleteResource(w, r, resource) })))
	}
	mux.Handle("GET /api/orders", s.roles(models.RoleAdmin, models.RoleCapturer, models.RoleAuditor)(http.HandlerFunc(s.listOrders)))
	mux.Handle("GET /api/orders/{id}", s.roles(models.RoleAdmin, models.RoleCapturer, models.RoleAuditor)(http.HandlerFunc(s.getOrder)))
	mux.Handle("PUT /api/orders/{id}/status", s.roles(models.RoleCapturer)(http.HandlerFunc(s.updateOrderStatus)))
	mux.Handle("POST /api/backups", s.roles(models.RoleAdmin)(http.HandlerFunc(s.backup)))
	mux.Handle("GET /api/backups", s.roles(models.RoleAdmin)(http.HandlerFunc(s.listBackups)))
	mux.Handle("/", s.web)
	return s.securityHeaders(s.recoverPanic(s.requestLog(mux)))
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	if !decode(w, r, &in) {
		return
	}
	username, err := validation.Text(in.Username, "username", 64, true)
	if err != nil || in.Password == "" || len(in.Password) > 128 {
		fail(w, http.StatusBadRequest, "VALIDATION_ERROR", "Credenciales inválidas")
		return
	}
	if !s.allowAttempt(s.loginAttempts, "ip:"+clientIP(r.RemoteAddr), 10) || !s.allowAttempt(s.loginAttempts, "account:"+strings.ToLower(username), 10) {
		fail(w, http.StatusTooManyRequests, "RATE_LIMITED", "Demasiados intentos; espera un minuto")
		return
	}
	res, err := s.db.Call(r.Context(), http.MethodGet, "auth/user", map[string]any{"username": username})
	if err != nil || len(res.Data.Rows) != 1 {
		s.logger.Warn("login denied", "username", username)
		fail(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Usuario o contraseña incorrectos")
		return
	}
	row := res.Data.Rows[0]
	if !asBool(row["active"]) || bcrypt.CompareHashAndPassword([]byte(asString(row["password_hash"])), []byte(in.Password)) != nil {
		s.logger.Warn("login denied", "username", username)
		fail(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Usuario o contraseña incorrectos")
		return
	}
	token := randomToken(32)
	csrf := randomToken(32)
	expires := time.Now().UTC().Add(s.sessionTTL)
	_, err = s.db.Call(r.Context(), http.MethodPost, "sessions", map[string]any{"token_hash": hash(token), "csrf_hash": hash(csrf), "user_id": asInt64(row["id"]), "expires_at": expires.Format("2006-01-02 15:04:05")})
	if err != nil {
		s.dataError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: s.secureCookie, SameSite: http.SameSiteLaxMode, Expires: expires, MaxAge: int(s.sessionTTL.Seconds())})
	http.SetCookie(w, &http.Cookie{Name: csrfCookie, Value: csrf, Path: "/", HttpOnly: false, Secure: s.secureCookie, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(s.sessionTTL.Seconds())})
	s.logger.Info("login succeeded", "user_id", asInt64(row["id"]))
	writeJSON(w, http.StatusOK, map[string]any{"user": publicUser(row), "csrf_token": csrf})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(sessionCookie)
	if c != nil {
		_, _ = s.db.Call(r.Context(), http.MethodDelete, "sessions", map[string]any{"token_hash": hash(c.Value)})
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", HttpOnly: true, Secure: s.secureCookie, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	http.SetCookie(w, &http.Cookie{Name: csrfCookie, Value: "", Path: "/", HttpOnly: false, Secure: s.secureCookie, SameSite: http.SameSiteStrictMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": principal(r.Context())})
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("view") == "categories" {
		s.catalogCategories(w, r)
		return
	}
	params, ok := listParams(w, r)
	if !ok {
		return
	}
	res, err := s.db.Call(r.Context(), http.MethodGet, "catalog", params)
	if err != nil {
		s.dataError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res.Data.Rows)
}
func (s *Server) catalogCategories(w http.ResponseWriter, r *http.Request) {
	s.callRows(w, r, http.MethodGet, "catalog/categories", map[string]any{"page": 1, "page_size": 100}, http.StatusOK, false)
}
func (s *Server) catalogDetail(w http.ResponseWriter, r *http.Request) {
	s.publicGet(w, r, "catalog/item")
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	params, ok := listParams(w, r)
	if !ok {
		return
	}
	s.callRows(w, r, http.MethodGet, "users", params, http.StatusOK, true)
}
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	s.callOne(w, r, http.MethodGet, "users/item", map[string]any{"id": r.PathValue("id")}, true)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	in, ok := userInput(w, r, true)
	if !ok {
		return
	}
	existing, err := s.db.Call(r.Context(), http.MethodGet, "auth/user", map[string]any{"username": in.Username})
	if err != nil {
		s.logger.Error("user creation precheck failed", "endpoint", "auth/user", "error", err.Error())
		s.dataError(w, err)
		return
	}
	if len(existing.Data.Rows) > 0 {
		fail(w, http.StatusConflict, "CONFLICT", "Ya existe un usuario con ese nombre")
		return
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, 500, "INTERNAL_ERROR", "No fue posible crear el usuario")
		return
	}
	dbUsername := identity.DatabaseUsername(in.Username, s.sqlUserPrefix)
	sqlAccount, err := s.db.Call(r.Context(), http.MethodPost, "users/sql_account", map[string]any{"db_username": dbUsername, "db_password": in.Password})
	if err != nil {
		s.logger.Error("SQL account creation failed", "endpoint", "users/sql_account", "db_username", dbUsername, "error", err.Error())
		s.dataError(w, err)
		return
	}
	if len(sqlAccount.Data.Rows) != 1 || asString(sqlAccount.Data.Rows[0]["db_username"]) != dbUsername {
		fail(w, http.StatusBadGateway, "SQL_ACCOUNT_UNVERIFIED", "No fue posible confirmar la cuenta SQL")
		return
	}
	params := map[string]any{"username": in.Username, "db_username": dbUsername, "full_name": in.FullName, "role": in.Role, "active": in.Active, "password_hash": string(hashBytes)}
	res, err := s.db.Call(r.Context(), http.MethodPost, "users", params)
	if err != nil {
		s.logger.Error("application user insert failed", "endpoint", "users", "db_username", dbUsername, "error", err.Error())
		s.dataError(w, err)
		return
	}
	for _, row := range res.Data.Rows {
		delete(row, "password_hash")
	}
	s.logger.Info("application and SQL user created", "actor_user_id", principal(r.Context()).ID, "db_username", dbUsername)
	writeJSON(w, http.StatusCreated, res.Data.Rows)
}
func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	in, ok := userInput(w, r, false)
	if !ok {
		return
	}
	params := map[string]any{"id": r.PathValue("id"), "username": in.Username, "full_name": in.FullName, "role": in.Role, "active": in.Active, "password_hash": ""}
	if in.Password != "" {
		b, hashErr := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			fail(w, 500, "INTERNAL_ERROR", "No fue posible actualizar la contraseña")
			return
		}
		params["password_hash"] = string(b)
	}
	s.callRows(w, r, http.MethodPut, "users/item", params, http.StatusOK, true)
}
func (s *Server) disableUser(w http.ResponseWriter, r *http.Request) {
	s.callNoContent(w, r, http.MethodDelete, "users/item", map[string]any{"id": r.PathValue("id")})
}

type userPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

func userInput(w http.ResponseWriter, r *http.Request, passwordRequired bool) (userPayload, bool) {
	var in userPayload
	if !decode(w, r, &in) {
		return in, false
	}
	var err error
	if in.Username, err = validation.Text(in.Username, "username", 64, true); err != nil {
		fail(w, 422, "VALIDATION_ERROR", err.Error())
		return in, false
	}
	if in.FullName, err = validation.Text(in.FullName, "full_name", 120, true); err != nil {
		fail(w, 422, "VALIDATION_ERROR", err.Error())
		return in, false
	}
	if validation.Role(in.Role) != nil {
		fail(w, 422, "VALIDATION_ERROR", "Rol no permitido")
		return in, false
	}
	if (passwordRequired && len(in.Password) < 10) || len([]byte(in.Password)) > 72 {
		fail(w, 422, "VALIDATION_ERROR", "La contraseña debe tener entre 10 y 72 bytes")
		return in, false
	}
	return in, true
}

func (s *Server) listResource(w http.ResponseWriter, r *http.Request, resource string) {
	params, ok := listParams(w, r)
	if !ok {
		return
	}
	s.callRows(w, r, http.MethodGet, resource, params, http.StatusOK, false)
}
func (s *Server) getResource(w http.ResponseWriter, r *http.Request, resource string) {
	s.callOne(w, r, http.MethodGet, resource+"/item", map[string]any{"id": r.PathValue("id")}, false)
}
func (s *Server) createResource(w http.ResponseWriter, r *http.Request, resource string) {
	params, ok := resourceInput(w, r, resource)
	if ok {
		s.callRows(w, r, http.MethodPost, resource, params, http.StatusCreated, false)
	}
}
func (s *Server) updateResource(w http.ResponseWriter, r *http.Request, resource string) {
	params, ok := resourceInput(w, r, resource)
	if ok {
		params["id"] = r.PathValue("id")
		s.callRows(w, r, http.MethodPut, resource+"/item", params, http.StatusOK, false)
	}
}
func (s *Server) deleteResource(w http.ResponseWriter, r *http.Request, resource string) {
	s.callNoContent(w, r, http.MethodDelete, resource+"/item", map[string]any{"id": r.PathValue("id")})
}

func resourceInput(w http.ResponseWriter, r *http.Request, resource string) (map[string]any, bool) {
	var p map[string]any
	if !decode(w, r, &p) {
		return nil, false
	}
	allowed := map[string]bool{"name": true, "description": true, "active": true}
	if resource == "supplements" {
		allowed = map[string]bool{"name": true, "brand": true, "description": true, "price": true, "stock": true, "presentation": true, "flavor": true, "weight": true, "image_url": true, "active": true, "category_ids": true}
	}
	for k := range p {
		if !allowed[k] {
			fail(w, 400, "UNKNOWN_FIELD", "Campo no permitido: "+k)
			return nil, false
		}
	}
	name, _ := p["name"].(string)
	clean, err := validation.Text(name, "name", 140, true)
	if err != nil {
		fail(w, 422, "VALIDATION_ERROR", err.Error())
		return nil, false
	}
	p["name"] = clean
	if resource == "supplements" {
		brand, _ := p["brand"].(string)
		if p["brand"], err = validation.Text(brand, "brand", 100, true); err != nil {
			fail(w, 422, "VALIDATION_ERROR", err.Error())
			return nil, false
		}
		price, priceOK := p["price"].(float64)
		stock, stockOK := p["stock"].(float64)
		if !priceOK || !stockOK || price < 0 || stock < 0 || math.Trunc(stock) != stock {
			fail(w, 422, "VALIDATION_ERROR", "Precio y stock deben ser mayores o iguales a cero")
			return nil, false
		}
		for _, field := range []struct {
			name string
			max  int
		}{{"description", 4000}, {"presentation", 100}, {"flavor", 100}, {"weight", 100}, {"image_url", 2048}} {
			value, _ := p[field.name].(string)
			cleaned, fieldErr := validation.Text(value, field.name, field.max, false)
			if fieldErr != nil {
				fail(w, 422, "VALIDATION_ERROR", fieldErr.Error())
				return nil, false
			}
			p[field.name] = cleaned
		}
		if validation.ImageURL(p["image_url"].(string)) != nil {
			fail(w, 422, "VALIDATION_ERROR", "La URL de imagen no es válida")
			return nil, false
		}
		for _, field := range []string{"category_ids"} {
			ids, valid := relationIDs(p[field])
			if !valid {
				fail(w, 422, "VALIDATION_ERROR", field+" contiene IDs inválidos o repetidos")
				return nil, false
			}
			p[field] = ids
		}
	} else {
		description, _ := p["description"].(string)
		cleaned, fieldErr := validation.Text(description, "description", 500, false)
		if fieldErr != nil {
			fail(w, 422, "VALIDATION_ERROR", fieldErr.Error())
			return nil, false
		}
		p["description"] = cleaned
	}
	return p, true
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	if !s.allowAttempt(s.orderAttempts, "ip:"+clientIP(r.RemoteAddr), 5) {
		fail(w, http.StatusTooManyRequests, "RATE_LIMITED", "Demasiados pedidos; espera un minuto")
		return
	}
	var in struct {
		CustomerName   string `json:"customer_name"`
		CustomerPhone  string `json:"customer_phone"`
		CustomerEmail  string `json:"customer_email"`
		IdempotencyKey string `json:"idempotency_key"`
		Items          []struct {
			SupplementID int64 `json:"supplement_id"`
			Quantity     int   `json:"quantity"`
		} `json:"items"`
	}
	if !decode(w, r, &in) {
		return
	}
	var err error
	if in.CustomerName, err = validation.Text(in.CustomerName, "nombre", 120, true); err != nil {
		fail(w, 422, "VALIDATION_ERROR", err.Error())
		return
	}
	if in.CustomerPhone, err = validation.Text(in.CustomerPhone, "teléfono", 30, true); err != nil {
		fail(w, 422, "VALIDATION_ERROR", err.Error())
		return
	}
	in.CustomerEmail = strings.TrimSpace(in.CustomerEmail)
	if len(in.CustomerEmail) > 160 {
		fail(w, 422, "VALIDATION_ERROR", "El correo electrónico es demasiado largo")
		return
	}
	if in.CustomerEmail != "" {
		address, emailErr := mail.ParseAddress(in.CustomerEmail)
		if emailErr != nil || address.Address != in.CustomerEmail {
			fail(w, 422, "VALIDATION_ERROR", "El correo electrónico no es válido")
			return
		}
	}
	if len(in.Items) == 0 || len(in.Items) > 50 {
		fail(w, 422, "VALIDATION_ERROR", "El pedido debe contener entre 1 y 50 productos")
		return
	}
	if len(in.IdempotencyKey) < 16 || len(in.IdempotencyKey) > 80 || strings.IndexFunc(in.IdempotencyKey, func(char rune) bool {
		return !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' || char == '.')
	}) >= 0 {
		fail(w, 422, "VALIDATION_ERROR", "La clave de idempotencia no es válida")
		return
	}
	seen := map[int64]bool{}
	for _, item := range in.Items {
		if item.SupplementID < 1 || item.Quantity < 1 || item.Quantity > 99 || seen[item.SupplementID] {
			fail(w, 422, "VALIDATION_ERROR", "Productos o cantidades inválidos")
			return
		}
		seen[item.SupplementID] = true
	}
	itemsJSON, _ := json.Marshal(in.Items)
	idHash := sha256.Sum256([]byte(in.IdempotencyKey))
	id := hex.EncodeToString(idHash[:16])
	orderNumber := "MAZ-" + strings.ToUpper(id[:10])
	s.callRows(w, r, http.MethodPost, "orders", map[string]any{"id": id, "order_number": orderNumber, "customer_name": in.CustomerName, "customer_phone": in.CustomerPhone, "customer_email": in.CustomerEmail, "items_json": string(itemsJSON)}, http.StatusCreated, false)
}
func (s *Server) listOrders(w http.ResponseWriter, r *http.Request) {
	params, ok := listParams(w, r)
	if !ok {
		return
	}
	s.callRows(w, r, http.MethodGet, "orders", params, 200, false)
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	s.callOne(w, r, http.MethodGet, "orders/item", map[string]any{"id": r.PathValue("id")}, false)
}
func (s *Server) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &in) {
		return
	}
	if validation.OrderStatus(in.Status) != nil {
		fail(w, 422, "VALIDATION_ERROR", "Estado no permitido")
		return
	}
	s.callRows(w, r, http.MethodPut, "orders/status", map[string]any{"id": r.PathValue("id"), "status": in.Status}, 200, false)
}

func (s *Server) backup(w http.ResponseWriter, r *http.Request) {
	task, err := s.exporter.Create(r.Context())
	if err != nil {
		s.exportError(w, err)
		return
	}
	s.logger.Info("TiDB Cloud export requested", "actor_user_id", principal(r.Context()).ID, "export_id", task.ExportID)
	writeJSON(w, http.StatusAccepted, task)
}

func (s *Server) listBackups(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.exporter.List(r.Context())
	if err != nil {
		s.exportError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) exportError(w http.ResponseWriter, err error) {
	s.logger.Error("TiDB Cloud export failed", "error", err.Error())
	if errors.Is(err, cloudexport.ErrCLIUnavailable) {
		fail(w, http.StatusServiceUnavailable, "TIDB_CLI_UNAVAILABLE", "TiDB Cloud CLI no está instalado o configurado en el servidor")
		return
	}
	fail(w, http.StatusBadGateway, "TIDB_EXPORT_ERROR", "TiDB Cloud no pudo completar la operación de exportación")
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil || cookie.Value == "" {
			fail(w, 401, "UNAUTHORIZED", "Inicia sesión para continuar")
			return
		}
		res, err := s.db.Call(r.Context(), http.MethodGet, "sessions/current", map[string]any{"token_hash": hash(cookie.Value)})
		if err != nil || len(res.Data.Rows) != 1 {
			fail(w, 401, "UNAUTHORIZED", "La sesión no es válida o expiró")
			return
		}
		row := res.Data.Rows[0]
		p := models.Principal{ID: asInt64(row["user_id"]), Username: asString(row["username"]), FullName: asString(row["full_name"]), Role: asString(row["role"])}
		csrf := asString(row["csrf_hash"])
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			supplied := r.Header.Get("X-CSRF-Token")
			if subtle.ConstantTimeCompare([]byte(hash(supplied)), []byte(csrf)) != 1 {
				fail(w, 403, "CSRF_INVALID", "Token CSRF inválido")
				return
			}
		}
		r.Header.Set("X-Session-CSRF", asString(row["csrf_token"]))
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	})
}
func (s *Server) roles(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := principal(r.Context())
			if !allowed[p.Role] {
				s.logger.Warn("authorization denied", "user_id", p.ID, "role", p.Role, "path", r.URL.Path)
				fail(w, 403, "FORBIDDEN", "No tienes permisos para realizar esta acción")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func (s *Server) publicGet(w http.ResponseWriter, r *http.Request, endpoint string) {
	id, err := positiveID(r.PathValue("id"))
	if err != nil {
		fail(w, 400, "INVALID_ID", "ID inválido")
		return
	}
	s.callOne(w, r, http.MethodGet, endpoint, map[string]any{"id": id}, false)
}
func (s *Server) callRows(w http.ResponseWriter, r *http.Request, method, endpoint string, p map[string]any, status int, strip bool) {
	res, err := s.db.Call(r.Context(), method, endpoint, p)
	if err != nil {
		s.dataError(w, err)
		return
	}
	rows := res.Data.Rows
	if strip {
		for _, row := range rows {
			delete(row, "password_hash")
		}
	}
	writeJSON(w, status, rows)
}
func (s *Server) callOne(w http.ResponseWriter, r *http.Request, method, endpoint string, p map[string]any, strip bool) {
	if _, err := positiveID(fmt.Sprint(p["id"])); err != nil && endpoint != "orders/item" {
		fail(w, 400, "INVALID_ID", "ID inválido")
		return
	}
	res, err := s.db.Call(r.Context(), method, endpoint, p)
	if err != nil {
		s.dataError(w, err)
		return
	}
	if len(res.Data.Rows) == 0 {
		fail(w, 404, "NOT_FOUND", "Registro no encontrado")
		return
	}
	if strip {
		delete(res.Data.Rows[0], "password_hash")
	}
	writeJSON(w, 200, res.Data.Rows[0])
}
func (s *Server) callNoContent(w http.ResponseWriter, r *http.Request, method, endpoint string, p map[string]any) {
	if _, err := positiveID(fmt.Sprint(p["id"])); err != nil {
		fail(w, 400, "INVALID_ID", "ID inválido")
		return
	}
	_, err := s.db.Call(r.Context(), method, endpoint, p)
	if err != nil {
		s.dataError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) dataError(w http.ResponseWriter, err error) {
	var de *tidb.Error
	if errors.As(err, &de) {
		switch de.Code {
		case 1062:
			fail(w, 409, "CONFLICT", "Ya existe un registro con esos datos")
		case 1451, 1452:
			fail(w, 409, "REFERENCE_CONFLICT", "La operación entra en conflicto con registros relacionados")
		default:
			fail(w, 502, "DATA_SERVICE_ERROR", "No fue posible completar la operación")
		}
		return
	}
	fail(w, 503, "DATA_SERVICE_UNAVAILABLE", "El servicio de datos no está disponible")
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self'; script-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.logger.Error("panic recovered", "path", r.URL.Path)
				fail(w, 500, "INTERNAL_ERROR", "Ocurrió un error inesperado")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (s *Server) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
func clientIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
func (s *Server) allowAttempt(bucket map[string][]time.Time, key string, limit int) bool {
	s.attemptMu.Lock()
	defer s.attemptMu.Unlock()
	now := time.Now()
	cut := now.Add(-time.Minute)
	a := bucket[key][:0]
	for _, t := range bucket[key] {
		if t.After(cut) {
			a = append(a, t)
		}
	}
	if len(a) >= limit {
		bucket[key] = a
		return false
	}
	bucket[key] = append(a, now)
	if len(bucket) > 10000 {
		for candidate, times := range bucket {
			if len(times) == 0 || times[len(times)-1].Before(cut) {
				delete(bucket, candidate)
			}
		}
	}
	return true
}
func relationIDs(value any) ([]int64, bool) {
	if value == nil {
		return []int64{0}, true
	}
	values, ok := value.([]any)
	if !ok {
		return nil, false
	}
	if len(values) == 0 {
		return []int64{0}, true
	}
	result := make([]int64, 0, len(values))
	seen := map[int64]bool{}
	for _, raw := range values {
		number, ok := raw.(float64)
		id := int64(number)
		if !ok || number < 1 || math.Trunc(number) != number || seen[id] {
			return nil, false
		}
		seen[id] = true
		result = append(result, id)
	}
	return result, true
}
func principal(ctx context.Context) models.Principal {
	p, _ := ctx.Value(principalKey{}).(models.Principal)
	return p
}
func listParams(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	query := r.URL.Query()
	p := map[string]any{}
	for _, field := range []string{"search", "category", "brand"} {
		value := strings.TrimSpace(query.Get(field))
		if len([]rune(value)) > 100 {
			fail(w, 400, "INVALID_FILTER", "Filtro demasiado largo: "+field)
			return nil, false
		}
		if value != "" {
			p[field] = value
		}
	}
	for _, field := range []string{"min_price", "max_price"} {
		if value := query.Get(field); value != "" {
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil || parsed < 0 || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
				fail(w, 400, "INVALID_FILTER", "Rango de precio inválido")
				return nil, false
			}
			p[field] = parsed
		}
	}
	if value := query.Get("in_stock"); value != "" {
		if value != "true" && value != "false" {
			fail(w, 400, "INVALID_FILTER", "Disponibilidad inválida")
			return nil, false
		}
		p["in_stock"] = value
	}
	if value := query.Get("status"); value != "" {
		if validation.OrderStatus(value) != nil {
			fail(w, 400, "INVALID_FILTER", "Estado inválido")
			return nil, false
		}
		p["status"] = value
	}
	page, pageSize := 1, 50
	var err error
	if value := query.Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 || page > 100000 {
			fail(w, 400, "INVALID_FILTER", "Página inválida")
			return nil, false
		}
	}
	if value := query.Get("page_size"); value != "" {
		pageSize, err = strconv.Atoi(value)
		if err != nil || pageSize < 1 || pageSize > 200 {
			fail(w, 400, "INVALID_FILTER", "Tamaño de página inválido")
			return nil, false
		}
	}
	p["page"], p["page_size"] = page, pageSize
	return p, true
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		fail(w, 400, "INVALID_JSON", "El cuerpo JSON no es válido")
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		fail(w, 400, "INVALID_JSON", "Solo se permite un objeto JSON")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, models.ErrorEnvelope{Error: models.APIError{Code: code, Message: message}})
}
func positiveID(v string) (int64, error) {
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("secure random unavailable")
	}
	return hex.EncodeToString(b)
}
func hash(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
func asInt64(v any) int64 { i, _ := strconv.ParseInt(asString(v), 10, 64); return i }
func asBool(v any) bool   { return asString(v) == "1" || strings.EqualFold(asString(v), "true") }
func number(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		f, _ := strconv.ParseFloat(asString(v), 64)
		return f
	}
}
func publicUser(row map[string]any) map[string]any {
	return map[string]any{"id": asInt64(row["id"]), "username": asString(row["username"]), "full_name": asString(row["full_name"]), "role": asString(row["role"]), "active": asBool(row["active"])}
}
