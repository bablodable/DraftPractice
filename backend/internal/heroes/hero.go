package heroes

// Hero represents a simplified Dota 2 hero descriptor used by the draft simulator.
type Hero struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PrimaryRole string `json:"primaryRole"`
}

// List provides a curated subset of heroes that can be expanded with real data later on.
func List() []Hero {
	return []Hero{
		{ID: "antimage", Name: "Anti-Mage", PrimaryRole: "Carry"},
		{ID: "axe", Name: "Axe", PrimaryRole: "Offlane"},
		{ID: "bane", Name: "Bane", PrimaryRole: "Support"},
		{ID: "crystal_maiden", Name: "Crystal Maiden", PrimaryRole: "Support"},
		{ID: "dragon_knight", Name: "Dragon Knight", PrimaryRole: "Mid"},
		{ID: "juggernaut", Name: "Juggernaut", PrimaryRole: "Carry"},
		{ID: "lina", Name: "Lina", PrimaryRole: "Mid"},
		{ID: "lion", Name: "Lion", PrimaryRole: "Support"},
		{ID: "phantom_assassin", Name: "Phantom Assassin", PrimaryRole: "Carry"},
		{ID: "puck", Name: "Puck", PrimaryRole: "Mid"},
		{ID: "shadow_shaman", Name: "Shadow Shaman", PrimaryRole: "Support"},
		{ID: "sven", Name: "Sven", PrimaryRole: "Carry"},
	}
}
