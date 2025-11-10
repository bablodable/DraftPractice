package draft

import "fmt"

// Phase — категория текущего действия в драфте.
type Phase string

const (
	PhaseBan  Phase = "ban"
	PhasePick Phase = "pick"
)

// Team — состав команды (баны и пики).
type Team struct {
	Name  string `json:"name"`
	Bans  []int  `json:"bans"`
	Picks []int  `json:"picks"`
}

// DraftSession — состояние одной сессии драфта.
type DraftSession struct {
	ID        string
	Radiant   Team
	Dire      Team
	Stage     Phase
	Side      Side
	Completed bool
	Step      int
	Order     []Turn

	// Таймеры и резервы
	CurrentTimer   int
	ReserveRadiant int
	ReserveDire    int
	taken          map[int]struct{}

	// Кто получил первый пик
	FirstPick Side `json:"firstPick"`
	// Скорость бота
	BotSpeed string `json:"botSpeed"`
	// Какая сторона управляется ботом (radiant или dire)
	BotSide Side `json:"botSide"`
}

// newDraftSession — инициализация новой сессии.
func newDraftSession(id, radiantName, direName string, firstPick Side) *DraftSession {
	s := &DraftSession{
		ID:             id,
		Radiant:        Team{Name: radiantName},
		Dire:           Team{Name: direName},
		Order:          schedule(firstPick),
		Step:           0,
		ReserveRadiant: ReserveTimeSeconds,
		ReserveDire:    ReserveTimeSeconds,
		taken:          make(map[int]struct{}),
		FirstPick:      firstPick,
	}

	s.Stage = s.Order[0].Phase
	s.Side = s.Order[0].Side
	s.CurrentTimer = s.Order[0].Timer
	return s
}

// ApplyAction — записывает героя для текущей стороны и двигает драфт вперёд.
func (s *DraftSession) ApplyAction(heroID int) error {
	if s.Completed {
		return fmt.Errorf("draft already completed")
	}

	if heroID <= 0 {
		return fmt.Errorf("invalid hero id %d", heroID)
	}
	if _, exists := s.taken[heroID]; exists {
		return fmt.Errorf("hero %d already selected", heroID)
	}

	team := s.activeTeam()
	switch s.Stage {
	case PhaseBan:
		team.Bans = append(team.Bans, heroID)
	case PhasePick:
		team.Picks = append(team.Picks, heroID)
	default:
		return fmt.Errorf("unknown stage %q", s.Stage)
	}

	s.taken[heroID] = struct{}{}
	s.NextStep()
	return nil
}

// NextStep — переход к следующему действию.
func (s *DraftSession) NextStep() {
	s.Step++
	if s.Step >= len(s.Order) {
		s.Completed = true
		return
	}
	next := s.Order[s.Step]
	s.Stage = next.Phase
	s.Side = next.Side
	s.CurrentTimer = next.Timer
}

// activeTeam — возвращает команду, которая сейчас ходит.
func (s *DraftSession) activeTeam() *Team {
	if s.Side == SideRadiant {
		return &s.Radiant
	}
	return &s.Dire
}

// IsHeroUsed — проверяет, использовался ли герой.
func (s *DraftSession) IsHeroUsed(heroID int) bool {
	for _, h := range s.Radiant.Bans {
		if h == heroID {
			return true
		}
	}
	for _, h := range s.Dire.Bans {
		if h == heroID {
			return true
		}
	}
	for _, h := range s.Radiant.Picks {
		if h == heroID {
			return true
		}
	}
	for _, h := range s.Dire.Picks {
		if h == heroID {
			return true
		}
	}
	return false
}

// Clone — делает глубокую копию сессии.
func (s *DraftSession) Clone() DraftSession {
	copySession := DraftSession{
		ID:             s.ID,
		Stage:          s.Stage,
		Side:           s.Side,
		Completed:      s.Completed,
		Step:           s.Step,
		Order:          append([]Turn(nil), s.Order...),
		CurrentTimer:   s.CurrentTimer,
		ReserveRadiant: s.ReserveRadiant,
		ReserveDire:    s.ReserveDire,
		FirstPick:      s.FirstPick,
		taken:          make(map[int]struct{}, len(s.taken)),
	}
	for heroID := range s.taken {
		copySession.taken[heroID] = struct{}{}
	}
	copySession.Radiant = Team{
		Name:  s.Radiant.Name,
		Bans:  append([]int(nil), s.Radiant.Bans...),
		Picks: append([]int(nil), s.Radiant.Picks...),
	}
	copySession.Dire = Team{
		Name:  s.Dire.Name,
		Bans:  append([]int(nil), s.Dire.Bans...),
		Picks: append([]int(nil), s.Dire.Picks...),
	}
	return copySession
}
