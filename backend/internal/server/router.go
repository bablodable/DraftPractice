package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/heroes"
)

type RouterConfig struct {
	DraftStore *draft.Store
}

func NewHandler(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// ---- Healthcheck ----
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// ---- Герои ----
	mux.HandleFunc("/api/heroes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, heroes.All())
	})

	// ---- Создание новой сессии ----
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		var req struct {
			Radiant   string `json:"radiant"`
			Dire      string `json:"dire"`
			FirstPick string `json:"firstPick"`
			BotSide   string `json:"botSide"`
			BotSpeed  string `json:"botSpeed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}

		if req.Radiant == "" || req.Dire == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "radiant and dire names are required"})
			return
		}

		// кто делает первый пик
		firstPick := draft.SideRadiant
		switch strings.ToLower(req.FirstPick) {
		case "dire":
			firstPick = draft.SideDire
		case "radiant":
			firstPick = draft.SideRadiant
		default:
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 1 {
				firstPick = draft.SideDire
			}
		}

		// скорость бота
		botSpeed := strings.ToLower(req.BotSpeed)
		switch botSpeed {
		case "fast", "slow":
		default:
			botSpeed = "medium"
		}

		// какая сторона бот
		botSide := draft.SideDire
		switch strings.ToLower(req.BotSide) {
		case "radiant":
			botSide = draft.SideRadiant
		case "dire":
			botSide = draft.SideDire
		}

		// создаём сессию
		session, err := cfg.DraftStore.CreateSession(
			r.Context(),
			req.Radiant,
			req.Dire,
			firstPick,
			botSide,
			botSpeed,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		fmt.Printf("[DEBUG] Created session: %s (bot=%s, speed=%s, firstPick=%s)\n",
			session.ID, botSide, botSpeed, firstPick)

		writeJSON(w, http.StatusCreated, session)
	})

	// ---- Получение сессии, экшены и WebSocket ----
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/stream") {
			streamHandler(cfg.DraftStore).ServeHTTP(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session id"})
			return
		}

		parts := strings.Split(path, "/")
		id := parts[0]

		// GET /api/sessions/{id}
		if len(parts) == 1 && r.Method == http.MethodGet {
			session, err := cfg.DraftStore.GetSession(id)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, session)
			return
		}

		// POST /api/sessions/{id}/action
		if len(parts) == 2 && parts[1] == "action" && r.Method == http.MethodPost {
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

		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown endpoint"})
	})

	return mux
}

// ---- JSON writer ----
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ---- WebSocket ----
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func streamHandler(store *draft.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
		path = strings.TrimSuffix(path, "/stream")
		id := strings.TrimSpace(path)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "failed to upgrade websocket", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		fmt.Printf("[WS] New connection for session %s\n", id)

		for {
			session, err := store.GetSession(id)
			if err != nil {
				conn.WriteJSON(map[string]any{
					"event": "error",
					"data":  "session not found",
				})
				return
			}

			payload := map[string]any{
				"event": "tick",
				"data": map[string]any{
					"stage":          session.Stage,
					"side":           session.Side,
					"currentTimer":   session.CurrentTimer,
					"reserveRadiant": session.ReserveRadiant,
					"reserveDire":    session.ReserveDire,
					"completed":      session.Completed,
					"radiant":        session.Radiant,
					"dire":           session.Dire,
				},
			}

			conn.WriteJSON(payload)

			if session.Completed {
				conn.WriteJSON(map[string]any{
					"event": "complete",
					"data":  "draft finished",
				})
				return
			}

			time.Sleep(time.Second)
		}
	}
}
