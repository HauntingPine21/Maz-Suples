package database

import _ "embed"

// SchemaSQL contains the database schema included in downloadable backups.
//
//go:embed schema.sql
var SchemaSQL string
