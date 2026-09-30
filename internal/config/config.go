package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env, Port                           string
	DataServiceBaseURL, DataAppID       string
	DataAPIPublicKey, DataAPIPrivateKey string
	CookieSecure                        bool
	SessionTTL                          time.Duration
}

func Load() (Config, error) {
	loadDotEnv(".env")
	ttl, err := strconv.Atoi(value("SESSION_TTL_HOURS", "12"))
	if err != nil || ttl < 1 || ttl > 168 {
		return Config{}, errors.New("SESSION_TTL_HOURS debe estar entre 1 y 168")
	}
	c := Config{
		Env: value("APP_ENV", "development"), Port: value("APP_PORT", "8080"),
		DataServiceBaseURL: strings.TrimRight(os.Getenv("TIDB_DATA_SERVICE_BASE_URL"), "/"),
		DataAppID:          os.Getenv("TIDB_DATA_APP_ID"), DataAPIPublicKey: os.Getenv("TIDB_DATA_API_PUBLIC_KEY"),
		DataAPIPrivateKey: os.Getenv("TIDB_DATA_API_PRIVATE_KEY"),
		CookieSecure:      strings.EqualFold(value("COOKIE_SECURE", "false"), "true"), SessionTTL: time.Duration(ttl) * time.Hour,
	}
	if c.DataServiceBaseURL == "" || c.DataAppID == "" || c.DataAPIPublicKey == "" || c.DataAPIPrivateKey == "" {
		return c, errors.New("faltan variables TIDB_DATA_SERVICE_*; consulta .env.example")
	}
	u, err := url.Parse(c.DataServiceBaseURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return c, fmt.Errorf("TIDB_DATA_SERVICE_BASE_URL debe ser una URL HTTPS regional sin credenciales ni query")
	}
	if strings.EqualFold(c.Env, "production") && !c.CookieSecure {
		return c, errors.New("COOKIE_SECURE debe ser true cuando APP_ENV=production")
	}
	return c, nil
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), "\"'")
		if k != "" && os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func value(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
