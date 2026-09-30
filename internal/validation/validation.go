package validation

import (
	"errors"
	"net/url"
	"strings"
)

var Roles = map[string]bool{"ADMINISTRADOR": true, "CAPTURISTA": true, "AUDITOR": true}
var OrderStatuses = map[string]bool{"PENDIENTE": true, "CONFIRMADO": true, "PREPARANDO": true, "COMPLETADO": true, "CANCELADO": true}

func Text(value, field string, max int, required bool) (string, error) {
	v := strings.TrimSpace(value)
	if required && v == "" {
		return "", errors.New(field + " es obligatorio")
	}
	if len([]rune(v)) > max {
		return "", errors.New(field + " es demasiado largo")
	}
	return v, nil
}

func ImageURL(value string) error {
	if value == "" {
		return nil
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("image_url debe ser una URL http(s) válida")
	}
	return nil
}

func Role(value string) error {
	if !Roles[value] {
		return errors.New("rol no permitido")
	}
	return nil
}
func OrderStatus(value string) error {
	if !OrderStatuses[value] {
		return errors.New("estado de pedido no permitido")
	}
	return nil
}
