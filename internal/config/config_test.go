package config

import "testing"

func TestLoadRequiresHTTPSDataServiceURL(t *testing.T) {
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "http://example.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err == nil {
		t.Fatal("expected insecure URL to be rejected")
	}
}

func TestLoadAcceptsRegionalHTTPSURL(t *testing.T) {
	t.Setenv("TIDB_DATA_SERVICE_BASE_URL", "https://us-east-1.data.tidbcloud.com")
	t.Setenv("TIDB_DATA_APP_ID", "app")
	t.Setenv("TIDB_DATA_API_PUBLIC_KEY", "public")
	t.Setenv("TIDB_DATA_API_PRIVATE_KEY", "private")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestProductionRequiresSecureCookie(t *testing.T) {
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
