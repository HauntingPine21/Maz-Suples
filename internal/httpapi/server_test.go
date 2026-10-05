package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"maz-suplementos/internal/cloudexport"
	"maz-suplementos/internal/tidb"
)

type fakeDB struct {
	role         string
	calls        []string
	params       []map[string]any
	passwordHash string
	userAbsent   bool
}

type fakeExporter struct {
	tasks []cloudexport.Task
	err   error
}

func (f *fakeExporter) Create(context.Context) (cloudexport.Task, error) {
	if f.err != nil {
		return cloudexport.Task{}, f.err
	}
	if len(f.tasks) == 0 {
		return cloudexport.Task{ExportID: "exp-test", State: "PENDING"}, nil
	}
	return f.tasks[0], nil
}
func (f *fakeExporter) List(context.Context) ([]cloudexport.Task, error) { return f.tasks, f.err }

func (f *fakeDB) Call(_ context.Context, method, path string, p map[string]any) (tidb.Response, error) {
	f.calls = append(f.calls, method+" "+path)
	f.params = append(f.params, p)
	rows := []map[string]any{}
	switch path {
	case "sessions/current":
		rows = []map[string]any{{"user_id": "7", "username": "ana", "full_name": "Ana Ruiz", "role": f.role, "csrf_hash": hash("csrf-test")}}
	case "auth/user":
		if !f.userAbsent {
			rows = []map[string]any{{"id": "7", "username": "ana", "full_name": "Ana Ruiz", "role": f.role, "active": "1", "password_hash": f.passwordHash}}
		}
	case "users/sql_account":
		rows = []map[string]any{{"db_username": p["db_username"], "db_host": "%"}}
	case "users":
		rows = []map[string]any{{"id": "8", "username": p["username"], "db_username": p["db_username"], "full_name": p["full_name"], "role": p["role"], "active": p["active"]}}
	case "sessions":
		rows = []map[string]any{{"user_id": "7"}}
	default:
		rows = []map[string]any{{"id": "1", "name": "Creatina", "active": "1"}}
	}
	return tidb.Response{Data: tidb.Data{Rows: rows, Result: tidb.Result{Code: 200, RowCount: len(rows)}}}, nil
}
func request(t *testing.T, s *Server, method, path, body, role string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "csrf-test")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}
func TestRoleMatrixDeniesWrites(t *testing.T) {
	tests := []struct{ role, method, path, body string }{{"AUDITOR", "DELETE", "/api/supplements/1", ""}, {"CAPTURISTA", "POST", "/api/users", `{"username":"x"}`}, {"ADMINISTRADOR", "POST", "/api/supplements", `{"name":"x"}`}}
	for _, tc := range tests {
		t.Run(tc.role+tc.path, func(t *testing.T) {
			s := New(&fakeDB{role: tc.role}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
			rec := request(t, s, tc.method, tc.path, tc.body, tc.role)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
func TestCSRFRequiredForAuthenticatedMutation(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("DELETE", "/api/supplements/1", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}
func TestCapturerCanCreateCatalogData(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := request(t, s, "POST", "/api/categories", `{"name":"Proteína","description":"","active":true}`, "CAPTURISTA")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCapturerCanCreateSupplementWithoutGoalsOrIngredients(t *testing.T) {
	db := &fakeDB{role: "CAPTURISTA"}
	s := New(db, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"name":"Creatina","brand":"Maz","description":"Producto de prueba","price":449,"stock":10,"presentation":"Bote","flavor":"","weight":"300 g","image_url":"","active":true,"category_ids":[2]}`
	rec := request(t, s, http.MethodPost, "/api/supplements", body, "CAPTURISTA")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	params := db.params[len(db.params)-1]
	for field, want := range map[string][]int64{"category_ids": {2}} {
		if got, ok := params[field].([]int64); !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s=%#v, want %#v", field, params[field], want)
		}
	}
	for _, removed := range []string{"goal_ids", "ingredient_ids"} {
		if _, exists := params[removed]; exists {
			t.Fatalf("removed field %s was forwarded", removed)
		}
	}
}

func TestSupplementRejectsRemovedRelationFields(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"name":"Creatina","brand":"Maz","description":"","price":100,"stock":1,"presentation":"Bote","flavor":"","weight":"300 g","image_url":"","active":true,"category_ids":[1],"goal_ids":[1]}`
	rec := request(t, s, http.MethodPost, "/api/supplements", body, "CAPTURISTA")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRemovedResourcesReturnNotFound(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, path := range []string{"/api/goals", "/api/ingredients"} {
		rec := request(t, s, http.MethodGet, path, "", "CAPTURISTA")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}
func TestLoginUsesBcryptAndSetsSecureCookieAttributes(t *testing.T) {
	b, _ := bcrypt.GenerateFromPassword([]byte("correct horse battery"), bcrypt.MinCost)
	db := &fakeDB{role: "AUDITOR", passwordHash: string(b)}
	s := New(db, &fakeExporter{}, http.NotFoundHandler(), true, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"ana","password":"correct horse battery"}`))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) < 2 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookies: %#v", cookies)
	}
}

func TestAdminCreatesApplicationAndSQLUser(t *testing.T) {
	db := &fakeDB{role: "ADMINISTRADOR", userAbsent: true}
	s := New(db, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)), "cluster.")
	body := `{"username":"capturista_1","password":"correct horse battery","full_name":"Capturista Uno","role":"CAPTURISTA","active":true}`
	rec := request(t, s, http.MethodPost, "/api/users", body, "ADMINISTRADOR")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	wantCalls := []string{"GET sessions/current", "GET auth/user", "POST users/sql_account", "POST users"}
	if !reflect.DeepEqual(db.calls, wantCalls) {
		t.Fatalf("calls=%#v", db.calls)
	}
	if db.params[2]["db_password"] != "correct horse battery" {
		t.Fatal("SQL password was not forwarded transiently")
	}
	if db.params[2]["db_username"] != "cluster.capturista_1" || db.params[3]["db_username"] != "cluster.capturista_1" {
		t.Fatalf("TiDB SQL prefix was not persisted consistently: %#v %#v", db.params[2], db.params[3])
	}
	if !strings.HasPrefix(asString(db.params[3]["password_hash"]), "$2") {
		t.Fatal("application password was not bcrypt hashed")
	}
	if _, exists := db.params[3]["db_password"]; exists {
		t.Fatal("plaintext password reached users table endpoint")
	}
	if strings.Contains(rec.Body.String(), "correct horse battery") || strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatal("password leaked in response")
	}
}
func TestOrderValidationRejectsDuplicateItems(t *testing.T) {
	s := New(&fakeDB{}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"customer_name":"Luis","customer_phone":"6141234567","idempotency_key":"12345678-1234-1234-1234-123456789abc","items":[{"supplement_id":1,"quantity":1},{"supplement_id":1,"quantity":2}]}`
	rec := request(t, s, "POST", "/api/orders", body, "")
	if rec.Code != 422 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSupplementValidationRejectsFractionalStock(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"name":"Creatina","brand":"Maz","description":"","price":100,"stock":1.5,"presentation":"Bote","flavor":"","weight":"300 g","image_url":"","active":true,"category_ids":[1]}`
	rec := request(t, s, "POST", "/api/supplements", body, "CAPTURISTA")
	if rec.Code != 422 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSupplementValidationRejectsDuplicateCategories(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"name":"Creatina","brand":"Maz","description":"","price":100,"stock":1,"presentation":"Bote","flavor":"","weight":"300 g","image_url":"","active":true,"category_ids":[1,1]}`
	rec := request(t, s, http.MethodPost, "/api/supplements", body, "CAPTURISTA")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
func TestPublicCatalogDoesNotRequireSession(t *testing.T) {
	s := New(&fakeDB{}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/catalog", nil))
	if rec.Code != 200 {
		t.Fatalf("status=%d", rec.Code)
	}
	var rows []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil || len(rows) != 1 {
		t.Fatalf("body=%s err=%v", rec.Body.String(), err)
	}
}

func TestBackupCreatesRealCloudExportTask(t *testing.T) {
	db := &fakeDB{role: "ADMINISTRADOR"}
	exporter := &fakeExporter{tasks: []cloudexport.Task{{ExportID: "exp-real123", State: "PENDING", FileType: "SQL", TargetType: "LOCAL"}}}
	s := New(db, exporter, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := request(t, s, "POST", "/api/backups", `{}`, "ADMINISTRADOR")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"export_id":"exp-real123"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestSecurityHeadersOnFrontend(t *testing.T) {
	s := New(&fakeDB{}, &fakeExporter{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("Content-Security-Policy") == "" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("security headers missing: %#v", rec.Header())
	}
}

func TestListFiltersAreBounded(t *testing.T) {
	s := New(&fakeDB{}, &fakeExporter{}, http.NotFoundHandler(), false, time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/catalog?page_size=10000", nil))
	if rec.Code != 400 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
