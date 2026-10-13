package game

import (
	"rpg/util"
)

// debugMenu lists the switches that exist for playtesting, not for play.
func (g *GameInstance) debugMenu() []util.MenuItem {
	return []util.MenuItem{
		{
			MainText: "Daywalker: " + onOff(g.daywalker) + " -> " + onOff(!g.daywalker),
			Action:   func() { g.SetDaywalker(!g.daywalker) },
		},
	}
}

// SetDaywalker makes the vampire unseen and unkillable while on: nobody
// witnesses him, trespass goes unnoticed, Hunters cannot find or catch him, and
// his Blood neither drains nor burns. It is for walking the town by day to
// watch the schedules play out, not a mode of play.
func (g *GameInstance) SetDaywalker(enabled bool) {
	g.daywalker = enabled
	g.PlayerMessage("Daywalker " + onOff(enabled) + ".")
}

func onOff(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}
