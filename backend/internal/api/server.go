package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ploxybee/OpenPortfolioTracker/internal/portfolio"
)

type Server struct{ provider portfolio.Provider }

func NewServer(provider portfolio.Provider) *Server { return &Server{provider: provider} }
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/portfolio", s.portfolio)
	return cors(mux)
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) portfolio(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.provider.Snapshot(r.Context())
	if err != nil {
		log.Printf("portfolio request failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Unable to fetch your Tiger portfolio. Check the API credentials and account permissions."})
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
