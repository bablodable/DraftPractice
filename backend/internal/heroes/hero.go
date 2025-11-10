package heroes

import "context"

// Hero represents a simplified Dota 2 hero descriptor used by the draft simulator.
type Hero struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PrimaryRole string   `json:"primaryRole"`
	Roles       []string `json:"roles"`
	AttackType  string   `json:"attackType"`
}

// Source provides access to an up-to-date hero catalog.
type Source interface {
	List(ctx context.Context) ([]Hero, error)
}
