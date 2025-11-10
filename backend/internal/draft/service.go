package draft

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/example/draftpractice/internal/heroes"
)

// Service coordinates draft sessions in-memory. It will be replaced by a persistent
// implementation once storage is wired in.
type Service struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewService constructs a Service with an empty session registry.
func NewService() *Service {
	return &Service{
		sessions: make(map[string]*Session),
	}
}

// CreateSession instantiates a new draft session with default parameters.
func (s *Service) CreateSession(radiantName, direName string) *Session {
	id := generateID()
	now := time.Now().UTC()

	session := &Session{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Radiant: Team{
			Name:  radiantName,
			Bans:  make([]string, 0, 6),
			Picks: make([]string, 0, 5),
		},
		Dire: Team{
			Name:  direName,
			Bans:  make([]string, 0, 6),
			Picks: make([]string, 0, 5),
		},
		CurrentStage:  0,
		IsCompleted:   false,
		LastAction:    "",
		NextActor:     "radiant",
		AvailablePool: buildHeroPool(),
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session
}

// GetSession retrieves a session by identifier.
func (s *Service) GetSession(id string) (*Session, error) {
	s.mu.RLock()
	session, ok := s.sessions[id]
	s.mu.RUnlock()

	if !ok {
		return nil, errors.New("session not found")
	}

	return session, nil
}

func buildHeroPool() []string {
	list := heroes.List()
	pool := make([]string, 0, len(list))
	for _, hero := range list {
		pool = append(pool, hero.ID)
	}
	return pool
}

func generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
