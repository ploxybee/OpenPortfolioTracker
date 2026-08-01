package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ploxybee/OpenPortfolioTracker/internal/portfolio"
)

type Server struct{ provider portfolio.Provider }

// NewServer creates an API server backed by the supplied portfolio provider.
func NewServer(provider portfolio.Provider) *Server { return &Server{provider: provider} }

// Routes registers the public HTTP endpoints.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/portfolio", s.portfolio)
	mux.HandleFunc("POST /api/v1/portfolio/refresh", s.refresh)
	return cors(mux)
}

// health confirms that the API is running.
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// portfolio returns the latest portfolio snapshot.
func (s *Server) portfolio(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.provider.Snapshot(r.Context())
	if err != nil {
		log.Printf("portfolio request failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Unable to fetch your portfolio. Check the broker connection and account permissions."})
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// refresh bypasses the hourly cache and fetches a new portfolio snapshot.
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	refresher, ok := s.provider.(interface {
		Refresh(context.Context) (portfolio.Snapshot, error)
	})
	if !ok {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "Portfolio refresh is not configured."})
		return
	}
	snapshot, err := refresher.Refresh(r.Context())
	if err != nil {
		log.Printf("portfolio refresh failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Unable to refresh your portfolio. Check the broker connection and account permissions."})
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// writeJSON sends a JSON response with the supplied status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// cors allows the local dashboard to call this API.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
