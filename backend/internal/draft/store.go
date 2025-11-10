package draft

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"math/rand"
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
func (s *Store) CreateSession(
	ctx context.Context,
	radiantName, direName string,
	firstPick Side,
	botSide Side,
	botSpeed string,
) (*DraftSession, error) {
	if len(heroes.All()) == 0 {
		return nil, errors.New("hero cache is empty")
	}

	id := generateID()
	session := newDraftSession(id, radiantName, direName, firstPick)
	session.BotSide = botSide
	session.BotSpeed = botSpeed

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	fmt.Printf("[SESSION] New draft %s started: %s vs %s\n", id, radiantName, direName)
	fmt.Printf("[SESSION] Bot: %s (%s)\n", botSide, botSpeed)
	fmt.Printf("[SESSION] First move: %s %s (timer %d sec)\n",
		session.Side, session.Stage, session.CurrentTimer)

	// Запускаем фонового тикера для этой сессии.
	go s.runTimer(session)

	// Если первый пик принадлежит боту — он начинает сам
	if session.FirstPick == session.BotSide {
		go func() {
			delay := botThinkDelay(botSpeed)
			fmt.Printf("[BOT] starting draft: %s (%s) thinking for %v...\n",
				session.BotSide, botSpeed, delay)
			time.Sleep(delay)

			s.mu.Lock()
			defer s.mu.Unlock()

			if session.Completed {
				return
			}

			hero := randomAvailableHero(session)
			_ = session.ApplyAction(hero)
			fmt.Printf("[BOT] %s starts with hero %d\n", session.BotSide, hero)
			fmt.Printf("[NEXT] Now %s %s (timer %d sec)\n",
				session.Side, session.Stage, session.CurrentTimer)
		}()
	}

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

// ApplyAction — применяет действие игрока и двигает сессию.
// Если следующий ход принадлежит боту, запускает ThinkAndAct.
func (s *Store) ApplyAction(id string, actionType Phase, heroID int) (*DraftSession, error) {
	if actionType != PhaseBan && actionType != PhasePick {
		return nil, fmt.Errorf("unsupported action type %q", actionType)
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
		return nil, fmt.Errorf("expected %s action but got %s", session.Stage, actionType)
	}

	if err := session.ApplyAction(heroID); err != nil {
		s.mu.Unlock()
		return nil, err
	}

	fmt.Printf("[ACTION] %s %s hero %d\n", session.Side, session.Stage, heroID)
	fmt.Printf("[NEXT] Now %s %s (timer %d sec)\n",
		session.Side, session.Stage, session.CurrentTimer)

	nextSide := session.Side
	nextPhase := session.Stage
	botSpeed := session.BotSpeed
	sessionCopy := session.Clone()
	s.mu.Unlock()

	// Проверяем, ход ли теперь бота
	if session.BotSide == nextSide {
		go func() {
			delay := botThinkDelay(botSpeed)
			fmt.Printf("[BOT] %s bot (%s) thinking for %v...\n", nextSide, botSpeed, delay)
			time.Sleep(delay)

			s.mu.Lock()
			defer s.mu.Unlock()

			sess, ok := s.sessions[id]
			if !ok || sess.Completed {
				return
			}

			hero := randomAvailableHero(sess)
			_ = sess.ApplyAction(hero)
			fmt.Printf("[BOT] %s auto-%s hero %d after %v\n", nextSide, nextPhase, hero, delay)
			fmt.Printf("[NEXT] Now %s %s (timer %d sec)\n",
				sess.Side, sess.Stage, sess.CurrentTimer)
		}()
	}

	return &sessionCopy, nil
}

func randomAvailableHero(s *DraftSession) int {
	for _, h := range heroes.All() {
		if !s.IsHeroUsed(h.ID) {
			return h.ID
		}
	}
	return 1 // fallback
}

// botThinkDelay возвращает задержку в зависимости от скорости бота.
func botThinkDelay(speed string) time.Duration {
	switch speed {
	case "fast":
		return time.Duration(1+rand.Intn(3)) * time.Second // 1–3 сек
	case "slow":
		return time.Duration(7+rand.Intn(6)) * time.Second // 7–12 сек
	default:
		return time.Duration(3+rand.Intn(5)) * time.Second // 3–7 сек
	}
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

// ClonePtr — удобный способ вернуть указатель на клон.
func (s *DraftSession) ClonePtr() *DraftSession {
	clone := s.Clone()
	return &clone
}

func generateID() string {
	buf := make([]byte, 8)
	if _, err := crand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprint(time.Now().UnixNano())))
	}
	return hex.EncodeToString(buf)
}
