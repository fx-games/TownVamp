package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"rpg/common"
	"rpg/geometry"
)

type InputReceiver interface {
	OnCommand(command CommandType) bool
	OnMouseClicked(button ebiten.MouseButton, x int, y int) bool
	OnMouseMoved(x int, y int) (bool, Tooltip)
	OnMouseWheel(x int, y int, dy float64) bool
}

type TextInputReceiver interface {
	InputReceiver
	OnKeyPressed(key ebiten.Key, shiftPressed bool)
	SetTick(tick uint64)
}
type MouseEvent int

const (
	MouseEventLeftClicked MouseEvent = iota
	MouseEventMoved
)

type CommandType int

const (
	PlayerCommandUp CommandType = iota
	PlayerCommandDown
	PlayerCommandLeft
	PlayerCommandRight

	PlayerCommandUpLeft
	PlayerCommandUpRight
	PlayerCommandDownLeft
	PlayerCommandDownRight
	PlayerCommandConfirm
	PlayerCommandCancel
	PlayerCommandOptions
)

func isBeingPressedRepeatedly(key ebiten.Key) bool {
	pressedForTicks := inpututil.KeyPressDuration(key)
	ticksBetweenMovement := 20
	if pressedForTicks > 30 {
		ticksBetweenMovement = 8
	}
	isRepetitionTick := pressedForTicks > 0 && pressedForTicks%ticksBetweenMovement == 0
	return inpututil.IsKeyJustPressed(key) || (ebiten.IsKeyPressed(key) && isRepetitionTick)
}

func (u *UIController) handleKeyboardInput() {
	// must-have
	//  - movement
	//  - wait
	//  - aim ranged attack
	//  - quick ranged attack
	//  - show log
	//  - show inventory
	//  - show stats
	//  - auto explore
	//  - auto attack

	// extra
	//  - pick up item
	//  - quick travel

	act := u.context.PlayerAction()

	u.shiftIsBeingPressedThisTick = ebiten.IsKeyPressed(ebiten.KeyShift)

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		u.showHelpText()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		u.Cancel()
		return
	}
	if u.IsUIModalVisible() {
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			u.OnCommand(PlayerCommandUp)
			return
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			u.OnCommand(PlayerCommandDown)
			return
		} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			u.OnCommand(PlayerCommandLeft)
			return
		} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			u.OnCommand(PlayerCommandRight)
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			u.OnCommand(PlayerCommandConfirm)
			return
		}
		var keys []ebiten.Key
		keys = inpututil.AppendJustPressedKeys(keys)
		for _, key := range keys {
			if key == ebiten.KeyShift || key == ebiten.KeyControl || key == ebiten.KeyAlt {
				continue
			}
			u.OnKeyPressed(key, u.shiftIsBeingPressedThisTick)
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		u.ShowLog()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		u.mapRenderer.CenterOnFloat(u.context.GameState().GetPlayer().DrawPosition())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		u.showFlags()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		u.context.TogglePause()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		u.context.Kill()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		u.context.HideBody()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		u.context.WaitUntil(common.WaitUntilDawn)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
		u.context.WaitUntil(common.WaitUntilDusk)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		u.ShowFullStats()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		u.context.DebugActions().KillLevel()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
		u.renderingMode = (u.renderingMode + 1) % 3
		u.configuration.FlipActorsHorizontally = u.renderingMode == GraphicRendering
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		u.disableFogOfWar = !u.disableFogOfWar
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		u.isFullScreen = !ebiten.IsFullscreen()
		ebiten.SetFullscreen(u.isFullScreen)
		u.OnScreenSizeChanged()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		u.FlashScreen(color.White, 0.1)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		//g.actionBroker.Execute(common.NewEndPlayerTurnAction(common.AbstractCause("Travel"), 10))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		act.StartAutoExplore()
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		u.mapRenderer.ScrollBy(geometry.Point{X: 8, Y: 0})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		u.mapRenderer.ScrollBy(geometry.Point{X: -8, Y: 0})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		u.mapRenderer.ScrollBy(geometry.Point{X: 0, Y: -8})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		u.mapRenderer.ScrollBy(geometry.Point{X: 0, Y: 8})
	}

	if isBeingPressedRepeatedly(ebiten.KeyW) {
		act.ManualMovePlayer(geometry.Point{X: 0, Y: -1})
	} else if isBeingPressedRepeatedly(ebiten.KeyS) {
		act.ManualMovePlayer(geometry.Point{X: 0, Y: 1})
	} else if isBeingPressedRepeatedly(ebiten.KeyA) {
		act.ManualMovePlayer(geometry.Point{X: -1, Y: 0})
	} else if isBeingPressedRepeatedly(ebiten.KeyD) {
		act.ManualMovePlayer(geometry.Point{X: 1, Y: 0})
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		u.ChangeMapZoom(1)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		u.ChangeMapZoom(-1)
	}
}

func (u *UIController) handleMouseInput() {
	mousePosInPixelsX, mousePosInPixelsY := ebiten.CursorPosition()
	mousePosInPixelsX = int(float64(mousePosInPixelsX) / u.deviceDPIScale)
	mousePosInPixelsY = int(float64(mousePosInPixelsY) / u.deviceDPIScale)

	mouseMoved := false
	if u.mousePosInPixels.X != mousePosInPixelsX || u.mousePosInPixels.Y != mousePosInPixelsY {
		mouseMoved = true
	}
	u.mousePosInPixels.X = mousePosInPixelsX
	u.mousePosInPixels.Y = mousePosInPixelsY

	newMapPos := u.mapRenderer.GetMapCellAtScreenPos(u.mousePosInPixels)
	/*
	   if newMapPos != u.mousePosOnMap {
	       println(fmt.Sprintf("Mouse pos: %d, %d", newMapPos.X, newMapPos.Y))
	   }

	*/
	u.mousePosOnMap = newMapPos

	if mouseMoved {
		handled, toolTip := u.OnMouseMoved(mousePosInPixelsX, mousePosInPixelsY)
		if handled {
			u.ShowTooltipWithDelay(toolTip)
		} else {
			u.HideTooltip()
		}

	}
	noButton := ebiten.MouseButton(-1)
	mouseButton := noButton
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseButton = ebiten.MouseButtonLeft
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		mouseButton = ebiten.MouseButtonRight
	}
	if mouseButton != noButton {
		handledByUI := u.OnMouseClicked(mouseButton, mousePosInPixelsX, mousePosInPixelsY)
		if handledByUI {
			return
		}
		visibleMap := u.mapRenderer.GetVisibleMap()
		state := u.context.GameState()
		if !state.GetMap().Contains(u.mousePosOnMap) || !visibleMap.Contains(u.mousePosOnMap) {
			return
		}
		if !handledByUI {
			u.onMapClicked(mouseButton, u.mousePosOnMap)
		}
	}
}
func (u *UIController) onMapClicked(mouseButton ebiten.MouseButton, mapPos geometry.Point) {
	if u.IsUIModalVisible() {
		return
	}
	act := u.context.PlayerAction()
	state := u.context.GameState()
	currentMap := state.GetMap()
	if mouseButton == ebiten.MouseButtonLeft && currentMap.IsExplored(mapPos) {
		act.CancelAutoMove()
		act.StartPlayerAutoMovement(mapPos)
	}
}
