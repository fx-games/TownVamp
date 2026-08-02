package common

import (
	"math/rand"
	"rpg/defs"
	"rpg/util"
)

// Factory built the dungeon crawler's content: random enemies, random loot,
// enchantments, potions, scrolls and skills. All of that generation went with
// the RPG systems. What survives is the handful of world objects the town map
// itself asks for while loading.
type Factory struct {
	definitions  defs.DataContainer
	actionBroker ActionBroker
	randomSource *rand.Rand
	state        *State
	gui          func() UserInterface

	getTimeKeeper  func() *TimeKeeper
	logError       func(msg string)
	queueAfterTurn func(func())
}

type FactoryContext struct {
	Definitions  defs.DataContainer
	ActionBroker ActionBroker
	RandomSource *rand.Rand
	State        *State

	GetTimeKeeper func() *TimeKeeper

	Msg func(msg string)
}

func NewFactory(context FactoryContext) *Factory {
	return &Factory{
		definitions:   context.Definitions,
		randomSource:  context.RandomSource,
		state:         context.State,
		actionBroker:  context.ActionBroker,
		getTimeKeeper: context.GetTimeKeeper,
		logError:      context.Msg,
	}
}

// NewSign is what the town map places for readable inscriptions.
func (g *Factory) NewSign(inscription []string) *BumpObject {
	return NewBumpObject("a sign", defs.IconSign, func(actor *Actor, startAnimating func() bool) {
		g.gui().OpenTextWindow(util.ToLabelText(inscription))
	})
}

func (g *Factory) SetQueueAfterTurn(afterTurn func(call func())) {
	g.queueAfterTurn = afterTurn
}

func (g *Factory) SetUserInterface(ui func() UserInterface) {
	g.gui = ui
}
