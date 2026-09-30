package tidb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientChecksAuthAndInnerResultCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "public" || p != "private" {
			t.Errorf("missing basic auth")
		}
		if r.URL.Path != "/api/v1beta/app/app-1/endpoint/catalog" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"sql_endpoint","data":{"columns":[],"rows":[],"result":{"code":1146,"message":"table not found"}}}`))
	}))
	defer server.Close()
	c := New(server.URL, "app-1", "public", "private")
	_, err := c.Call(context.Background(), http.MethodGet, "catalog", map[string]any{"page": 1})
	dataErr, ok := err.(*Error)
	if !ok || dataErr.Code != 1146 {
		t.Fatalf("expected inner Data Service error, got %v", err)
	}
}
func TestClientAcceptsSuccessfulRows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"type":"sql_endpoint","data":{"columns":[],"rows":[{"id":"1"}],"result":{"code":200,"row_count":1}}}`))
	}))
	defer server.Close()
	res, err := New(server.URL, "app", "a", "b").Call(context.Background(), http.MethodGet, "catalog", nil)
	if err != nil || len(res.Data.Rows) != 1 {
		t.Fatalf("res=%v err=%v", res, err)
	}
}

func TestClientEncodesBooleanParametersAsStrings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["active"] != "true" {
			t.Fatalf("active=%#v, want string true", body["active"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"sql_endpoint","data":{"columns":[],"rows":[],"result":{"code":200}}}`))
	}))
	defer server.Close()

	_, err := New(server.URL, "app", "a", "b").Call(context.Background(), http.MethodPost, "users", map[string]any{"active": true})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}
