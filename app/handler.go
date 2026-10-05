// Package app assembles the Maz-Suplementos HTTP application for hosted
// runtimes that need an http.Handler instead of a listening server.
package app

import (
	"log/slog"
	"net/http"
	"os"

	"maz-suplementos/internal/cloudexport"
	"maz-suplementos/internal/config"
	"maz-suplementos/internal/httpapi"
	"maz-suplementos/internal/tidb"
)

func NewHandler() (http.Handler, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := tidb.New(cfg.DataServiceBaseURL, cfg.DataAppID, cfg.DataAPIPublicKey, cfg.DataAPIPrivateKey)
	exporter, err := cloudexport.NewWithCredentials(
		cfg.TiDBClusterID,
		cfg.TiDBDatabase,
		cfg.TiDBCloudProfile,
		cfg.TiDBCloudAPIPublicKey,
		cfg.TiDBCloudAPIPrivateKey,
	)
	if err != nil {
		return nil, err
	}
	server := httpapi.New(client, exporter, http.NotFoundHandler(), cfg.CookieSecure, cfg.SessionTTL, logger, cfg.TiDBSQLUserPrefix)
	return server.Handler(), nil
}
