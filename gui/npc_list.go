package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"rpg/defs"
	"rpg/geometry"
	"strconv"
)

// NPCList is the watch list: every Townsperson who is both on the rendered part
// of the map and inside the vampire's field of view. Rows are ordered by the
// Townsperson's fixed Number, so the same person always sits under the same
// digit no matter who else is on screen.
type NPCList struct {
	uiController             *UIController
	offsetFromTopRightCorner geometry.Point
	topLeft                  geometry.Point
	iconScale                float64
}

func NewNPCList(uiController *UIController, offsetFromTopRightCorner geometry.Point) *NPCList {
	n := &NPCList{
		uiController:             uiController,
		offsetFromTopRightCorner: offsetFromTopRightCorner,
		iconScale:                4,
	}
	n.Layout()
	return n
}

func (n *NPCList) OnCommand(command CommandType) bool { return false }

func (n *NPCList) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool { return false }

func (n *NPCList) OnMouseMoved(x int, y int) (bool, Tooltip) { return false, NoTooltip{} }

func (n *NPCList) OnMouseWheel(x int, y int, dy float64) bool { return false }

func (n *NPCList) ShouldClose() bool { return false }

func (n *NPCList) Layout() {
	screenSize := n.uiController.GetScreenSize()
	n.topLeft = geometry.Point{
		X: screenSize.X - n.offsetFromTopRightCorner.X,
		Y: n.offsetFromTopRightCorner.Y,
	}
}

func (n *NPCList) Draw() {
	u := n.uiController
	uiRenderer := u.GetRenderer()
	iconSize := int(float64(u.worldTexture.GetTileSize().Y) * n.iconScale)
	rowGap := 4
	nameX := float64(n.topLeft.X + (2 * iconSize) + rowGap)

	drawY := n.topLeft.Y
	for _, actor := range u.watchedTownspeople() {
		baseline := float64(drawY + iconSize)
		uiRenderer.DrawTTFOnScreen(float64(n.topLeft.X), baseline, strconv.Itoa(actor.GetTownsperson().Number), defs.DarkTextColor)
		// Drawn unflipped on purpose: the spritesheet default looks right, and a
		// row that flips about as its subject walks is harder to read at a glance.
		uiRenderer.DrawOnScreenWithScale(n.topLeft.X+iconSize, drawY, u.getActorIcon(actor, u.worldTicks), n.iconScale)
		uiRenderer.DrawTTFOnScreen(nameX, baseline, actor.Name(), defs.DarkTextColor)
		drawY += iconSize + rowGap
	}
}
