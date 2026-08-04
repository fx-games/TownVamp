package game

import (
	"math"
	"rpg/geometry"
	"testing"
)

// Free movement has one invariant the rest of the game depends on: wherever the
// player is drawn, the single grid cell he counts as is the cell under him, and
// that cell is always one he could stand in. Everything that asks where he is —
// field of view, zones, perception — reads that cell and must never be handed a
// wall.
func TestWalkingLeavesTheGridButNotTheMap(t *testing.T) {
	game := newTestGame(t)
	state := game.GetWorldState()
	player := state.GetPlayer()

	direction, ok := anyOpenDirection(game)
	if !ok {
		t.Fatal("player starts with no walkable neighbour")
	}
	game.actions.SetWalkDirection(direction)

	startCell := player.MapPosition()
	leftTheGrid := false

	// A cell costs sixty in-game seconds at the default move speed, so this is
	// two cells' worth of walking taken in small slices.
	for i := 0; i < 120; i++ {
		game.Advance(1)

		drawnAt := player.DrawPosition()
		if drawnAt.X != math.Trunc(drawnAt.X) || drawnAt.Y != math.Trunc(drawnAt.Y) {
			leftTheGrid = true
		}
		if cell := player.MapPosition(); cell != cellOf(drawnAt) {
			t.Fatalf("player is drawn at %v but counts as standing in %v", drawnAt, cell)
		}
		if !state.GetMap().IsWalkable(player.MapPosition()) {
			t.Fatalf("player walked into %v, which is not walkable", player.MapPosition())
		}
	}

	if !leftTheGrid {
		t.Error("player never left a whole-numbered position; movement is still grid bound")
	}
	if player.MapPosition() == startCell {
		t.Errorf("player never left %v after two cells' worth of walking", startCell)
	}
}

// Walking into a wall has to stop, not stutter through it and not stall the game.
func TestWalkingIntoAWallGoesNowhere(t *testing.T) {
	game := newTestGame(t)
	player := game.GetWorldState().GetPlayer()

	facingWall, ok := cellFacingAWall(game)
	if !ok {
		t.Fatal("the map has no walkable cell next to a wall")
	}
	game.GetWorldState().GetMap().MoveActor(player, facingWall)
	player.SetDrawPosition(facingWall.ToPointF())
	game.actions.SetWalkDirection(geometry.PointF{X: 1})

	wall := facingWall.Add(geometry.Point{X: 1})
	before := player.DrawPosition()
	for i := 0; i < 60; i++ {
		game.Advance(1)
		if cellOf(player.DrawPosition()) == wall {
			t.Fatalf("player walked into the wall at %v", wall)
		}
	}

	// He does close the gap: he stops with his box flush against the wall, which
	// is under a tenth of a cell from where he started.
	restingAt := player.DrawPosition()
	if drift := restingAt.X - before.X; drift > 0.5-playerWalkRadius+1e-9 {
		t.Errorf("player pushed %v into the wall, further than his box allows", drift)
	}

	for i := 0; i < 10; i++ {
		game.Advance(1)
	}
	if player.DrawPosition() != restingAt {
		t.Errorf("player kept creeping into the wall: %v then %v", restingAt, player.DrawPosition())
	}
}

func anyOpenDirection(game *GameInstance) (geometry.PointF, bool) {
	currentMap := game.GetWorldState().GetMap()
	from := game.GetWorldState().GetPlayer().MapPosition()
	for _, step := range []geometry.Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		if currentMap.IsWalkable(from.Add(step)) {
			return geometry.PointF{X: float64(step.X), Y: float64(step.Y)}, true
		}
	}
	return geometry.PointF{}, false
}

// cellFacingAWall is a walkable cell with something solid immediately to its
// right, and nobody standing in either.
func cellFacingAWall(game *GameInstance) (geometry.Point, bool) {
	currentMap := game.GetWorldState().GetMap()
	size := currentMap.MapSize()
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			here := geometry.Point{X: x, Y: y}
			if currentMap.IsWalkable(here) && !currentMap.IsActorAt(here) && !currentMap.IsWalkable(here.Add(geometry.Point{X: 1})) {
				return here, true
			}
		}
	}
	return geometry.Point{}, false
}
