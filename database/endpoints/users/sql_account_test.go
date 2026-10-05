package users

import (
	"os"
	"strings"
	"testing"
)

func TestSQLAccountUsesQuotedDynamicStatement(t *testing.T) {
	sql, err := os.ReadFile("post-users-sql-account.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := string(sql)
	if strings.Contains(text, "CREATE USER IF NOT EXISTS ${db_username}") {
		t.Fatal("Data Service cannot bind a placeholder in the SQL account-name position")
	}
	for _, required := range []string{"QUOTE(${db_username})", "QUOTE(${db_password})", "PREPARE maz_create_user", "EXECUTE maz_create_user", "DEALLOCATE PREPARE maz_create_user"} {
		if !strings.Contains(text, required) {
			t.Fatalf("SQL account endpoint is missing %q", required)
		}
	}
}
