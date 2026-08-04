package gui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/renderer"
	"rpg/util"
	"strings"
)

const (
	// npcWindowColumns is the window's width in characters. The font is a fixed
	// pitch bitmap face, so measuring one run of that many glyphs gives the exact
	// pixel width and the wrap can be done in characters.
	npcWindowColumns = 56

	// npcWindowMaxScreenFraction is how much of the screen height the window may
	// take before it starts scrolling instead of growing.
	npcWindowMaxScreenFraction = 0.7

	// npcWindowActivityIndent is the width of the marker and the hours in front
	// of a schedule line: "> 06:00-12:00  ".
	npcWindowActivityIndent = 15
)

// NPCWindow is what the player gets for pressing a Townsperson's number: who
// they are and the day they intend to have, with the hour they are in right now
// marked. It scrolls, and wraps at word boundaries.
type NPCWindow struct {
	uiController *UIController
	actor        *common.Actor

	scroll      ScrollingContent
	shouldClose bool

	topLeft      geometry.Point
	bounds       geometry.Rect
	lineHeight   int
	charWidth    int
	visibleLines int
	padding      geometry.Point
}

func NewNPCWindow(uiController *UIController, actor *common.Actor) *NPCWindow {
	w := &NPCWindow{
		uiController: uiController,
		actor:        actor,
		padding:      geometry.Point{X: 12, Y: 10},
	}
	w.scroll = NewScrollingContentWithFunctions(
		func() int { return len(w.contentLines()) },
		// In lines, not pixels: everything the scroller does is index arithmetic.
		func() geometry.Rect { return geometry.NewRect(0, 0, npcWindowColumns, w.visibleLines) },
	)
	w.Layout()
	return w
}

func (w *NPCWindow) Layout() {
	uiRenderer := w.uiController.GetRenderer()
	width, height := uiRenderer.MeasureString(strings.Repeat("M", npcWindowColumns))
	w.charWidth = int(width) / npcWindowColumns
	w.lineHeight = int(height) + 4

	screenSize := w.uiController.GetScreenSize()
	maxLines := int(float64(screenSize.Y)*npcWindowMaxScreenFraction) / w.lineHeight
	w.visibleLines = min(max(len(w.contentLines()), 1), max(maxLines, 1))

	size := geometry.Point{
		X: int(width) + (2 * w.padding.X),
		Y: (w.visibleLines * w.lineHeight) + (2 * w.padding.Y),
	}
	w.topLeft = geometry.Point{X: (screenSize.X - size.X) / 2, Y: (screenSize.Y - size.Y) / 2}
	w.bounds = geometry.NewRect(w.topLeft.X, w.topLeft.Y, w.topLeft.X+size.X, w.topLeft.Y+size.Y)
}

func (w *NPCWindow) Draw() {
	uiRenderer := w.uiController.GetRenderer()
	uiRenderer.DrawDefaultBorder(w.topLeft, w.bounds.Size())

	lines := w.contentLines()
	textX := float64(w.topLeft.X + w.padding.X)
	for row := 0; row < w.visibleLines; row++ {
		index := w.scroll.getLine(row)
		if index < 0 || index >= len(lines) {
			continue
		}
		// DrawTTFOnScreen takes the baseline, so a row sits one line below its top.
		baseline := float64(w.topLeft.Y + w.padding.Y + ((row + 1) * w.lineHeight))
		uiRenderer.DrawTTFOnScreen(textX, baseline, lines[index].Text, lines[index].TextColor)
	}
	w.drawScrollHints(uiRenderer, len(lines))
}

// drawScrollHints says there is more in either direction. The scroller's own
// indicators draw on the font grid, and this window is laid out in pixels.
func (w *NPCWindow) drawScrollHints(uiRenderer *renderer.TileRenderer, lineCount int) {
	if lineCount <= w.visibleLines {
		return
	}
	x := float64(w.bounds.Max.X - w.padding.X - w.charWidth)
	if w.scroll.canScrollUp() {
		uiRenderer.DrawTTFOnScreen(x, float64(w.topLeft.Y+w.padding.Y+w.lineHeight), "^", defs.LightTextColor)
	}
	if w.scroll.canScrollDown() {
		uiRenderer.DrawTTFOnScreen(x, float64(w.bounds.Max.Y-w.padding.Y), "v", defs.LightTextColor)
	}
}

// contentLines is the window's whole text, already wrapped. Rebuilt on every
// draw because the marked entry moves with the clock.
func (w *NPCWindow) contentLines() []util.LabelText {
	person := w.actor.GetTownsperson()
	if person == nil {
		return nil
	}
	clock := w.uiController.context.GameState().GetClock()
	return npcWindowLines(person, w.actor.Name(), clock.SecondsSinceMidnight())
}

// npcWindowLines is the whole of what the window says. Free of the renderer and
// of the game so that what it claims about someone's day can be checked.
func npcWindowLines(person *common.Townsperson, name string, secondsSinceMidnight float64) []util.LabelText {
	current, hasCurrent := person.Schedule().EntryAt(secondsSinceMidnight)

	var lines []util.LabelText
	add := func(text string, textColor color.Color) {
		lines = append(lines, util.LabelText{Text: text, TextColor: textColor})
	}
	add(fmt.Sprintf("%d  %s", person.Number, name), defs.LightTextColor)
	add(titleCase(string(person.Role)), defs.DarkTextColor)
	add("", defs.DarkTextColor)
	add("Their day", defs.DarkTextColor)

	for _, entry := range person.Schedule() {
		isNow := hasCurrent && entry.StartSecond == current.StartSecond && entry.Activity == current.Activity
		marker, textColor := "  ", defs.DarkTextColor
		if isNow {
			marker, textColor = "> ", defs.LightTextColor
		}
		activity := strings.ReplaceAll(entry.Activity, "_", " ")
		if entry.IsObligation() {
			activity += " (owed)"
		}
		// The hours are a column of their own: the activity wraps under itself
		// rather than back under the clock.
		head := fmt.Sprintf("%s%s-%s  ", marker, hourMinute(entry.StartSecond), hourMinute(entry.EndSecond))
		chunks := util.AutoLayout(activity, npcWindowColumns-npcWindowActivityIndent)
		add(head+chunks[0], textColor)
		for _, rest := range chunks[1:] {
			add(strings.Repeat(" ", npcWindowActivityIndent)+rest, textColor)
		}
	}
	return lines
}

func hourMinute(secondsSinceMidnight int) string {
	return fmt.Sprintf("%02d:%02d", secondsSinceMidnight/3600, (secondsSinceMidnight%3600)/60)
}

func titleCase(word string) string {
	if word == "" {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

func (w *NPCWindow) OnCommand(command CommandType) bool {
	switch command {
	case PlayerCommandCancel, PlayerCommandConfirm:
		w.shouldClose = true
	case PlayerCommandUp:
		w.scroll.ScrollUp(1)
	case PlayerCommandDown:
		w.scroll.ScrollDown(1)
	}
	return true
}

func (w *NPCWindow) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
	if button == ebiten.MouseButtonRight || !w.bounds.Contains(geometry.Point{X: x, Y: y}) {
		w.shouldClose = true
	}
	return true
}

func (w *NPCWindow) OnMouseMoved(x int, y int) (bool, Tooltip) { return false, NoTooltip{} }

func (w *NPCWindow) OnMouseWheel(x int, y int, dy float64) bool {
	return w.scroll.OnMouseWheel(x, y, dy)
}

func (w *NPCWindow) ShouldClose() bool { return w.shouldClose }
