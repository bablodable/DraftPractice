package draft

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/example/draftpractice/internal/heroes"
)

// Store хранит сессии в памяти.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*DraftSession
}

// NewStore создаёт новый Store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*DraftSession),
	}
}

// CreateSession создаёт новую сессию и запускает таймер.
func (s *Store) CreateSession(ctx context.Context, radiantName, direName string) (*DraftSession, error) {
	if len(heroes.All()) == 0 {
		return nil, errors.New("hero cache is empty")
	}

	id := generateID()
	session := newDraftSession(id, radiantName, direName)

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	fmt.Printf("[SESSION] New draft %s started: %s vs %s\n", id, radiantName, direName)
	fmt.Printf("[SESSION] First move: %s %s (timer %d sec)\n", session.Side, session.Stage, session.CurrentTimer)

	// Запускаем фонового тикера для этой сессии.
	go s.runTimer(session)

	return session.ClonePtr(), nil
}

// runTimer — отслеживает время хода и делает автоход при истечении.
func (s *Store) runTimer(session *DraftSession) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		if session.Completed {
			s.mu.Unlock()
			fmt.Printf("[SESSION] Draft %s completed.\n", session.ID)
			return
		}

		session.CurrentTimer--

		if session.CurrentTimer%5 == 0 || session.CurrentTimer < 5 {
			fmt.Printf("[TIMER] %s %s: %d sec left (reserve R:%ds / D:%ds)\n",
				session.Side, session.Stage,
				session.CurrentTimer,
				session.ReserveRadiant, session.ReserveDire)
		}

		if session.CurrentTimer <= 0 {
			var reserve *int
			if session.Side == SideRadiant {
				reserve = &session.ReserveRadiant
			} else {
				reserve = &session.ReserveDire
			}

			if *reserve > 0 {
				*reserve--
				session.CurrentTimer = 1 // дышим секунду и продолжаем
			} else {
				// резерв закончился — автопик или автобан
				autoHero := randomAvailableHero(session)
				fmt.Printf("[AUTO] %s auto-%s hero %d (no time left)\n",
					session.Side, session.Stage, autoHero)

				_ = session.ApplyAction(autoHero)
				fmt.Printf("[NEXT] Now %s %s (timer %d sec)\n",
					session.Side, session.Stage, session.CurrentTimer)
			}
		}

		s.mu.Unlock()
	}
}

func randomAvailableHero(s *DraftSession) int {
	for _, h := range heroes.All() {
		if !s.IsHeroUsed(h.ID) {
			return h.ID
		}
	}
	return 1 // fallback
}

// GetSession возвращает состояние сессии.
func (s *Store) GetSession(id string) (*DraftSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	clone := session.Clone()
	return &clone, nil
}

// ApplyAction — применяет действие и двигает сессию.
func (s *Store) ApplyAction(id string, actionType Phase, heroID int) (*DraftSession, error) {
	if actionType != PhaseBan && actionType != PhasePick {
		return nil, fmt.Errorf("unsupported action type %q", actionType)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}

	if session.Completed {
		return nil, errors.New("draft is already completed")
	}

	if session.Stage != actionType {
		return nil, fmt.Errorf("expected %s action but got %s", session.Stage, actionType)
	}

	if err := session.ApplyAction(heroID); err != nil {
		return nil, err
	}

	fmt.Printf("[ACTION] %s %s hero %d\n", session.Side, session.Stage, heroID)
	fmt.Printf("[NEXT] Now %s %s (timer %d sec)\n",
		session.Side, session.Stage, session.CurrentTimer)

	return session.ClonePtr(), nil
}

// ClonePtr — удобный способ вернуть указатель на клон.
func (s *DraftSession) ClonePtr() *DraftSession {
	clone := s.Clone()
	return &clone
}

func generateID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprint(time.Now().UnixNano())))
	}
	return hex.EncodeToString(buf)
}
