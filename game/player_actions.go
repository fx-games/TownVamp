package game

import (
	"fmt"
	"rpg/audio"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
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

	getTimeKeeper  func() *common.TimeKeeper
	logError       func(message string)
	autoExplore    bool
	playerPath     []geometry.Point
	isResting      bool
	actorSpeak     func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	queueAfterTurn func(call func())
}

func (a *Actions) EndTurn(autTaken int) {
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(common.AbstractCause("Wait"), autTaken))
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
	pickUpAction := common.NewPickUpAction(common.AbstractCause("PickUpAction"), player)
	a.actionBroker.Execute(pickUpAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(pickUpAction.Cause(), pickUpAction.GetDelayInAut()))
}

type PlayerActionsContext struct {
	Player         *common.Actor
	ActionBroker   common.ActionBroker
	State          *common.State
	UI             func() common.UserInterface
	Speak          func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	GetTimeKeeper  func() *common.TimeKeeper
	Msg            func(message string)
	Animator       common.Animator
	Configuration  common.Configuration
	QueueAfterTurn func(call func())
}

func NewActions(context PlayerActionsContext) *Actions {
	return &Actions{
		player:        context.Player,
		actionBroker:  context.ActionBroker,
		state:         context.State,
		ui:            context.UI,
		actorSpeak:    context.Speak,
		animator:      context.Animator,
		getTimeKeeper: context.GetTimeKeeper,
		logError:      context.Msg,
		config:        context.Configuration,
		queueAfterTurn: context.QueueAfterTurn,
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

	stepCost := 100 / player.GetMoveSpeed()
	if player.GetEnergy() < stepCost {
		return // not enough time has passed for another step
	}

	playerAction := common.PlayerMoveStep(player, direction, a.getActionContext())
	if playerAction == nil {
		return
	}

	a.actionBroker.Execute(playerAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(playerAction.Cause(), playerAction.GetDelayInAut()))
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
				openDoorAction := common.NewOpenDoorAction(common.AbstractCause("Player Movement"), state.GetPlayer(), nextStep)
				a.actionBroker.QueueForExecution(openDoorAction)
				a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(openDoorAction.Cause(), openDoorAction.GetDelayInAut()))
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
		a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(autoMoveAction.Cause(), autoMoveAction.GetDelayInAut()))
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
				pickUpAction := common.NewPickUpAction(common.AbstractCause("PickUp"), player)
				a.actionBroker.Execute(pickUpAction)
				a.actionBroker.Execute(common.NewEndPlayerTurnAction(pickUpAction.Cause(), pickUpAction.GetDelayInAut()))
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
	moveAction := common.NewMoveAction(common.NewActorCause("Move", player), player, targetPos, true)
	a.actionBroker.Execute(moveAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(moveAction.Cause(), moveAction.GetDelayInAut()))
}
func (a *Actions) RestUntilHealed() {
	state := a.state
	if len(state.GetVisibility().GetVisibleEnemies()) > 0 {
		a.logError("You can't rest with enemies nearby")
		return
	}
	/*
	hpRegenRate := state.GetPlayer().GetStats().GetAttributeAsFloat(stats.HealthRegeneration)
	if hpRegenRate <= 0 {
		a.isResting = false
		a.logError("You won't heal through resting")
		return
	}
	 */
	healthResource := state.GetPlayer().GetStats().GetResource(stats.Health)
	if healthResource.IsAtMax() {
		a.isResting = false
		if a.isResting {
			a.logError("Your spirits have been restored")
		} else {
			a.logError("You are already at full health")
		}
		return
	}

	playerInventory := state.GetPlayer().GetInventory()
	if !playerInventory.HasFood() {
		a.isResting = false
		a.logError("You have no food to eat")
		return
	}

	playerInventory.UseFood()

	a.isResting = true
	userInterface := a.ui()
	userInterface.FadeToBlack()
	a.queueAfterTurn(func() {
		a.isResting = false
		userInterface.FadeFromBlack()
		a.actionBroker.QueueForExecution(common.NewChangeResourceAction(common.AbstractCause("Resting"), state.GetPlayer(), stats.Health, float64(healthResource.GetMax())))
	})
	a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(common.AbstractCause("Resting"), 10))
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

