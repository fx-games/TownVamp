package common

import (
	"rpg/geometry"
	"rpg/util"
)

type MoveAction struct {
	BaseAction
	actor                     *Actor
	targetPos                 geometry.Point
	originPos                 geometry.Point
	isPlayer                  bool
	animationDelay            float64
	isMovementVisibleToPlayer bool
	animationDisabled         bool
	keepDrawPosition          bool
}

// KeepDrawPosition leaves the actor where he is drawn. The free-walking player
// owns his own position between cells, so his move action only hands the grid
// its new occupant — it must neither snap him to the cell nor animate him into
// it.
func (d *MoveAction) KeepDrawPosition() {
	d.keepDrawPosition = true
}

func (d *MoveAction) IsMovementVisibleToPlayer() bool {
	return d.isMovementVisibleToPlayer
}

func NewMoveAction(source Cause, actor *Actor, targetPos geometry.Point, isPlayer bool) *MoveAction {
	return &MoveAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameMove,
		},
		actor:          actor,
		targetPos:      targetPos,
		isPlayer:       isPlayer,
		animationDelay: 0.2,
	}
}

func (d *MoveAction) SetAnimationDelay(delay float64) {
	d.animationDelay = delay
}

// PROBLEM: Since all enemy actions are triggered in one tick,
// but the re-positioning of the actor takes place over multiple ticks,
// all the following actors will have a false idea of the world state.
// SOLUTION: The actor should be moved immediately, but the animation should be delayed.

func (d *MoveAction) Execute(context ActionContext) []Action {
	state := context.GameState
	gameMap := state.GetMap() // read state
	actor := d.actor
	d.originPos = actor.MapPosition()
	gameMap.MoveActor(actor, d.targetPos) // write state

	d.isMovementVisibleToPlayer = false
	if d.isPlayer {
		d.isMovementVisibleToPlayer = true
	} else {
		player := state.GetPlayer()
		d.isMovementVisibleToPlayer = player.CanSee(d.GetOldPos()) || player.CanSee(d.GetNewPos())
	}
	if d.isNotAnimated() && !d.keepDrawPosition {
		actor.SetDrawPosition(d.targetPos.ToPointF())
	}
	//println(fmt.Sprintf("Actor %s moved to %s", actor.Action(), d.targetPos))
	return nil
}

func (d *MoveAction) GetAnimations() []Animation {
	if d.isNotAnimated() || d.keepDrawPosition {
		return nil
	}
	ticksNeeded := util.SecondsToTicks(d.animationDelay)
	return OneAnimation(&MovementAnimation{
		actor:       d.actor,
		path:        []geometry.Point{d.originPos, d.targetPos},
		ticksNeeded: ticksNeeded,
		isForPlayer: d.isPlayer,
	})
	//return &DummyAnimation{time: 5, cancelRunningAnimations: true}
}

func (d *MoveAction) isNotAnimated() bool {
	return !d.isMovementVisibleToPlayer || d.animationDisabled
}

func (d *MoveAction) GetActor() *Actor {
	return d.actor
}

func (d *MoveAction) GetOldPos() geometry.Point {
	return d.originPos
}

func (d *MoveAction) GetNewPos() geometry.Point {
	return d.targetPos
}

func (d *MoveAction) DisableAnimation() {
	d.animationDisabled = true
}

// PlayerMoveStep resolves a step in one direction into whatever it means where
// the player is standing: opening a door, bumping an object, talking to a vendor,
// or moving. Bumping someone no longer attacks them.
func PlayerMoveStep(actor *Actor, relativeDirection geometry.Point, context ActionContext) Action {
	state := context.GameState
	gameMap := state.GetMap()
	targetPos := actor.MapPosition().Add(relativeDirection)
	isObjectAt := gameMap.IsObjectAt(targetPos)

	if isObjectAt {
		objectAtPos := gameMap.ObjectAt(targetPos)
		if state.IsClosedDoorAt(targetPos) {
			return NewOpenDoorAction(NewActorCause("Open door", actor), actor, targetPos)
		} else {
			objectAtPos.ActorBumpedInto(actor, AnimateNow())
		}
	}

	isActorAt := gameMap.IsActorAt(targetPos)
	isTargetFree := !isActorAt && gameMap.IsWalkable(targetPos)
	if isActorAt {
		actorAtPos := gameMap.ActorAt(targetPos)
		if actorAtPos.IsVendor() {
			return NewTalkToAction(NewActorCause("Talk to", actor), actorAtPos)
		}
		return nil
	} else if isTargetFree {
		return NewMoveAction(NewActorCause("Move", actor), actor, targetPos, state.GetPlayer() == actor)
	}

	state.GetMap().SetExplored(targetPos)
	return nil
}

func AnimateNow() func() bool {
	return func() bool {
		return true
	}
}
