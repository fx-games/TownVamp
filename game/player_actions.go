package game

import (
	"fmt"
	"rpg/audio"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/util"
)

type Actions struct {
	player       *common.Actor
	actionBroker common.ActionBroker
	state        *common.State
	animator     common.Animator
	ui           func() common.UserInterface
	config       common.Configuration
	audio        *audio.Player
	definitions  defs.DataContainer

	logError       func(message string)
	autoExplore    bool
	playerPath     []geometry.Point
	walkDirection  geometry.PointF
	lastBumpedTile geometry.Point
	isResting      bool
	actorSpeak     func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	queueAfterTurn func(call func())
}

func (a *Actions) DropItem(item *common.Item) {
	dropItemAction := common.NewDropAction(common.AbstractCause("DropItem"), a.player, item)
	a.actionBroker.Execute(dropItemAction)
}

func (a *Actions) PickUpItem() {
	state := a.state
	player := state.GetPlayer()
	if !player.IsAlive() || !player.CanPickup() {
		return
	}
	if !state.GetMap().IsItemAt(player.MapPosition()) {
		return
	}
	a.actionBroker.Execute(common.NewPickUpAction(common.AbstractCause("PickUpAction"), player))
}

type PlayerActionsContext struct {
	Player         *common.Actor
	ActionBroker   common.ActionBroker
	State          *common.State
	UI             func() common.UserInterface
	Speak          func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	Msg            func(message string)
	Animator       common.Animator
	Configuration  common.Configuration
	QueueAfterTurn func(call func())
}

func NewActions(context PlayerActionsContext) *Actions {
	return &Actions{
		player:         context.Player,
		actionBroker:   context.ActionBroker,
		state:          context.State,
		ui:             context.UI,
		actorSpeak:     context.Speak,
		animator:       context.Animator,
		logError:       context.Msg,
		config:         context.Configuration,
		queueAfterTurn: context.QueueAfterTurn,
		lastBumpedTile: geometry.Point{X: noBumpTileX, Y: noBumpTileX},
	}
}
func (a *Actions) EquipItem(item *common.Item) {
	a.actionBroker.Execute(common.NewEquipItemAction(common.AbstractCause("equip"), a.player, item))
}

func (a *Actions) UnequipItem(item *common.Item) {
	a.actionBroker.Execute(common.NewUnequipItemAction(common.AbstractCause("unequip"), a.player, item))
}

func (a *Actions) GetContextActionsForItem(item *common.Item) []util.MenuItem {
	// Wands, scrolls, potions and throwables all went with the combat systems.
	// Items have no activatable actions left, so there is nothing to offer.
	_ = item
	return nil
}

func (a *Actions) ManualMovePlayer(direction geometry.Point) {
	state := a.state
	player := state.GetPlayer()
	if !player.IsAlive() {
		return
	}
	a.CancelAutoMove()

	playerAction := common.PlayerMoveStep(player, direction, a.getActionContext())
	if playerAction == nil {
		return
	}

	a.actionBroker.Execute(playerAction)
}

func (a *Actions) getActionContext() common.ActionContext {
	return common.ActionContext{
		GameState:     a.state,
		UserInterface: a.ui,
		AudioPlayer:   a.audio,
		Data:          a.definitions,
		ActionBroker:  a.actionBroker,
		Speak:         a.actorSpeak,
	}
}

