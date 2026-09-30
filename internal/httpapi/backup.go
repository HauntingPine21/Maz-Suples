package httpapi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"maz-suplementos/database"
)

func encodeSQLBackup(createdAt time.Time, tables []string, rowsByTable map[string][]map[string]any) []byte {
	var sql strings.Builder
	fmt.Fprintf(&sql, "-- Maz-Suplementos logical backup\n-- Created at: %s\n\n", createdAt.UTC().Format(time.RFC3339))
	sql.WriteString(database.SchemaSQL)
	if !strings.HasSuffix(database.SchemaSQL, "\n") {
		sql.WriteByte('\n')
	}
	sql.WriteString("\nSET FOREIGN_KEY_CHECKS=0;\nSTART TRANSACTION;\n\n")

	for _, table := range tables {
		rows := rowsByTable[table]
		if len(rows) == 0 {
			fmt.Fprintf(&sql, "-- `%s`: no rows\n\n", table)
			continue
		}
		for _, row := range rows {
			columns := make([]string, 0, len(row))
			for column := range row {
				columns = append(columns, column)
			}
			sort.Strings(columns)
			values := make([]string, 0, len(columns))
			quotedColumns := make([]string, 0, len(columns))
			for _, column := range columns {
				quotedColumns = append(quotedColumns, quoteIdentifier(column))
				values = append(values, sqlValue(row[column]))
			}
			fmt.Fprintf(&sql, "INSERT INTO %s (%s) VALUES (%s);\n", quoteIdentifier(table), strings.Join(quotedColumns, ", "), strings.Join(values, ", "))
		}
		sql.WriteByte('\n')
	}

	sql.WriteString("COMMIT;\nSET FOREIGN_KEY_CHECKS=1;\n")
	return []byte(sql.String())
}

func quoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

func sqlValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "NULL"
	case bool:
		if typed {
			return "TRUE"
		}
		return "FALSE"
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'g', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		text := fmt.Sprint(value)
		escaped := strings.NewReplacer(
			"\\", "\\\\",
			"'", "''",
			"\x00", "\\0",
			"\n", "\\n",
			"\r", "\\r",
			"\x1a", "\\Z",
		).Replace(text)
		return "'" + escaped + "'"
	}
}
