package draft

import "time"

// Team represents one side of the draft.
type Team struct {
	Name  string   `json:"name"`
	Bans  []string `json:"bans"`
	Picks []string `json:"picks"`
}

// Session stores the state of an ongoing or finished Captains Mode draft.
type Session struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Radiant       Team      `json:"radiant"`
	Dire          Team      `json:"dire"`
	CurrentStage  int       `json:"currentStage"`
	IsCompleted   bool      `json:"isCompleted"`
	LastAction    string    `json:"lastAction"`
	NextActor     string    `json:"nextActor"`
	AvailablePool []string  `json:"availablePool"`
}