func (a *Actions) nextStepOnPlayerPath() {
	state := a.state
	msg := a.logError
	if !state.GetPlayer().IsAlive() {
		a.stopAutoMove()
		return
	}
	if a.IsPlayerAutoMoving() {
		if state.GetVisibility().GetVisibleEnemyCount(func(actor *common.Actor) bool {
			return actor.IsHostile() && actor.IsHostileTo(state.GetPlayer())
		}) > 0 {
			a.stopAutoMove()
			msg("You can see an enemy")
			return
		}

		// pop first element
		nextStep := a.playerPath[0]

		if state.GetMap().IsActorAt(nextStep) {
			//actorAt := state.GetMap().ActorAt(nextStep)
			a.stopAutoMove()
			if !a.autoExplore { // never explore into another actor, only deliberately
				direction := nextStep.Sub(state.GetPlayer().MapPosition())
				a.ManualMovePlayer(direction)
			}
			return
		}

		if state.GetMap().IsObjectAt(nextStep) {
			if a.state.IsClosedDoorAt(nextStep) {
				a.actionBroker.QueueForExecution(common.NewOpenDoorAction(common.AbstractCause("Player Movement"), state.GetPlayer(), nextStep))
				return
			} else if !a.autoExplore {
				objectAt := state.GetMap().ObjectAt(nextStep)
				if !objectAt.IsWalkable(state.GetPlayer()) {
					a.stopAutoMove()
					msg(fmt.Sprintf("You bump into %s", objectAt.Name()))
					objectAt.ActorBumpedInto(state.GetPlayer(), common.AnimateNow())
					return
				}
			}
		}

		if !state.GetMap().IsCurrentlyPassable(nextStep) {
			a.stopAutoMove()
			msg("You can't move there")
			return
		}

		a.playerPath = a.playerPath[1:]
		autoMoveAction := common.NewMoveAction(common.AbstractCause("Player Movement"), state.GetPlayer(), nextStep, true)
		autoMoveAction.SetAnimationDelay(0.08)
		a.actionBroker.QueueForExecution(autoMoveAction)
	} else if a.autoExplore {
		if target, hasTarget := a.getNextAutoExploreTarget(); hasTarget {
			a.continuePlayerAutoMovement(target)
		} else {
			msg("Nothing left to explore")
			a.stopAutoMove()
		}
	}
}
func (a *Actions) continuePlayerAutoMovement(targetPos geometry.Point) {
	currentMap := a.state.GetMap()
	var isWalkable func(point geometry.Point) bool
	if a.autoExplore {
		isWalkable = a.allowedForAutoExplore
	} else {
		isWalkable = a.allowedForAutoMove(targetPos)
	}
	pathToTarget := currentMap.GetJPSPath(a.state.GetPlayer().MapPosition(), targetPos, isWalkable)
	if len(pathToTarget) <= 1 {
		return
	}
	a.playerPath = pathToTarget[1:]
	a.nextStepOnPlayerPath()
}
func (a *Actions) allowedForAutoExplore(point geometry.Point) bool {
	currentMap := a.state.GetMap()
	isWalkable := currentMap.CurrentlyPassableAndSafeForActor(a.state.GetPlayer())
	return isWalkable(point) && !currentMap.IsStairsAt(point)
}
func (a *Actions) allowedForAutoMove(targetPos geometry.Point) func(point geometry.Point) bool {
	return func(point geometry.Point) bool {
		currentMap := a.state.GetMap()
		isWalkable := currentMap.CurrentlyPassableAndSafeForActor(a.state.GetPlayer())
		return isWalkable(point) || point == targetPos
	}
}

func (a *Actions) StartPlayerAutoMovement(targetPos geometry.Point) {
	player := a.state.GetPlayer()
	if !player.IsAlive() {
		return
	}
	currentMap := a.state.GetMap()

	// single step, like a manual move..
	if currentMap.MoveDistance(player.MapPosition(), targetPos) == 1 {
		// this seems to be an intended action..
		direction := targetPos.Sub(player.MapPosition())
		a.ManualMovePlayer(direction)
		return
	}

	if len(a.state.GetVisibility().GetVisibleEnemies()) > 0 {
		// player is trying to move while enemies are visible..
		a.logError("You can't run while enemies are visible")
		return
	}

	if !player.CanAutoMove() {
		a.logError("You can't move in a coordinated fashion right now")
		return
	}

	playerPosition := player.MapPosition()

	// check for item directly beneath us.. if so, pick it up
	if a.config.AutoPickup {
		if currentMap.IsItemAt(playerPosition) {
			itemAt := currentMap.ItemAt(playerPosition)
			if !itemAt.IsDroppedByPlayer() {
				a.actionBroker.Execute(common.NewPickUpAction(common.AbstractCause("PickUp"), player))
			}
		}
	}

	var isWalkable func(point geometry.Point) bool
	if a.autoExplore {
		isWalkable = a.allowedForAutoExplore
	} else {
		isWalkable = a.allowedForAutoMove(targetPos)
	}
	pathToTarget := currentMap.GetJPSPath(playerPosition, targetPos, isWalkable)
	if len(pathToTarget) <= 1 {
		a.stopAutoMove()
		return
	}
	a.animator.CancelAllAnimations()
	a.actionBroker.ClearQueuedActions(func(action common.Action) bool {
		if moveAction, ok := action.(*common.MoveAction); ok {
			return moveAction.GetActor() == player
		}
		return false
	})
	a.playerPath = pathToTarget[1:]
	a.nextStepOnPlayerPath()
}

