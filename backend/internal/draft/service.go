package draft

import (
	"context"
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
	mu         sync.RWMutex
	sessions   map[string]*Session
	heroSource heroes.Source
}

// NewService constructs a Service with an empty session registry.
func NewService(heroSource heroes.Source) *Service {
	return &Service{
		sessions:   make(map[string]*Session),
		heroSource: heroSource,
	}
}

// CreateSession instantiates a new draft session with default parameters.
func (s *Service) CreateSession(ctx context.Context, radiantName, direName string) (*Session, error) {
	pool, err := s.buildHeroPool(ctx)
	if err != nil {
		return nil, err
	}

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
		AvailablePool: pool,
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session, nil
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

func (s *Service) buildHeroPool(ctx context.Context) ([]string, error) {
	if s.heroSource == nil {
		return nil, errors.New("hero source is not configured")
	}

	list, err := s.heroSource.List(ctx)
	if err != nil {
		return nil, err
	}

	pool := make([]string, 0, len(list))
	for _, hero := range list {
		pool = append(pool, hero.ID)
	}
	return pool, nil
}

func generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
