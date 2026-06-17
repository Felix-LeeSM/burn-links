package httpapi

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Felix-LeeSM/burn-links/internal/secrets"
)

type Server struct {
	db                    *sql.DB
	secrets               *secrets.Store
	payloadInlineMaxBytes int64
	allowedOrigin         string
}

type Options struct {
	PayloadInlineMaxBytes int64
	AllowedOrigin         string
}

func NewRouter(db *sql.DB, secretStore *secrets.Store, opts Options) http.Handler {
	payloadInlineMaxBytes := opts.PayloadInlineMaxBytes
	if payloadInlineMaxBytes <= 0 {
		payloadInlineMaxBytes = 1048576
	}

	server := Server{
		db:                    db,
		secrets:               secretStore,
		payloadInlineMaxBytes: payloadInlineMaxBytes,
		allowedOrigin:         strings.TrimRight(opts.AllowedOrigin, "/"),
	}

	r := chi.NewRouter()
	r.Use(server.cors)
	r.Get("/healthz", server.healthz)
	r.Get("/readyz", server.readyz)
	r.Get("/metrics", server.metrics)
	r.Post("/api/secrets", server.createSecret)
	r.Get("/api/secrets/{id}", server.getSecret)
	r.Post("/api/secrets/{id}/consume", server.consumeSecret)
	return r
}

func (s Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if s.allowedOrigin != "" && origin == s.allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s Server) readyz(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "database is not configured")
		return
	}
	if err := s.db.PingContext(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "database is not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("# HELP burnlink_api_info BurnLink API process info\n# TYPE burnlink_api_info gauge\nburnlink_api_info 1\n"))
}