func (a *Actions) getNextAutoExploreTarget() (geometry.Point, bool) {
	state := a.state
	player := state.GetPlayer()
	playerInv := player.GetInventory()
	playerPos := player.MapPosition()

	if !playerInv.IsFull() && a.config.AutoPickup {
		nearestItem := state.GetMap().GetNearestExploredItem(playerPos, func(item *common.Item) bool {
			return !item.IsDroppedByPlayer() && (!item.IsHidden() || state.GetFlags().IsSet(string(common.FlagHasTrueSight)))
		})
		if nearestItem != playerPos {
			return nearestItem, true
		}
	}

	unexplored := state.GetMap().GetNearestUnexploredWalkablePosition(playerPos)
	if unexplored != playerPos {
		return unexplored, true
	}
	return geometry.Point{}, false
}

func (a *Actions) CancelAutoMove() {
	a.stopAutoMove()
	a.animator.CancelAllAnimations()

}
func (a *Actions) stopAutoMove() {
	state := a.state
	a.actionBroker.ClearQueuedActions(func(action common.Action) bool {
		if moveAction, ok := action.(*common.MoveAction); ok {
			return moveAction.GetActor() == state.GetPlayer()
		}
		return false
	})

	a.autoExplore = false
	a.playerPath = nil
}

func (a *Actions) StartAutoExplore() {
	if !a.state.GetPlayer().CanAutoMove() {
		a.logError("You can't explore right now")
		return
	}
	target, hasTarget := a.getNextAutoExploreTarget()
	if a.autoExplore {
		a.stopAutoMove()
		a.logError("You stop exploring")
		return
	}
	if hasTarget {
		a.logError("You start exploring")
		a.autoExplore = true
		a.StartPlayerAutoMovement(target)
	} else {
		a.logError("Nothing to explore")
	}
}

func (a *Actions) goOneLevelUp() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	currentMap := state.GetMap()
	mapName := currentMap.GetName()
	if mapName == string(defs.LevelNameTown) {
		return
	}

	ourPos := state.GetPlayer().MapPosition()
	nearestStairsUp := currentMap.GetNearestTile(ourPos, func(tile gridmap.Tile) bool {
		return tile.IsStairsUp()
	})

	if nearestStairsUp == ourPos {
		a.logError("No known path")
		return
	}
	a.CancelAutoMove()
	a.StartPlayerAutoMovement(nearestStairsUp)
}

func (a *Actions) goOneLevelDown() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	currentMap := state.GetMap()
	//mapName := currentMap.Action()

	ourPos := state.GetPlayer().MapPosition()
	nearestStairsDown := currentMap.GetNearestTile(ourPos, func(tile gridmap.Tile) bool {
		return tile.IsStairsDown()
	})
	if nearestStairsDown == ourPos {
		a.logError("No known path")
		return
	}
	a.CancelAutoMove()
	a.StartPlayerAutoMovement(nearestStairsDown)
}

func (a *Actions) playerMoveOneStep(player *common.Actor, targetPos geometry.Point) {
	a.actionBroker.Execute(common.NewMoveAction(common.NewActorCause("Move", player), player, targetPos, true))
}
func (a *Actions) SetQueueAfterTurn(afterTurn func(call func())) {
	a.queueAfterTurn = afterTurn
}
func (a *Actions) updateFoV() {
	state := a.state
	player := state.GetPlayer()
	a.actionBroker.Execute(common.NewUpdateFoVAction(common.AbstractCause("FoV Update"), player))
}

func (a *Actions) IsPlayerAutoMoving() bool {
	return len(a.playerPath) > 0
}
