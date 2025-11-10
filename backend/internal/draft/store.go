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

// Store хранит активные сессии и ссылку на используемого бота.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*DraftSession
	bot      Bot
}

// NewStore создаёт хранилище драфтов с базовым ботом.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*DraftSession),
		bot:      RandomBot{}, // пока просто случайный выбор
	}
}

// CreateSession — создаёт новую сессию.
func (s *Store) CreateSession(ctx context.Context, radiantName, direName string) (*DraftSession, error) {
	if len(heroes.All()) == 0 {
		return nil, errors.New("hero cache is empty")
	}

	id := generateID()
	session := newDraftSession(id, radiantName, direName)

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	clone := session.Clone()
	return &clone, nil
}

// GetSession — возвращает сессию по ID.
func (s *Store) GetSession(id string) (*DraftSession, error) {
	s.mu.RLock()
	session, ok := s.sessions[id]
	s.mu.RUnlock()

	if !ok {
		return nil, errors.New("session not found")
	}

	clone := session.Clone()
	return &clone, nil
}

// ApplyAction — применяет действие игрока и, если нужно, вызывает бота.
func (s *Store) ApplyAction(id string, actionType Phase, heroID int) (*DraftSession, error) {
	if actionType != PhaseBan && actionType != PhasePick {
		return nil, fmt.Errorf("unsupported action type %q", actionType)
	}

	if heroID <= 0 {
		return nil, errors.New("hero id must be positive")
	}

	if !heroExists(heroID) {
		return nil, fmt.Errorf("unknown hero id %d", heroID)
	}

	s.mu.Lock()
	session, ok := s.sessions[id]
	if !ok {
		s.mu.Unlock()
		return nil, errors.New("session not found")
	}

	if session.Completed {
		s.mu.Unlock()
		return nil, errors.New("draft is already completed")
	}

	if session.Stage != actionType {
		s.mu.Unlock()
		return nil, fmt.Errorf("expected %s action but received %s", session.Stage, actionType)
	}

	if err := session.ApplyAction(heroID); err != nil {
		s.mu.Unlock()
		return nil, err
	}

	// Когда очередь Dire — бот делает ход
	if session.Side == "dire" && !session.Completed {
		botHero := s.bot.ChooseHero(session)
		if botHero > 0 {
			_ = session.ApplyAction(botHero)
		}
	}

	clone := session.Clone()
	s.mu.Unlock()
	return &clone, nil
}

func heroExists(id int) bool {
	heroesList := heroes.All()
	for _, hero := range heroesList {
		if hero.ID == id || hero.HeroID == id {
			return true
		}
	}
	return false
}

func generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
