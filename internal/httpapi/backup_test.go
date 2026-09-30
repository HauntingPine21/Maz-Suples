package httpapi

import (
	"strings"
	"testing"
	"time"
)

func TestEncodeSQLBackupEscapesValues(t *testing.T) {
	backup := string(encodeSQLBackup(
		time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC),
		[]string{"users"},
		map[string][]map[string]any{
			"users": {{"id": 1, "full_name": "O'Brien\\Admin", "active": true, "note": nil}},
		},
	))

	for _, expected := range []string{"'O''Brien\\\\Admin'", "TRUE", "NULL"} {
		if !strings.Contains(backup, expected) {
			t.Fatalf("backup missing escaped value %q", expected)
		}
	}
}
