package common

import (
	"rpg/audio"
	"rpg/defs"
)

type UIContext struct {
    // AdvanceWorld moves the simulation on by some number of in-game seconds.
    // Called once per fixed update; the world does not wait for the player.
    AdvanceWorld func(inGameSeconds float64)

    // TogglePause stops the clock so the player can think. WaitUntil runs the
    // simulation forward without drawing until a named moment arrives.
    TogglePause func() bool
    WaitUntil   func(condition WaitCondition)

    // The vampire's verbs. Killing is the only way to refill Blood; hiding the
    // body is what turns a killing into a disappearance.
    Kill      func() bool
    HideBody  func() bool
    LeaveTown func()
    Blood     func() *Blood

    GameState    func() *State
    PlayerAction func() PlayerActions
    AudioPlayer  *audio.Player
    ResetGame    func(class defs.LoadoutDefinition)
    Animator     func() Animator
    DebugActions func() DebugActions
    GetLog       func() []string
    StartInfos   StartInfo
    Definitions  defs.DataContainer
}

func SortStacksForTextUITyped(stacks []InventoryStack, getLetterForStack func(stack InventoryStack) rune, selected func(invItemKey rune)) []*TextualStack {
	items := make([]*TextualStack, 0)
	SortInventory(stacks)
	for _, itemStack := range stacks {
		items = append(items, NewTextualStack(itemStack, getLetterForStack(itemStack), selected))
	}
	return items
}
