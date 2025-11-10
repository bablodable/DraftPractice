package draft

import "fmt"

// Phase represents the current action category within the draft.
type Phase string

const (
	PhaseBan  Phase = "ban"
	PhasePick Phase = "pick"
)

// Side identifies which team is expected to act.
type Side string

const (
	SideRadiant Side = "radiant"
	SideDire    Side = "dire"
)

// DraftTeam keeps track of bans and picks for one side of the draft.
type DraftTeam struct {
	Name  string `json:"name"`
	Bans  []int  `json:"bans"`
	Picks []int  `json:"picks"`
}

// DraftSession stores the state of an ongoing or finished Captains Mode draft.
type DraftSession struct {
	ID        string    `json:"id"`
	Stage     Phase     `json:"stage"`
	Side      Side      `json:"side"`
	Radiant   DraftTeam `json:"radiant"`
	Dire      DraftTeam `json:"dire"`
	Completed bool      `json:"completed"`

	stageIndex int
	taken      map[int]struct{}
}

// TODO: replace this prototype schedule with the full Captains Mode sequence.
var turnSchedule = []Phase{
	PhaseBan,
	PhaseBan,
	PhaseBan,
	PhaseBan,
	PhasePick,
	PhasePick,
	PhasePick,
	PhasePick,
}

func newDraftSession(id, radiantName, direName string) *DraftSession {
	return &DraftSession{
		ID:    id,
		Stage: turnSchedule[0],
		Side:  SideRadiant,
		Radiant: DraftTeam{
			Name:  radiantName,
			Bans:  make([]int, 0, 4),
			Picks: make([]int, 0, 4),
		},
		Dire: DraftTeam{
			Name:  direName,
			Bans:  make([]int, 0, 4),
			Picks: make([]int, 0, 4),
		},
		taken: make(map[int]struct{}),
	}
}

// ApplyAction records the hero for the active team and advances the session.
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
	s.advance()
	return nil
}

// Clone creates a deep copy that can be safely shared with callers.
func (s *DraftSession) Clone() DraftSession {
	copySession := DraftSession{
		ID:         s.ID,
		Stage:      s.Stage,
		Side:       s.Side,
		Completed:  s.Completed,
		stageIndex: s.stageIndex,
		taken:      make(map[int]struct{}, len(s.taken)),
	}

	for heroID := range s.taken {
		copySession.taken[heroID] = struct{}{}
	}

	copySession.Radiant = DraftTeam{
		Name:  s.Radiant.Name,
		Bans:  append([]int(nil), s.Radiant.Bans...),
		Picks: append([]int(nil), s.Radiant.Picks...),
	}

	copySession.Dire = DraftTeam{
		Name:  s.Dire.Name,
		Bans:  append([]int(nil), s.Dire.Bans...),
		Picks: append([]int(nil), s.Dire.Picks...),
	}

	return copySession
}

func (s *DraftSession) activeTeam() *DraftTeam {
	if s.Side == SideRadiant {
		return &s.Radiant
	}
	return &s.Dire
}

func (s *DraftSession) advance() {
	s.stageIndex++
	if s.Side == SideRadiant {
		s.Side = SideDire
	} else {
		s.Side = SideRadiant
	}

	if s.stageIndex >= len(turnSchedule) {
		s.Stage = ""
		s.Completed = true
		return
	}

	s.Stage = turnSchedule[s.stageIndex]
}

// IsHeroUsed проверяет, использовался ли герой в драфте
// (в банах или пиках любой команды).
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
