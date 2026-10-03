package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"maz-suplementos/internal/config"
	"maz-suplementos/internal/identity"
	"maz-suplementos/internal/tidb"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	username := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_USERNAME"))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	fullName := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_FULL_NAME"))
	if username == "" || fullName == "" || len(password) < 10 {
		fatal(fmt.Errorf("define BOOTSTRAP_ADMIN_USERNAME, BOOTSTRAP_ADMIN_FULL_NAME y una BOOTSTRAP_ADMIN_PASSWORD de al menos 10 caracteres"))
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fatal(err)
	}
	c := tidb.New(cfg.DataServiceBaseURL, cfg.DataAppID, cfg.DataAPIPublicKey, cfg.DataAPIPrivateKey)
	dbUsername := identity.DatabaseUsername(username, cfg.TiDBSQLUserPrefix)
	_, err = c.Call(context.Background(), http.MethodPost, "users/sql_account", map[string]any{"db_username": dbUsername, "db_password": password})
	if err != nil {
		fatal(err)
	}
	_, err = c.Call(context.Background(), http.MethodPost, "users/bootstrap", map[string]any{"username": username, "db_username": dbUsername, "full_name": fullName, "password_hash": string(hash)})
	if err != nil {
		fatal(err)
	}
	fmt.Println("Administrador inicial creado o ya existente; la contraseña no fue registrada.")
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
