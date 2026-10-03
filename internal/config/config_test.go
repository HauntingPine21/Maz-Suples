package config

import "testing"

func setExportEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TIDB_CLUSTER_ID", "123456")
	t.Setenv("TIDB_DATABASE", "maz_suplementos")
	t.Setenv("TIDB_SQL_USER_PREFIX", "cluster.")
}

func TestLoadRejectsInvalidSQLUserPrefix(t *testing.T) {
	setExportEnv(t)
	t.Setenv("TIDB_SQL_USER_PREFIX", "invalid prefix")
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "https://us-east-1.data.tidbcloud.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SQL user prefix to be rejected")
	}
}

func TestLoadRequiresHTTPSDataServiceURL(t *testing.T) {
	setExportEnv(t)
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "http://example.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err == nil {
		t.Fatal("expected insecure URL to be rejected")
	}
}

func TestLoadAcceptsRegionalHTTPSURL(t *testing.T) {
	setExportEnv(t)
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "https://us-east-1.data.tidbcloud.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestProductionRequiresSecureCookie(t *testing.T) {
	setExportEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "https://us-east-1.data.tidbcloud.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err == nil {
		t.Fatal("expected insecure production cookie to be rejected")
	}
}
