package game

import (
	"rpg/geometry"
	"testing"
)

// Windows are the only tiles that block movement without blocking sight, and
// that gap is what lets the vampire watch a Private Action happen in a room he
// cannot walk into. Tagging one IsBlockingView in LDtk would silently make every
// house opaque again, so the guard lives here rather than in the map file.
func TestWindowsBlockMovementButNotSight(t *testing.T) {
	game := newTestGame(t)
	m := game.GetWorldState().GetMap()

	windows := 0
	for y := 0; y < 40; y++ {
		for x := 0; x < 72; x++ {
			p := geometry.Point{X: x, Y: y}
			if !m.Contains(p) {
				continue
			}
			icon := m.GetCell(p).TileType.Icon()
			if icon != 763 && icon != 764 {
				continue
			}
			windows++
			if m.IsWalkable(p) {
				t.Errorf("window at %v is walkable", p)
			}
			if !m.IsTransparent(p) {
				t.Errorf("window at %v blocks sight; it is meant to be seen through", p)
			}
		}
	}
	if windows == 0 {
		t.Fatal("no windows on the map")
	}
	t.Logf("%d windows, all solid and see-through", windows)
}
