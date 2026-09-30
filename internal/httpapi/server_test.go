package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"maz-suplementos/internal/tidb"
)

type fakeDB struct {
	role         string
	calls        []string
	passwordHash string
}

func (f *fakeDB) Call(_ context.Context, method, path string, p map[string]any) (tidb.Response, error) {
	f.calls = append(f.calls, method+" "+path)
	rows := []map[string]any{}
	switch path {
	case "sessions/current":
		rows = []map[string]any{{"user_id": "7", "username": "ana", "full_name": "Ana Ruiz", "role": f.role, "csrf_hash": hash("csrf-test")}}
	case "auth/user":
		rows = []map[string]any{{"id": "7", "username": "ana", "full_name": "Ana Ruiz", "role": f.role, "active": "1", "password_hash": f.passwordHash}}
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
			s := New(&fakeDB{role: tc.role}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
			rec := request(t, s, tc.method, tc.path, tc.body, tc.role)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
func TestCSRFRequiredForAuthenticatedMutation(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("DELETE", "/api/supplements/1", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}
func TestCapturerCanCreateCatalogData(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := request(t, s, "POST", "/api/categories", `{"name":"Proteína","description":"","active":true}`, "CAPTURISTA")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
func TestLoginUsesBcryptAndSetsSecureCookieAttributes(t *testing.T) {
	b, _ := bcrypt.GenerateFromPassword([]byte("correct horse battery"), bcrypt.MinCost)
	db := &fakeDB{role: "AUDITOR", passwordHash: string(b)}
	s := New(db, http.NotFoundHandler(), true, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
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
func TestOrderValidationRejectsDuplicateItems(t *testing.T) {
	s := New(&fakeDB{}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"customer_name":"Luis","customer_phone":"6141234567","idempotency_key":"12345678-1234-1234-1234-123456789abc","items":[{"supplement_id":1,"quantity":1},{"supplement_id":1,"quantity":2}]}`
	rec := request(t, s, "POST", "/api/orders", body, "")
	if rec.Code != 422 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSupplementValidationRejectsFractionalStockAndDuplicateRelations(t *testing.T) {
	s := New(&fakeDB{role: "CAPTURISTA"}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"name":"Creatina","brand":"Maz","description":"","price":100,"stock":1.5,"presentation":"Bote","flavor":"","weight":"300 g","image_url":"","active":true,"category_ids":[1,1],"goal_ids":[],"ingredient_ids":[]}`
	rec := request(t, s, "POST", "/api/supplements", body, "CAPTURISTA")
	if rec.Code != 422 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
func TestPublicCatalogDoesNotRequireSession(t *testing.T) {
	s := New(&fakeDB{}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
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

func TestBackupWritesPrivateLogicalFile(t *testing.T) {
	dir := t.TempDir()
	s := New(&fakeDB{role: "ADMINISTRADOR"}, http.NotFoundHandler(), false, time.Hour, dir, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := request(t, s, "POST", "/api/backups", `{}`, "ADMINISTRADOR")
	if rec.Code != 201 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("files=%v err=%v", files, err)
	}
	info, err := os.Stat(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("backup permissions too broad: %o", info.Mode().Perm())
	}
}

func TestSecurityHeadersOnFrontend(t *testing.T) {
	s := New(&fakeDB{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("Content-Security-Policy") == "" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("security headers missing: %#v", rec.Header())
	}
}

func TestListFiltersAreBounded(t *testing.T) {
	s := New(&fakeDB{}, http.NotFoundHandler(), false, time.Hour, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/catalog?page_size=10000", nil))
	if rec.Code != 400 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
