package common

import (
	"fmt"

	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
)

type OnMoveHook struct {
	BaseActionHook
	autoPickup bool
}

func (d OnMoveHook) IsBefore() bool {
	return false
}

func (d OnMoveHook) IsAfter() bool {
	return true
}

func (d OnMoveHook) Name() string {
	return "OnMoveHook"
}

func NewMainOnMoveHook(autoPickup bool) *OnMoveHook {
	return &OnMoveHook{
		BaseActionHook: BaseActionHook{
			hookedAction: ActionNameMove,
		},
		autoPickup: autoPickup,
	}
}

type Mover interface {
	GetActor() *Actor
	GetNewPos() geometry.Point
}

func (d OnMoveHook) After(action Action, context ActionContext) []Action {
	var actions []Action
	state := context.GameState
	moveAction := action.(Mover)
	actor := moveAction.GetActor()
	newPos := moveAction.GetNewPos()
	currentMap := state.GetMap()

	actions = append(actions, TriggerCellEffectsDuringMovementForActor(state, action, actor, newPos, AnimateNow())...)

	if actor != state.GetPlayer() {
		actions = append(actions, NewUpdateFoVAction(action.Cause(), actor))
		return actions
	}

	// player only
	transitionAtLocation, hasTransition := currentMap.GetTransitionAt(newPos)
	if hasTransition {
		actions = append(actions, NewTransitionAction(action.Cause(), transitionAtLocation))
	}

	actions = append(actions, []Action{
		NewUpdateFoVAction(action.Cause(), actor),
	}...)
	d.logObjectsAtLocation(context, newPos)
	return actions
}

func (d OnMoveHook) logObjectsAtLocation(context ActionContext, pos geometry.Point) *Item {
	state := context.GameState
	if state.GetMap().IsItemAt(pos) {
		item := state.GetMap().ItemAt(pos)
		context.Msg(fmt.Sprintf("You see [:darkText]%s", item.Name()))
		return item
	}
	return nil
}

// TriggerCellEffectsDuringMovementForActor fires whatever the tile an actor just
// stepped onto does to them. It lived in the knockback action, which went with
// the combat systems, but movement still needs it.
func TriggerCellEffectsDuringMovementForActor(state *State, sourceAction Action, actor *Actor, mapPos geometry.Point, startAnimating func() bool) []Action {
	currentMap := state.GetMap()

	if currentMap.IsObjectAt(mapPos) {
		currentMap.ObjectAt(mapPos).ActorWalkedOn(actor, startAnimating)
	}

	if currentMap.GetCell(mapPos).TileType.Special == gridmap.SpecialTileDamaging {
		action := NewChangeResourceAction(sourceAction.Cause(), actor, stats.Health, -3)
		action.SetAnimationStartPredicate(startAnimating)
		return OneAction(action)
	}
	return nil
}
