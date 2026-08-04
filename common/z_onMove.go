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
		// No field of view is ever read except the player's — perception queries
		// raycast on demand — so an NPC step used to pay for a vision map nobody
		// looked at. With the town walking in real time that was the single
		// biggest cost in the move path.
		return actions
	}

	// player only
	transitionAtLocation, hasTransition := currentMap.GetTransitionAt(newPos)
	if hasTransition {
		actions = append(actions, NewTransitionAction(action.Cause(), transitionAtLocation))
	}

	actions = append(actions, NewUpdateFoVAction(action.Cause(), actor))
	if pickUp, hasSomethingToPickUp := d.autoPickUpAt(context, newPos); hasSomethingToPickUp {
		actions = append(actions, pickUp)
	}
	d.logObjectsAtLocation(context, newPos)
	return actions
}

// autoPickUpAt is what the player picking things up by walking over them means
// now that walking is continuous: it hangs off arriving in a cell rather than
// off ending a turn, because there are no turns left to end.
func (d OnMoveHook) autoPickUpAt(context ActionContext, pos geometry.Point) (Action, bool) {
	if !d.autoPickup {
		return nil, false
	}
	state := context.GameState
	player := state.GetPlayer()
	if !state.GetMap().IsItemAt(pos) || !player.CanPickup() {
		return nil, false
	}
	if state.GetMap().ItemAt(pos).IsDroppedByPlayer() {
		return nil, false
	}
	return NewPickUpAction(AbstractCause("AutoPickup"), player), true
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
