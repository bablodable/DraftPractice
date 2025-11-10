package draft

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/example/draftpractice/internal/heroes"
)

// Bot — интерфейс для любого ИИ-драфтера (рандомный, эвристический, ML и т.д.)
type Bot interface {
	ChooseHero(*DraftSession) int
}

// RandomBot — базовый бот, выбирающий случайного доступного героя.
type RandomBot struct{}

// ChooseHero выбирает случайного героя, которого нет в банах или пиках.
func (b RandomBot) ChooseHero(session *DraftSession) int {
	all := heroes.All()
	if len(all) == 0 {
		return 0
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		h := all[rand.Intn(len(all))].ID
		if !session.IsHeroUsed(h) {
			fmt.Printf("[BOT] Dire auto-%s hero %d\n", session.Stage, h)
			return h
		}
	}
	return 0
}
