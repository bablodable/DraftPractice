package draft

// Side определяет сторону: Radiant или Dire.
type Side string

const (
	SideRadiant Side = "radiant"
	SideDire    Side = "dire"
)

// Turn описывает один ход драфта.
type Turn struct {
	Phase Phase // "ban" или "pick"
	Side  Side  // кто делает ход
	Timer int   // сколько секунд даётся на действие
}

// schedule возвращает порядок ходов Captains Mode по патчу 7.34 с реальными таймингами.
// Если Radiant имеет первый пик — используется оригинальный порядок, иначе зеркалится.
func schedule(firstPick Side) []Turn {
	fp, sp := firstPick, opposite(firstPick)

	return []Turn{
		// ---- BAN PHASE 1 (7 банов, по 15 сек) ----
		{Phase: PhaseBan, Side: fp, Timer: 15},
		{Phase: PhaseBan, Side: sp, Timer: 15},
		{Phase: PhaseBan, Side: fp, Timer: 15},
		{Phase: PhaseBan, Side: sp, Timer: 15},
		{Phase: PhaseBan, Side: sp, Timer: 15},
		{Phase: PhaseBan, Side: fp, Timer: 15},
		{Phase: PhaseBan, Side: sp, Timer: 15},

		// ---- PICK PHASE 1 (2 пика, по 30 сек) ----
		{Phase: PhasePick, Side: fp, Timer: 30},
		{Phase: PhasePick, Side: sp, Timer: 30},

		// ---- BAN PHASE 2 (3 бана, по 25 сек) ----
		{Phase: PhaseBan, Side: fp, Timer: 25},
		{Phase: PhaseBan, Side: fp, Timer: 25},
		{Phase: PhaseBan, Side: sp, Timer: 25},

		// ---- PICK PHASE 2 (6 пиков, по 35 сек) ----
		{Phase: PhasePick, Side: sp, Timer: 35},
		{Phase: PhasePick, Side: sp, Timer: 35},
		{Phase: PhasePick, Side: fp, Timer: 35},
		{Phase: PhasePick, Side: fp, Timer: 35},
		{Phase: PhasePick, Side: sp, Timer: 35},
		{Phase: PhasePick, Side: fp, Timer: 35},

		// ---- BAN PHASE 3 (4 бана, по 30 сек) ----
		{Phase: PhaseBan, Side: sp, Timer: 30},
		{Phase: PhaseBan, Side: sp, Timer: 30},
		{Phase: PhaseBan, Side: fp, Timer: 30},
		{Phase: PhaseBan, Side: fp, Timer: 30},

		// ---- PICK PHASE 3 (2 пика, по 40 сек) ----
		{Phase: PhasePick, Side: fp, Timer: 40},
		{Phase: PhasePick, Side: sp, Timer: 40},
	}
}

// opposite возвращает противоположную сторону.
func opposite(s Side) Side {
	if s == SideRadiant {
		return SideDire
	}
	return SideRadiant
}

// Reserve Time — общий запас на все ходы команды.
const ReserveTimeSeconds = 130
