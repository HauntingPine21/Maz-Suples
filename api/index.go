package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	mazapp "maz-suplementos/app"
)

var (
	initOnce sync.Once
	app      http.Handler
	initErr  error
)

// Handler is the Vercel Go Function entry point. The application handler is
// initialized once per warm function instance and reused for later requests.
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(func() {
		app, initErr = mazapp.NewHandler()
	})

	if initErr != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "CONFIGURATION_ERROR",
				"message": "El servicio no está configurado correctamente",
			},
		})
		return
	}

	app.ServeHTTP(w, r)
}
