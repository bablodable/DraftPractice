package heroes

// Hero describes the hero payload returned by OpenDota's heroStats endpoint.
type Hero struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	LocalizedName string   `json:"localized_name"`
	PrimaryAttr   string   `json:"primary_attr"`
	AttackType    string   `json:"attack_type"`
	Roles         []string `json:"roles"`
	Img           string   `json:"img"`
	Icon          string   `json:"icon"`
	BaseHealth    int      `json:"base_health"`
	BaseMana      int      `json:"base_mana"`
	BaseArmor     float64  `json:"base_armor"`
	BaseAttackMin int      `json:"base_attack_min"`
	BaseAttackMax int      `json:"base_attack_max"`
	MoveSpeed     int      `json:"move_speed"`
	HeroID        int      `json:"hero_id"`
	Legs          int      `json:"legs"`
}
