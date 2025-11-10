package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/heroes"
)

// RouterConfig bundles dependencies required for building the HTTP router.
type RouterConfig struct {
	DraftStore *draft.Store
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

		session, err := cfg.DraftStore.CreateSession(r.Context(), req.Radiant, req.Dire)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, session)
	})

	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		trimmed := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
		if trimmed == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session id"})
			return
		}

		if strings.HasSuffix(trimmed, "/") {
			trimmed = strings.TrimSuffix(trimmed, "/")
		}

		parts := strings.Split(trimmed, "/")
		id := parts[0]
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session id"})
			return
		}

		if len(parts) == 1 {
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}

			session, err := cfg.DraftStore.GetSession(id)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}

			writeJSON(w, http.StatusOK, session)
			return
		}

		if len(parts) == 2 && parts[1] == "action" {
			if r.Method != http.MethodPost {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}

			var req struct {
				Type   string `json:"type"`
				HeroID int    `json:"heroId"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
				return
			}

			actionType := draft.Phase(req.Type)
			session, err := cfg.DraftStore.ApplyAction(id, actionType, req.HeroID)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}

			writeJSON(w, http.StatusOK, session)
			return
		}

		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown resource"})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
