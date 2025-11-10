package server

import (
	"encoding/json"
	"net/http"

	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/heroes"
)

// RouterConfig bundles dependencies required for building the HTTP router.
type RouterConfig struct {
	DraftService *draft.Service
}

// NewHandler constructs an http.Handler that exposes the public API for the simulator.
func NewHandler(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/heroes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		writeJSON(w, http.StatusOK, heroes.All())
	})

	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		var req struct {
			Radiant string `json:"radiant"`
			Dire    string `json:"dire"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}

		if req.Radiant == "" || req.Dire == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "radiant and dire names are required"})
			return
		}

		session, err := cfg.DraftService.CreateSession(r.Context(), req.Radiant, req.Dire)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, session)
	})

	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		id := r.URL.Path[len("/api/sessions/"):]
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session id"})
			return
		}

		session, err := cfg.DraftService.GetSession(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, session)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
