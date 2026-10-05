package app

import "testing"

func TestNewHandlerLoadsServerlessConfiguration(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("SESSION_TTL_HOURS", "12")
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "https://us-east-1.data.tidbcloud.com")
	t.Setenv("TIDB_DATA_APP_ID", "test-app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	t.Setenv("TIDB_CLUSTER_ID", "123456")
	t.Setenv("TIDB_DATABASE", "maz_suplementos")
	t.Setenv("TIDB_SQL_USER_PREFIX", "cluster.")

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
}
