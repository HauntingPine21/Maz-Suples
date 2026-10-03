package models

import "time"

const (
	RoleAdmin    = "ADMINISTRADOR"
	RoleCapturer = "CAPTURISTA"
	RoleAuditor  = "AUDITOR"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	DBUsername   string    `json:"db_username"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type Principal struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}
type APIError struct{ Code, Message string }
