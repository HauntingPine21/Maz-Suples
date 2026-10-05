package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"maz-suplementos/internal/cloudexport"
	"maz-suplementos/internal/config"
	"maz-suplementos/internal/httpapi"
	"maz-suplementos/internal/tidb"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err.Error())
		os.Exit(1)
	}
	client := tidb.New(cfg.DataServiceBaseURL, cfg.DataAppID, cfg.DataAPIPublicKey, cfg.DataAPIPrivateKey)
	exporter, err := cloudexport.New(cfg.TiDBClusterID, cfg.TiDBDatabase, cfg.TiDBCloudProfile)
	if err != nil {
		logger.Error("export configuration error", "error", err.Error())
		os.Exit(1)
	}
	web := http.FileServer(http.Dir("web"))
	server := httpapi.New(client, exporter, web, cfg.CookieSecure, cfg.SessionTTL, logger, cfg.TiDBSQLUserPrefix)
	httpServer := &http.Server{Addr: ":" + cfg.Port, Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	logger.Info("Maz-Suplementos listening", "address", httpServer.Addr, "environment", cfg.Env)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}
