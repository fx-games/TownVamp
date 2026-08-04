package game

import (
	"math"
	"rpg/common"
	"rpg/geometry"
)

// The vampire walks off the grid; the town does not. He still owns exactly one
// grid cell at any moment — everything that asks where he is (field of view,
// zones, perception, the map's own occupancy) reads that cell and cannot tell
// the difference — but he crosses the space between cells continuously instead
// of being animated from one to the next.
const (
	// playerWalkRadius is how much of a cell he takes up when deciding whether
	// he fits. Under half a cell, so a one-tile doorway is passable without
	// having to line up on the grid first.
	//
	// ponytail: his sprite is a whole cell wide, so standing flush against a
	// wall overlaps it by a tenth of a cell. Shrink the sprite or inset the
	// wall's collision if that ever reads as a bug rather than as weight.
	playerWalkRadius = 0.4

	// noBumpTile is a cell no map contains, used to mean "nothing bumped yet".
	// Walking into a door has to fire once per approach, not once per tick.
	noBumpTileX = -1
)

// SetWalkDirection records which way the player is pushing. It is intent only:
// the movement itself happens in Advance, so it stops when the game is paused
// and runs at the same in-game rate as everything else.
func (a *Actions) SetWalkDirection(direction geometry.PointF) {
	if direction != (geometry.PointF{}) && a.IsPlayerAutoMoving() {
		a.CancelAutoMove()
	}
	a.walkDirection = direction
}

// walkPlayer integrates the player's position for one slice of in-game time.
func (a *Actions) walkPlayer(inGameSeconds float64) {
	player := a.state.GetPlayer()
	if player == nil || !player.IsAlive() || a.walkDirection == (geometry.PointF{}) {
		return
	}

	from := player.DrawPosition()
	delta := a.walkDirection.Normalize().Mul(a.playerTilesPerSecond() * inGameSeconds)

	// Axis at a time, so pushing into a wall at an angle slides along it rather
	// than stopping dead.
	to := from
	if a.playerFitsAt(geometry.PointF{X: from.X + delta.X, Y: from.Y}) {
		to.X += delta.X
	}
	if a.playerFitsAt(geometry.PointF{X: to.X, Y: from.Y + delta.Y}) {
		to.Y += delta.Y
	}

	if to == from {
		a.bumpInWalkDirection()
		return
	}
	a.lastBumpedTile.X = noBumpTileX

	player.SetDrawPosition(to)

	if newCell := cellOf(to); newCell != player.MapPosition() {
		a.enterCell(newCell)
	}
}

// playerTilesPerSecond keeps free movement on the same clock as a Townsperson's
// step: the move speed stat still says how long a cell costs, it is just spent
// continuously now instead of in one jump.
func (a *Actions) playerTilesPerSecond() float64 {
	return 1.0 / common.SecondsPerStep(a.state.GetPlayer().GetMoveSpeed())
}

// cellOf is the one grid cell a continuous position counts as. A draw position
// is the sprite's top-left corner, so the cell is the one holding its middle.
func cellOf(pos geometry.PointF) geometry.Point {
	return geometry.Point{X: int(math.Round(pos.X)), Y: int(math.Round(pos.Y))}
}

// playerFitsAt asks whether the player's box clears every cell it overlaps.
func (a *Actions) playerFitsAt(pos geometry.PointF) bool {
	currentMap := a.state.GetMap()
	player := a.state.GetPlayer()
	middle := pos.Add(geometry.PointF{X: 0.5, Y: 0.5})

	for _, corner := range boxCorners(middle, playerWalkRadius) {
		if !currentMap.IsWalkable(corner) {
			return false
		}
		if currentMap.IsActorAt(corner) && currentMap.ActorAt(corner) != player {
			return false
		}
	}
	return true
}

// boxCorners are the cells under the four corners of a box of the given half
// size. Duplicates are harmless — the caller only ever asks whether all of them
// are clear.
func boxCorners(middle geometry.PointF, halfSize float64) [4]geometry.Point {
	low, high := -halfSize, halfSize
	return [4]geometry.Point{
		cellUnder(middle, low, low),
		cellUnder(middle, high, low),
		cellUnder(middle, low, high),
		cellUnder(middle, high, high),
	}
}

func cellUnder(middle geometry.PointF, offsetX, offsetY float64) geometry.Point {
	return geometry.Point{
		X: int(math.Floor(middle.X + offsetX)),
		Y: int(math.Floor(middle.Y + offsetY)),
	}
}

// enterCell hands the grid its new occupant. It goes through the ordinary move
// action so that everything hung off a move still happens — field of view, map
// transitions, cell effects, what the player is told he is standing on — with
// only the drawing left alone, because the walker owns that now.
func (a *Actions) enterCell(cell geometry.Point) {
	player := a.state.GetPlayer()
	move := common.NewMoveAction(common.NewActorCause("Walk", player), player, cell, true)
	move.KeepDrawPosition()
	a.actionBroker.Execute(move)
}

// bumpInWalkDirection is what happens when the way is blocked: walking into a
// closed door opens it, walking into a vendor talks to them. Fired once per
// approach rather than once per tick, or holding a key against a door would
// queue sixty openings a second.
func (a *Actions) bumpInWalkDirection() {
	player := a.state.GetPlayer()
	step := dominantStep(a.walkDirection)
	target := player.MapPosition().Add(step)
	if target == a.lastBumpedTile {
		return
	}
	a.lastBumpedTile = target

	action := common.PlayerMoveStep(player, step, a.getActionContext())
	if action == nil {
		return
	}
	if _, isMove := action.(*common.MoveAction); isMove {
		return // the cell ahead is free; the walker will arrive there on its own
	}
	a.actionBroker.Execute(action)
}

// dominantStep reduces a free direction to the single grid step it most looks
// like, which is what the cell-based interactions expect.
func dominantStep(direction geometry.PointF) geometry.Point {
	if math.Abs(direction.X) >= math.Abs(direction.Y) {
		return geometry.Point{X: sign(direction.X)}
	}
	return geometry.Point{Y: sign(direction.Y)}
}

func sign(value float64) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
