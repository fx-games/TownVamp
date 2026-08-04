package gui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	"io"
	"math"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/renderer"
	"rpg/stats"
	"rpg/util"
	"slices"
)

type TilemapAnimInfo struct {
	FrameCount     int
	DelayInSeconds float64
}

type UIWidget interface {
	InputReceiver
	Draw()
	ShouldClose() bool
	Layout()
}

type UIRenderer interface {
	DrawStringOnGrid(gridX int, gridY int, text string, color color.Color)
	DrawCharOnGrid(gridX int, gridY int, char rune, textColor color.Color)

	DrawOnScreen(screenX int, screenY int, icon int32)
	DrawCharOnScreen(screenX int, screenY int, char rune, scale float64, textColor color.Color)
	DrawOnScreenWithScale(screenX int, screenY int, icon int32, scale float64)
	DrawOnMap(highlight geometry.Point, icon int32, color color.Color)
	DrawOnGrid(gridX int, gridY int, icon int32)
	DrawOnFontGrid(gridX int, gridY int, icon int32)
	GetFontScale() float64
	GetFontGridSize() geometry.Point
	DrawTTFOnScreen(screenX float64, screenY float64, line string, color color.Color)
	MeasureString(line string) (float64, float64)
	DrawDefaultBorder(topLeftScreen geometry.Point, size geometry.Point)
	DrawColoredBorder(topLeftScreen geometry.Point, size geometry.Point, fillColor, borderColor color.Color)
	DrawColoredRect(topLeftScreen geometry.Point, size geometry.Point, fillColor color.Color)
	DrawTTFOnScreenWithColorCodes(screenX float64, screenY float64, textWithColorCodes []util.ColoredTextPart)
	DrawImageOnScreen(screenX int, screenY int, size geometry.Point, image *ebiten.Image)

	ParseColorCodedText(string) []util.ColoredTextPart
}

func NewUIController(tileRenderer *renderer.TileRenderer, context common.UIContext, config common.Configuration) *UIController {
	u := &UIController{
		tileRenderer:   tileRenderer,
		screenSize:     config.WindowSize,
		configuration:  config,
		context:        context,
		worldTexture:   context.GameState().GetWorldTexture(),
		deviceDPIScale: ebiten.DeviceScaleFactor(),
	}
	u.setTileScale(4)
	return u
}

type UIConfig struct {
	DrawTilesBehindActors  bool
	DrawTilesBehindItems   bool
	DrawTilesBehindObjects bool
	ActorBobbingEnabled    bool
	ItemNamesInShop        common.ItemNamingScheme
	KeepPlayerCentered     bool
}

type UIController struct {
	modal        UIWidget
	overlays     []UIWidget
	tileRenderer *renderer.TileRenderer
	mapRenderer  *renderer.MapRenderer
	worldTexture renderer.TextureAtlas
	asciiTexture renderer.TextureAtlas

	context                     common.UIContext
	isMouseCapturedByUI         bool
	shiftIsBeingPressedThisTick bool

	currentTooltip           Tooltip
	ticksUntilTooltipAppears uint64
	screenSize               geometry.Point
	uiBarWidth               int
	elevatedOverlay          UIWidget

	statsLabel *BasicLabel
	inventory  *IconOverlay
	npcList    *NPCList
	miniMap    *MinimapOverlay
	message    PrintMessage

	configuration common.Configuration

	setToExactFit    bool
	isFullScreen     bool
	logHook          func(string)
	worldTicks       uint64
	shouldQuitGame   bool
	mousePosInPixels geometry.Point

	// lastPlayerTile is what tells a tick where the player moved from one where
	// he only stood there. Most of the per-turn UI work is a function of the
	// player's grid position and nothing else.
	lastPlayerTile geometry.Point

	tilemapAnimations          map[int32]TilemapAnimInfo
	deviceDPIScale             float64
	tileScale                  float64
	globalFlashColorTicksLeft  int
	globalFlashColorTicksStart int
	globalScaleColor           color.Color
	disableFogOfWar            bool
	mousePosOnMap              geometry.Point
	msg                        func(message string) // this might be a smell already? Should the UI really generate messages to the player?
	renderingMode              RenderingMode
	globalFadeColorTicksLeft   int
	globalFadeColorTicksStart  int
	fadeFromBlack              bool
	keepBlack                  bool
}

func (u *UIController) FadeToBlack() {
	u.globalFadeColorTicksLeft = int(util.SecondsToTicks(2))
	u.globalFadeColorTicksStart = u.globalFadeColorTicksLeft
	u.fadeFromBlack = false
}

func (u *UIController) FadeFromBlack() {
	u.globalFadeColorTicksLeft = int(util.SecondsToTicks(2))
	u.globalFadeColorTicksStart = u.globalFadeColorTicksLeft
	u.fadeFromBlack = true
}

func (u *UIController) SetDebugActions(actions func() common.DebugActions) {
	u.context.DebugActions = actions
}

func (u *UIController) Say(actor *common.Actor, lines []string) {
	// TODO
}
func (u *UIController) GetVisibleMap() geometry.Rect {
	return u.mapRenderer.GetVisibleMap()
}

func (u *UIController) ShowLogMessage(message string) {
	if message == "" {
		return
	}
	rawText := common.RemoveColorCodes(message)
	u.message.rawText = rawText
	u.message.textForRenderer = u.tileRenderer.ParseColorCodedText(message)
	ticks := util.SecondsToTicks(3)
	u.message.ticksLeft = ticks
	u.message.ticksStart = ticks
}

func GetAutoFitRectWithExtraSpace[T fmt.Stringer](text []T, screenSize geometry.Point, lineDist int, padding geometry.Point, measureString func(line string) (float64, float64)) (geometry.Point, PositionInfo) {
	posInfo := SizeNeededForText(measureString, text, lineDist)
	posInfo.SizeInPixels.X += padding.X
	posInfo.SizeInPixels.Y += padding.Y

	startX := (screenSize.X - posInfo.SizeInPixels.X) / 2
	//endX := startX + width

	startY := (screenSize.Y - posInfo.SizeInPixels.Y) / 2
	//endY := startY + height

	topLeft := geometry.Point{X: startX, Y: startY}
	//bottomRight := geometry.Point{X: endX, Y: endY}
	return topLeft, posInfo
}

func (u *UIController) GetXPosAndHeightForIconText(text []string) (int, int, int) {
	//TODO
	return 0, 0, 0
}

func (u *UIController) ChangeMapZoom(sign int) {

	change := 0.25 * float64(sign)
	minScale := 0.25
	maxScale := 16.0

	currentScale := u.tileScale
	if sign > 0 && u.setToExactFit {
		// round to next 0.25
		currentScale = math.Floor(currentScale*4) / 4
		u.setToExactFit = false
	}
	targetScale := util.Clamp(minScale, maxScale, currentScale+change)

	mapSize := u.context.GameState().GetMap().MapSize()
	tileSize := u.worldTexture.GetTileSize()

	targetMapSize := geometry.Point{
		X: int(float64(mapSize.X) * float64(tileSize.X) * targetScale),
		Y: int(float64(mapSize.Y) * float64(tileSize.Y) * targetScale),
	}

	windowSize := u.GetDeviceIndependentScreenSize()
	if u.HasOverlays() {
		windowSize = windowSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
	}

	centerMapPos := u.mapRenderer.GetVisibleMap().Center()

	tooSmall := targetMapSize.X < windowSize.X && targetMapSize.Y < windowSize.Y
	if tooSmall {
		u.disableScrolling()
		return
	}

	u.setTileScale(targetScale)
	u.mapRenderer.CenterOn(centerMapPos)
}

func (u *UIController) disableScrolling() {
	u.setToExactFit = true
	state := u.context.GameState()
	mapSize := state.GetMap().MapSize()
	tileSize := u.worldTexture.GetTileSize()

	windowSize := u.GetDeviceIndependentScreenSize()
	if u.HasOverlays() {
		windowSize = windowSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
	}
	// calculate exact scale to fit
	mapSizeW := float64(mapSize.X) * float64(tileSize.X)
	mapSizeH := float64(mapSize.Y) * float64(tileSize.Y)
	targetScaleH := float64(windowSize.X) / mapSizeW
	targetScaleV := float64(windowSize.Y) / mapSizeH
	targetScale := min(targetScaleH, targetScaleV)
	u.mapRenderer.ResetScrolling()

	u.setTileScale(targetScale)
}

func (u *UIController) GetTopModal() UIWidget {
	return u.modal
}

func (u *UIController) IsUIModalVisible() bool {
	return u.modal != nil
}
func (u *UIController) PushWidget(widget UIWidget) {
	u.modal = widget
}

func (u *UIController) PopWidget() {
	u.modal = nil
}

func (u *UIController) OnCommand(command CommandType) bool {
	if u.IsUIModalVisible() && u.GetTopModal().OnCommand(command) {
		return true
	}
	for _, overlay := range u.overlays {
		if handled := overlay.OnCommand(command); handled {
			return true
		}
	}
	return false
}

func (u *UIController) OnMouseClicked(button ebiten.MouseButton, screenX int, screenY int) bool {
	u.isMouseCapturedByUI = false

	if u.IsUIModalVisible() && u.GetTopModal().OnMouseClicked(button, screenX, screenY) {
		return true
	}

	for _, overlay := range u.overlays {
		if handled := overlay.OnMouseClicked(button, screenX, screenY); handled {
			u.isMouseCapturedByUI = true
			return true
		}
	}

	return false
}

func (u *UIController) OnMouseMoved(screenX int, screenY int) (bool, Tooltip) {
	u.isMouseCapturedByUI = false

	for _, overlay := range u.overlays {
		if handled, toolTip := overlay.OnMouseMoved(screenX, screenY); handled {
			u.isMouseCapturedByUI = true
			return handled, toolTip
		}
	}

	if !u.IsUIModalVisible() {
		return false, NoTooltip{}
	}

	u.isMouseCapturedByUI = true

	return u.GetTopModal().OnMouseMoved(screenX, screenY)
}

func (u *UIController) IsMouseCapturedByUI() bool {
	return u.isMouseCapturedByUI
}

func (u *UIController) OnMouseWheel(x int, y int, dy float64) bool {
	if !u.IsUIModalVisible() {
		return false
	}
	return u.GetTopModal().OnMouseWheel(x, y, dy)
}

func (u *UIController) drawUI() {

	defer u.drawTooltip()
	for _, overlay := range u.overlays {
		if overlay == u.elevatedOverlay {
			continue
		}
		overlay.Draw()
	}
	if !u.IsUIModalVisible() && u.elevatedOverlay == nil {
		return
	}
	renderer := u.GetRenderer()
	renderer.DrawColoredRect(geometry.Point{X: 0, Y: 0}, u.screenSize, color.RGBA{A: 140})

	if u.elevatedOverlay != nil {
		u.elevatedOverlay.Draw()
	}

	if !u.IsUIModalVisible() {
		return
	}

	if u.GetTopModal().ShouldClose() {
		u.PopWidget()
	} else {
		u.GetTopModal().Draw()
	}
}

func (u *UIController) drawTooltip() {
	if u.currentTooltip != nil && u.ticksUntilTooltipAppears == 0 {
		u.currentTooltip.Draw()
	}
}

func (u *UIController) OpenMenu(menuItems []util.MenuItem) {
	menu := NewBasicMenu(u)
	menu.SetItems(menuItems)
	menu.OnMouseMoved(
		u.mousePosInPixels.X,
		u.mousePosInPixels.Y,
	)
	u.PushWidget(menu)
}
func (u *UIController) GetRenderer() *renderer.TileRenderer {
	return u.tileRenderer
}

func (u *UIController) OpenTextWindow(buffer []util.LabelText) {
	if len(buffer) == 0 {
		return
	}
	textWindow := NewCenteredTextLabel(u)
	textWindow.SetCanBeClosed(true)
	textWindow.SetText(buffer)
	textWindow.SetBorderPadding(geometry.Point{X: 15, Y: 15})
	u.PushWidget(textWindow)
}
func (u *UIController) OpenInventoryOverlay(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
	inventory := NewIconOverlay(u, pos, tileSize, dataSource)
	inventory.SetShowEmptySlots(true)
	inventory.SetShowCounter(true)
	u.PushOverlay(inventory)
	return inventory
}

func (u *UIController) OpenInventoryModal(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
	inventory := NewIconOverlay(u, pos, tileSize, dataSource)
	inventory.SetShowEmptySlots(true)
	inventory.SetShowCounter(true)
	inventory.SetCanBeClosed(true)
	u.PushWidget(inventory)
	return inventory
}

func (u *UIController) OpenInventoryListModal(tileSize geometry.Point, dataSource ListDataSource) *IconList {
	inventory := NewIconList(u, tileSize, dataSource)
	u.PushWidget(inventory)
	return inventory
}

func (u *UIController) AddButton(offsetFromTopRightCorner, tileSize geometry.Point, icon int32, onClick func()) *ButtonOverlay {
	button := NewIconButton(u, offsetFromTopRightCorner, tileSize, icon)
	button.SetAction(onClick)
	u.PushOverlay(button)
	return button
}

func (u *UIController) SetUIBarWidth(uiBarWidth int) *ColoredRect {
	u.uiBarWidth = uiBarWidth
	background := NewColoredRect(u, uiBarWidth)
	background.SetFillColor(color.Black)
	u.PrependOverlay(background)
	return background
}

func (u *UIController) AddMinimap(offsetFromBottomRightCorner geometry.Point, maxHeight, availableWidth int, miniMapImage *ebiten.Image) *MinimapOverlay {
	minimap := NewMiniMap(u, offsetFromBottomRightCorner, maxHeight, availableWidth, miniMapImage)
	u.InsertOverlayAtIndex(1, minimap)
	return minimap
}

func (u *UIController) AddLabel(pos geometry.Point) *BasicLabel {
	label := NewTextLabel(u, pos)
	u.PushOverlay(label)
	return label
}

func (u *UIController) PushOverlay(widget UIWidget) {
	u.overlays = append(u.overlays, widget)
}

func (u *UIController) PrependOverlay(widget UIWidget) {
	u.overlays = append([]UIWidget{widget}, u.overlays...)
}

func (u *UIController) InsertOverlayAtIndex(index int, widget UIWidget) {
	u.overlays = append(u.overlays[:index], append([]UIWidget{widget}, u.overlays[index:]...)...)
}

func (u *UIController) ShowTooltipWithDelay(tip Tooltip) {
	u.handleTooltipWithDelay(tip, 0.5)
}

func (u *UIController) handleTooltipWithDelay(tooltip Tooltip, delay float64) {
	if tooltip.IsNull() {
		u.currentTooltip = nil
	} else {
		u.currentTooltip = tooltip
		u.ticksUntilTooltipAppears = util.SecondsToTicks(delay)
	}
}

func (u *UIController) HideTooltip() {
	u.currentTooltip = nil
}

func (u *UIController) PopAll() {
	u.modal = nil
}

func (u *UIController) SetTileMapAnimations(mapping map[int32]TilemapAnimInfo) {
	u.tilemapAnimations = mapping
}

func (u *UIController) SetAndInitTileRenderer(tileRenderer *renderer.TileRenderer) {
	u.tileRenderer = tileRenderer
	tileRenderer.SetColorFromName(defs.ColorFromName)
	tileRenderer.SetWhiteTile(defs.IconWhite)
	tileRenderer.SetDefaultColors(defs.DarkYellow, color.Black)
}

func (u *UIController) GetScreenSize() geometry.Point {
	return u.screenSize
}

func (u *UIController) OnScreenSizeChanged() {
	u.deviceDPIScale = ebiten.DeviceScaleFactor()
	newScreenSize := u.GetDeviceIndependentScreenSize()

	for _, overlay := range u.overlays {
		if overlay != nil {
			overlay.Layout()
		}
	}
	screenSizeForMap := newScreenSize
	if u.HasOverlays() {
		screenSizeForMap = newScreenSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
	}
	u.tileRenderer.SetDeviceScale(u.deviceDPIScale)
	u.mapRenderer.OnScreenSizeChanged(screenSizeForMap)
	u.miniMap.Layout()
	if u.setToExactFit {
		u.disableScrolling()
	}
}

func (u *UIController) ClearAllOverlays() {
	u.overlays = []UIWidget{}
}

func (u *UIController) HasOverlays() bool {
	return len(u.overlays) > 0
}

func (u *UIController) GetUIBarWidth() int {
	return u.uiBarWidth
}

func (u *UIController) OnKeyPressed(key ebiten.Key, shiftPressed bool) {
	if !u.IsUIModalVisible() {
		return
	}
	if keyListener, isKeylistener := u.GetTopModal().(Keylistener); isKeylistener {
		keyListener.OnKeyPressed(key, shiftPressed)
	}
}

func (u *UIController) ElevateOverlay(overlay UIWidget) {
	u.elevatedOverlay = overlay
}

func (u *UIController) StopElevation() {
	u.elevatedOverlay = nil
}

func (u *UIController) ParseColorCodedText(colorCodedText string) []util.ColoredTextPart {
	return u.tileRenderer.ParseColorCodedText(colorCodedText)
}

func (u *UIController) OpenStash(playerInv *common.Inventory, stashInTown *common.Inventory, drop func(item *common.Item)) {
	if u.inventory == nil {
		return // the vampire has no inventory panel to stash from
	}
	player := u.context.GameState().GetPlayer()
	toStashHandler := func(uiItem UIElement) {
		playerEquip := player.GetEquipment()
		itemStack := uiItem.(common.InventoryStack)
		firstItem := itemStack[0]
		if playerEquip.IsEquipped(firstItem) { // unequip first..
			u.context.PlayerAction().UnequipItem(firstItem)
		}
		playerInv.Remove(firstItem)
		if u.shiftIsBeingPressedThisTick {
			drop(firstItem)
		} else {
			stashInTown.Add(firstItem)
		}
	}

	dataSource := IconDataSource{
		IsHighlighted: func(element UIElement) bool {
			return false
		},
		GetElements: func() []UIElement {
			return sortStacksForUI(stashInTown.StackedItems())
		},
		OnSelection: func(uiItem UIElement) {
			itemStack := uiItem.(common.InventoryStack)
			firstItem := itemStack[0]
			stashInTown.Remove(firstItem)

			if u.shiftIsBeingPressedThisTick { // drop
				firstItem.DroppedByPlayer()
				u.context.PlayerAction().DropItem(firstItem)
			} else { // move
				playerInv.Add(firstItem)
			}
		},
	}

	tileSize := u.worldTexture.GetTileSize()

	screenSize := u.GetDeviceIndependentScreenSize()
	invBounds := u.inventory.GetBounds()
	paddingX := 20
	offsetFromBottomRightCorner := geometry.Point{X: screenSize.X - invBounds.Max.X + invBounds.Size().X + paddingX, Y: screenSize.Y - invBounds.Max.Y} //- invBounds.Size().X - paddingX, Y: invBounds.Min.Y}

	stashInv := u.OpenInventoryModal(offsetFromBottomRightCorner, tileSize, dataSource)
	u.ElevateOverlay(u.inventory)

	oldHandler := u.inventory.GetOnSelectedHandler()
	u.inventory.SetOnSelectedHandler(toStashHandler)

	onClose := func() { // revert inventory..
		u.StopElevation()
		u.inventory.SetOnSelectedHandler(oldHandler)
	}

	stashInv.SetOnCloseHandler(onClose)
}

// uiBarWidthInPixels is the width the sidebar used to inherit from the inventory
// grid that lived in it.
const uiBarWidthInPixels = 275

func (u *UIController) setupUI() {
	u.overlays = []UIWidget{}
	u.modal = nil

	paddingLeftAndRight := 6
	u.uiBarWidth = uiBarWidthInPixels
	u.SetUIBarWidth(u.uiBarWidth)
	xOffset := u.uiBarWidth - paddingLeftAndRight
	u.statsLabel = u.AddLabel(geometry.Point{X: xOffset, Y: paddingLeftAndRight})
	u.statsLabel.SetText(util.ToLabelText([]string{ // hacky way to measure the right bounding box..
		"Day 88, 88:88",
		"Blood: 100%",
		"60 fps / 60 tps",
	}))
	statsBounds := u.statsLabel.GetBounds()

	u.npcList = NewNPCList(u, geometry.Point{X: xOffset, Y: statsBounds.Max.Y + paddingLeftAndRight*2})
	u.PushOverlay(u.npcList)

	// The minimap sits at the foot of the bar, where the inventory grid was.
	maxHeightForMap := u.GetDeviceIndependentScreenSize().Y / 2
	u.miniMap = u.addMiniMap(geometry.Point{X: u.uiBarWidth, Y: paddingLeftAndRight}, maxHeightForMap, paddingLeftAndRight)

	u.OnScreenSizeChanged()
}

// OpenNPCWindow shows what the player knows about the Townsperson holding that
// number. A number nobody currently in view holds does nothing — the list on the
// right is the whole of what he can ask about.
func (u *UIController) OpenNPCWindow(number int) {
	for _, actor := range u.watchedTownspeople() {
		if actor.GetTownsperson().Number == number {
			u.PushWidget(NewNPCWindow(u, actor))
			return
		}
	}
}

// watchedTownspeople are the Townspeople the player can actually watch right
// now: on the part of the map being rendered, and inside his field of view.
func (u *UIController) watchedTownspeople() []*common.Actor {
	state := u.context.GameState()
	player := state.GetPlayer()
	fieldOfView := player.GetFOV()
	onScreen := u.mapRenderer.GetVisibleMap()

	var watched []*common.Actor
	for _, actor := range state.GetMap().Actors() {
		if actor == player || actor.GetTownsperson() == nil || !actor.IsVisible() {
			continue
		}
		if !onScreen.Contains(actor.MapPosition()) || !fieldOfView.Visible(actor.MapPosition()) {
			continue
		}
		watched = append(watched, actor)
	}
	slices.SortFunc(watched, func(a, b *common.Actor) int {
		return a.GetTownsperson().Number - b.GetTownsperson().Number
	})
	return watched
}

func (u *UIController) addInventoryOverlay(pixelEdgeFromBottomRight geometry.Point) *IconOverlay {
	state := u.context.GameState()
	isEquipped := func(uiItem UIElement) bool {
		player := state.GetPlayer()
		itemStack := uiItem.(common.InventoryStack)
		firstItem := itemStack[0]
		return player.GetEquipment().IsEquipped(firstItem)
	}

	dataSource := IconDataSource{
		IsHighlighted: isEquipped,
		GetElements: func() []UIElement {
			return sortStacksForUI(state.GetPlayer().GetInventory().StackedItems())
		},
		OnSelection: u.onItemInInventoryClicked,
	}
	tileSize := u.worldTexture.GetTileSize()
	return u.OpenInventoryOverlay(pixelEdgeFromBottomRight, tileSize, dataSource)
}

func firstItemFromUIStack(uiItem UIElement) *common.Item {
	var firstItem *common.Item

	if item, isItem := uiItem.(*common.Item); isItem {
		firstItem = item
	} else if textStack, isTextualStack := uiItem.(*common.TextualStack); isTextualStack {
		firstItem = textStack.InventoryStack[0]
	} else if itemStack, isItemStack := uiItem.(common.InventoryStack); isItemStack {
		firstItem = itemStack[0]
	}

	return firstItem
}

func (u *UIController) onItemInInventoryClicked(uiItem UIElement) {
	state := u.context.GameState()
	if !state.GetPlayer().IsAlive() {
		return
	}
	act := u.context.PlayerAction()
	// shift state
	u.shiftIsBeingPressedThisTick = ebiten.IsKeyPressed(ebiten.KeyShift)
	typedItem := firstItemFromUIStack(uiItem)

	player := state.GetPlayer()

	wasEquipped := player.GetEquipment().IsEquipped(typedItem)
	if u.shiftIsBeingPressedThisTick { // drop
		act.DropItem(typedItem)
		if wasEquipped {
			u.ShowFullStats()
		}
		return
	}

	// equip
	if typedItem.IsEquippable() {
		if wasEquipped {
			act.UnequipItem(typedItem)
		} else {
			act.EquipItem(typedItem)
		}
		u.ShowFullStats()
		return
	}

	// use
	contextActions := u.context.PlayerAction().GetContextActionsForItem(typedItem)
	if len(contextActions) == 0 {
		return // maybe open a description? or at least a menu with drop or sth?
	}

	if len(contextActions) == 1 {
		actionFunc := contextActions[0].Action
		actionFunc()
		// hmm, this will trigger even if unsuc
		return
	}
	u.OpenMenu(contextActions)
}
func (u *UIController) ShowFullStats() {
	u.miniMap.Hide()
	label := u.statsLabel

	player := u.context.GameState().GetPlayer()
	label.SetText(player.GetFullStats())
}
func (u *UIController) BeginPlayerSellingTo(vendor *common.Actor) {
	player := u.context.GameState().GetPlayer()
	playerInventory := player.GetInventory()
	msg := u.msg
	if playerInventory.IsEmpty() {
		msg("You have nothing to sell")
		return
	}
	var sellItems []ShopMenuItem
	textParser := u.tileRenderer.ParseColorCodedText

	for _, itemStack := range playerInventory.StackedItems() {
		typedItem := itemStack[0]
		sellPrice := typedItem.BuyPrice() / 2

		itemName := typedItem.Name()
		if u.configuration.ItemNamesInShop == common.ItemNamingSchemeEnchantments {
			itemName = typedItem.PlusName()
		} else if u.configuration.ItemNamesInShop == common.ItemNamingSchemeBandStyle {
			itemName = typedItem.BandStyleName()
		}
		sellItems = append(sellItems, ShopMenuItem{
			ShortCut: playerInventory.GetLetterForItemStack(itemStack),
			Icon:     typedItem.InventoryIcon(),
			Tooltip:  typedItem.GetTooltipLines(textParser),
			TextColumns: []util.LabelText{
				{
					Text:               itemName,
					TextWithColorCodes: textParser(itemName),
					UseColorCodes:      true,
				},
				{
					Text:               fmt.Sprintf("%d", sellPrice),
					TextWithColorCodes: textParser(fmt.Sprintf("%d", sellPrice)),
					UseColorCodes:      true,
				},
			},
			Action: func() {
				playerEquip := player.GetEquipment()
				if playerEquip.IsEquipped(typedItem) {
					u.context.PlayerAction().UnequipItem(typedItem)
				}
				playerInventory.Remove(typedItem)
				vendor.GetInventory().Add(typedItem)
				player.GetStats().IncrementBy(stats.Gold, sellPrice)
				//u.sendInventoryToWindowClient()
				msg(fmt.Sprintf("You sold [:darkText]%s[:white] for [:darkText]%d[:white] gold", typedItem.Name(), sellPrice))
				u.BeginPlayerSellingTo(vendor)
			},
		})
	}
	goldAmount := player.GetStats().GetAttribute(stats.Gold)
	menu := NewShopMenu(u, u.worldTexture.GetTileSize())
	u.PushWidget(menu)
	menu.SetHeader("You have:", fmt.Sprintf("%d gold", goldAmount))
	menu.SetItems(sellItems, []Alignment{AlignmentLeft, AlignmentRight})
}
func (u *UIController) addMiniMap(offset geometry.Point, maxHeight int, horizontalPadding int) *MinimapOverlay {
	miniMap := u.AddMinimap(offset.Sub(geometry.Point{X: horizontalPadding}), maxHeight, offset.X-(horizontalPadding*2), u.mapRenderer.GetMiniMap())
	miniMap.SetOnMapClickedHandler(u.onMapClicked)
	return miniMap
}

func (u *UIController) SetTTFFont(file io.ReaderAt, size float64) {
	tt, err := opentype.ParseReaderAt(file)
	if err != nil {
		println(err.Error())
		return
	}
	dpi := 72 * u.deviceDPIScale
	fontFace, faceErr := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if faceErr != nil {
		println(faceErr.Error())
	}
	//mplusBigFont = text.FaceWithLineHeight(mplusBigFont, 54) // adjust line height
	u.tileRenderer.SetTTF(fontFace)
}

func (u *UIController) OnMapChanged() {
	loadedMap := u.context.GameState().GetMap()

	gridSize := u.worldTexture.GetTileSize()
	mapWindow := renderer.NewMapWindow(
		u.GetDeviceIndependentScreenSize().Sub(geometry.Point{X: u.uiBarWidth, Y: 0}),
		geometry.Point{X: loadedMap.GetWidth(), Y: loadedMap.GetHeight()},
		gridSize,
		u.GetTileScaleFactor,
		u.StaticMapLookup)

	u.mapRenderer = renderer.NewMapRenderer(u.tileRenderer, mapWindow)
	u.mapRenderer.SetMiniMapColorHandler(u.getMiniMapColor)

	u.setupUI()

	u.OnPlayerTurnTaken()
}
func (u *UIController) getMiniMapColor(mapPos geometry.Point, tileColor *[4]byte) {
	state := u.context.GameState()
	currentMap := state.GetMap()
	if state.GetPlayer().MapPosition() == mapPos {
		*tileColor = [4]byte{100, 255, 100, 255}
		return
	}
	if !currentMap.IsExplored(mapPos) {
		*tileColor = [4]byte{10, 10, 10, 255}
		return
	}
	if currentMap.IsStairsAt(mapPos) {
		*tileColor = [4]byte{235, 52, 229, 255}
		return
	}
	if !state.GetPlayer().CanSee(mapPos) {
		if currentMap.IsSpecialAt(mapPos, gridmap.SpecialTileWater) {
			*tileColor = [4]byte{86 / 2, 112 / 2, 194 / 2, 255}
			return
		}
		if currentMap.IsWalkable(mapPos) {
			*tileColor = [4]byte{35, 30, 51, 255}
		} else {
			*tileColor = [4]byte{85, 95, 112, 255}
		}
		return
	}
	if currentMap.IsActorAt(mapPos) {
		actorAt := currentMap.ActorAt(mapPos)
		if actorAt.IsHostileTo(state.GetPlayer()) {
			*tileColor = [4]byte{255, 0, 0, 255}
			return
		}
		*tileColor = [4]byte{220, 120, 75, 255}
		return
	}

	if currentMap.IsSpecialAt(mapPos, gridmap.SpecialTileWater) {
		*tileColor = [4]byte{86, 112, 194, 255}
		return
	}

	if currentMap.IsWalkable(mapPos) {
		*tileColor = [4]byte{77, 60, 102, 255}
	} else {
		*tileColor = [4]byte{170, 191, 224, 255}
	}
}

func prepend[T any](slice []T, s T) []T {
	return append([]T{s}, slice...)
}
func (u *UIController) StaticMapLookup(x int, y int, tick uint64) []renderer.CellDrawInfo {
	mapPos := geometry.Point{X: x, Y: y}
	state := u.context.GameState()
	currentMap := state.GetMap()
	if currentMap == nil || !currentMap.Contains(mapPos) {
		return []renderer.CellDrawInfo{}
	}

	if !currentMap.IsExplored(mapPos) && !u.disableFogOfWar {
		return []renderer.CellDrawInfo{}
	}

	var drawColor color.RGBA = color.RGBA{255, 255, 255, 255}

	tileAt := currentMap.GetCell(mapPos).TileType
	worldIcon, worldTileColor := u.getTileIcon(tileAt)
	if tileAt.Special == gridmap.SpecialTileWater {
		worldTileColor = state.GetWaterColor()
	}

	// animate the worldIcon
	if u.renderingMode == GraphicRendering && (state.GetPlayer().CanSee(mapPos) || u.disableFogOfWar) {
		if !u.configuration.DrawTilesBehindActors && currentMap.IsActorAt(mapPos) {
			return []renderer.CellDrawInfo{}
		}
		if info, ok := u.tilemapAnimations[worldIcon]; ok {
			// if it is, then calculate the current frame based on the tick
			frameOffset := util.GetLoopingFrameFromTick(tick+1000-uint64(x*17)-uint64(y*19), info.DelayInSeconds, info.FrameCount)
			worldIcon += frameOffset
		}
	}
	// fake lighting
	if !u.disableFogOfWar {
		drawColor = state.ApplyLighting(mapPos, drawColor, tick)
		worldTileColor = state.ApplyLighting(mapPos, worldTileColor, tick)
	}

	var drawInfo []renderer.CellDrawInfo

	// 1. items
	if itemAt, isItemAt := currentMap.TryGetItemAt(mapPos); isItemAt && itemAt.IsDrawn() && (!itemAt.IsHidden() || u.context.GameState().GetFlags().IsSet(string(common.FlagHasTrueSight))) {
		itemIcon := itemAt.Icon(u.worldTicks)
		drawItem := renderer.CellDrawInfo{
			Atlas: u.worldTexture,
			Icon:  itemIcon,
			Color: drawColor,
		}
		drawInfo = prepend(drawInfo, drawItem)
		if !u.configuration.DrawTilesBehindItems {
			return drawInfo
		}
	}

	// 2. objects
	if currentMap.IsObjectAt(mapPos) {
		objectAt := currentMap.ObjectAt(mapPos)
		objIcon := objectAt.Icon(u.worldTicks)
		drawObject := renderer.CellDrawInfo{
			Atlas: u.worldTexture,
			Icon:  objIcon,
			Color: drawColor,
		}
		drawInfo = prepend(drawInfo, drawObject)
		if !u.configuration.DrawTilesBehindObjects {
			return drawInfo
		}
	}

	// 2. decals
	if decalIcon, exists := currentMap.GetDecal(mapPos); exists {
		drawInfo = prepend(drawInfo, renderer.CellDrawInfo{
			Atlas: u.worldTexture,
			Icon:  decalIcon,
			Color: color.White,
		})
	}

	// 1. world
	drawInfo = prepend(drawInfo, renderer.CellDrawInfo{
		Atlas: u.GetCurrentAtlas(),
		Icon:  worldIcon,
		Color: worldTileColor,
	})

	return drawInfo
}

func (u *UIController) GetDeviceIndependentScreenSize() geometry.Point {
	if u.isFullScreen {
		w, h := ebiten.ScreenSizeInFullscreen()
		return geometry.Point{X: w, Y: h}
	}
	return u.screenSize
}

func (u *UIController) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	panic("should use layoutf")
}

func (u *UIController) LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64) {
	//u.deviceDPIScale = ebiten.DeviceScaleFactor()
	intWidth := int(outsideWidth)
	intHeight := int(outsideHeight)
	if u.screenSize.X != intWidth || u.screenSize.Y != intHeight {
		u.screenSize = geometry.Point{X: intWidth, Y: intHeight}
		u.OnScreenSizeChanged()
	}
	return outsideWidth * u.deviceDPIScale, outsideHeight * u.deviceDPIScale
}

func (u *UIController) GetTileScaleFactor() float64 {
	return u.tileScale
}

func (u *UIController) Cancel() {
	if u.IsUIModalVisible() {
		u.OnCommand(PlayerCommandCancel)
		return
	}
	if u.inventory != nil && u.inventory.IsActive() {
		u.inventory.Deactivate()
		return
	}
}

func (u *UIController) ShowGameOverMenu() {
	u.OpenMenu([]util.MenuItem{
		{
			MainText: "Try again",
			Action: func() {
				u.context.ResetGame(u.context.GameState().SelectedClass)
			},
		},
		{
			MainText: "Change class",
			Action: func() {
				u.OpenClassSelectionMenu()
			},
		},
		{
			MainText: "Quit",
			Action: func() {
				u.shouldQuitGame = true
			},
		},
	})
}

func (u *UIController) OpenPlayerStashInTown() {
	state := u.context.GameState()
	gui := u
	stashInTown := state.GetPlayerStashInTown()
	playerInv := state.GetPlayer().GetInventory()
	//playerEquip := g.state.GetPlayer().GetEquipment()

	// NEW
	onDrop := func(item *common.Item) {
		item.DroppedByPlayer()
		state.AddItemToMapWithDistribution(item, state.GetPlayer().MapPosition())
	}
	gui.OpenStash(playerInv, stashInTown, onDrop)
	// on close of the stash modal, restore the onclick handler of the inventory..
}

func (u *UIController) OpenClassSelectionMenu() {
	var items []util.MenuItem
	for _, c := range u.context.Definitions.ClassDefinitions {
		classDef := c
		items = append(items, util.MenuItem{
			MainText: classDef.Name,
			Action: func() {
				u.context.ResetGame(classDef)
			},
		})
	}
	u.OpenMenu(items)
}
func (u *UIController) OnPlayerTurnTaken() {
	// Called once per tick now that the world runs in real time, so nothing in
	// here may cost a frame. Everything that is purely a function of where the
	// player stands runs on the step, not on the tick.
	player := u.context.GameState().GetPlayer()
	playerPos := player.MapPosition()
	playerMoved := playerPos != u.lastPlayerTile
	u.lastPlayerTile = playerPos

	if playerMoved {
		// Fog of war and the player's own dot only change when he does.
		u.UpdateMiniMap()

		// Scrolling the map back under the player is a function of his tile and
		// the visible rect. A resize is handled by OnScreenSizeChanged.
		if (!u.setToExactFit && !u.configuration.KeepPlayerCentered) || u.configuration.KeepPlayerCentered {
			distFromBorder := 4
			visibleMap := u.mapRenderer.GetVisibleMap().Shift(distFromBorder, distFromBorder, -distFromBorder, -distFromBorder)
			if !visibleMap.Contains(playerPos) {
				u.mapRenderer.CenterOnFloat(player.DrawPosition())
			}
		}
	}

	// The clock and the Townspeople's dots move whether the player does or not,
	// so these two still need a heartbeat — just not one per frame. Six times a
	// second is under one animation frame of staleness.
	// ponytail: fixed divisor. Drive it off the clock's minute if it ever shows.
	if u.worldTicks%10 == 0 {
		u.UpdateStatsLabel()
		if !playerMoved {
			u.UpdateMiniMap()
		}
	}
}

func (u *UIController) UpdateStatsLabel() {
	state := u.context.GameState()
	label := u.statsLabel
	clock := state.GetClock()
	bloodText := ""
	if u.context.Blood != nil {
		if blood := u.context.Blood(); blood != nil {
			bloodText = blood.String()
		}
	}
	label.SetText([]util.LabelText{
		{
			Text:      clock.String(),
			TextColor: defs.DarkTextColor,
		},
		{
			Text:      bloodText,
			TextColor: defs.DarkTextColor,
		},
		{
			Text:      fmt.Sprintf("%.0f fps / %.0f tps", ebiten.ActualFPS(), ebiten.ActualTPS()),
			TextColor: defs.DarkTextColor,
		},
	})
	u.miniMap.Show()
}

func (u *UIController) showHelpText() {
	textColor := defs.DarkTextColor
	lines := []string{
		"You are a vampire, living unseen in this town.",
		"",
		"Arrow keys or numpad to move.",
		"'space' to pause and unpause time.",
		"'.' to wait until dawn, ',' to wait until dusk.",
		"'e' to drain someone standing next to you.",
		"'h' to hide a body you are standing over.",
		"'end' to leave the town for good, whenever you choose.",
		"'1' to '6' to look up whoever is listed under that number on the right.",
		"'x' to centre the view on yourself.",
		"'l' to show the last lines of the message log.",
		"'c' to show your stats.",
		"'m' and 'n' to zoom the map.",
		"'esc' to open the menu.",
		"",
		"Time passes on its own. The townspeople keep their own hours",
		"whether or not you are watching.",
	}
	u.OpenTextWindow(util.ToLabelTextWithColor(lines, textColor))
}

func (u *UIController) ShowLog() {
	buffer := u.context.GetLog()
	if len(buffer) > 20 {
		buffer = buffer[len(buffer)-20:]
	}
	u.OpenTextWindow(toGradientLabelText(u.tileRenderer.ParseColorCodedText, buffer))
	//logWindow.ScrollToBottom()
}

func (u *UIController) OpenListInventory() {
	noFilter := func(element *common.Item) bool { return true }
	u.OpenListInventoryForSelection(noFilter, func(invItemKey rune) {
		inv := u.context.GameState().GetPlayer().GetInventory()
		u.onItemInInventoryClicked(inv.GetItemByLetter(invItemKey))
	})
}

func (u *UIController) FlashScreen(color color.Color, seconds float64) {
	u.globalScaleColor = color
	u.globalFlashColorTicksLeft = int(util.SecondsToTicks(seconds))
	u.globalFlashColorTicksStart = int(util.SecondsToTicks(seconds))
}

func (u *UIController) OpenListInventoryForSelection(filter func(element *common.Item) bool, onSelected func(invItemKey rune)) {
	dataSource := ListDataSource{
		GetElements: func() []IconListElement {
			inventory := u.context.GameState().GetPlayer().GetInventory()
			typed := common.SortStacksForTextUITyped(inventory.StackedItemsWithFilter(filter), inventory.GetLetterForItemStack, onSelected)
			var result []IconListElement
			for _, item := range typed {
				result = append(result, item)
			}
			return result
		},
	}
	u.PopAll()
	u.OpenInventoryListModal(u.worldTexture.GetTileSize(), dataSource)
}

type ColorCodeParser func(draw string) []util.ColoredTextPart

func toGradientLabelText(textParser ColorCodeParser, buffer []string) []util.LabelText {
	// we want each item to be less white than the one before
	intensityFromIndex := func(index int) uint8 {
		maxIntensity := uint8(255)
		minIntensity := uint8(100)
		maxIndex := len(buffer) - 1
		percentage := 1.0 - (float64(index) / float64(maxIndex))
		return uint8(float64(maxIntensity) - (float64(maxIntensity)-float64(minIntensity))*percentage)
	}
	var result []util.LabelText
	for i, line := range buffer {
		fadedBaseColor := color.RGBA{R: intensityFromIndex(i), G: intensityFromIndex(i), B: intensityFromIndex(i), A: 255}
		baseColorCode := defs.RGBAToColorCode(fadedBaseColor)
		result = append(result, util.LabelText{
			UseColorCodes:      true,
			TextWithColorCodes: textParser(baseColorCode + line),
			Text:               common.RemoveColorCodes(line),
			TextColor:          fadedBaseColor,
		})
	}
	return result
}

func sortStacksForUI(stacks []common.InventoryStack) []UIElement {
	items := make([]UIElement, 0)
	common.SortInventory(stacks)
	for _, item := range stacks {
		items = append(items, item)
	}
	return items
}

func (u *UIController) OpenShopMenu(vendor *common.Actor) {
	u.openShopMenuWithHandler(vendor.GetInventory(), func(item *common.Item) {
		u.tryToBuy(item, vendor)
	})
}
func (u *UIController) openShopMenuWithHandler(inventory *common.Inventory, handler func(item *common.Item)) {
	if inventory.IsEmpty() {
		return
	}
	player := u.context.GameState().GetPlayer()
	goldAmount := player.GetStats().GetAttribute(stats.Gold)

	var menuItems []ShopMenuItem
	inventoryStacks := inventory.StackedItems()
	common.SortInventory(inventoryStacks)
	textParser := u.ParseColorCodedText

	for itemIndex, i := range inventoryStacks {

		item := i[0]
		cost := item.BuyPrice()
		cannotPayForItem := cost > goldAmount

		colorCodeForPrice := "darkText"
		if cannotPayForItem {
			colorCodeForPrice = "negative"
		}
		itemName := item.Name()
		if u.configuration.ItemNamesInShop == common.ItemNamingSchemeEnchantments {
			itemName = item.PlusName()
		} else if u.configuration.ItemNamesInShop == common.ItemNamingSchemeBandStyle {
			itemName = item.BandStyleName()
		}
		if len(i) > 1 {
			itemName = fmt.Sprintf("%s (x%d)", itemName, len(i))
		}
		nameString := fmt.Sprintf("%s", itemName)
		//nameStringWithColor := fmt.Sprintf("%s", itemName)
		priceString := fmt.Sprintf("%d", cost)
		priceStringWithColor := fmt.Sprintf("[:%s]%d", colorCodeForPrice, cost)
		// shortcuts start with 0 => 'a'
		shortCut := rune('a') + rune(itemIndex)

		menuItems = append(menuItems, ShopMenuItem{
			ShortCut: shortCut,
			Tooltip:  item.GetTooltipLines(textParser),
			Icon:     item.InventoryIcon(),
			TextColumns: []util.LabelText{
				{
					Text: nameString,
					//TextWithColorCodes: textParser(nameStringWithColor),
					UseColorCodes: false,
					TextColor:     item.TintColor(),
				},
				{
					Text:               priceString,
					TextWithColorCodes: textParser(priceStringWithColor),
					UseColorCodes:      true,
				},
			},
			Action: func() {
				handler(item)
			},
		})
	}

	menu := NewShopMenu(u, u.worldTexture.GetTileSize())
	u.PushWidget(menu)
	menu.SetHeader("You have:", fmt.Sprintf("%d gold", goldAmount))
	menu.SetItems(menuItems, []Alignment{AlignmentLeft, AlignmentRight})
}

func (u *UIController) tryToBuy(item *common.Item, seller *common.Actor) {
	player := u.context.GameState().GetPlayer()
	goldAmount := player.GetStats().GetAttribute(stats.Gold)
	msg := u.msg
	cost := item.BuyPrice()
	cannotPayForItem := cost > goldAmount

	if cannotPayForItem {
		msg("You don't have enough gold")
		u.OpenShopMenu(seller)
		return
	}
	player.GetStats().DecrementBy(stats.Gold, cost)
	seller.GetStats().IncrementBy(stats.Gold, cost)

	alwaysItems := u.context.Definitions.VendorDefinitions[string(seller.GetVendorType())].Always
	seller.GetInventory().Remove(item)
	player.GetInventory().Add(item)
	itemName := item.GetInternalName()
	if slices.Contains(alwaysItems, itemName) {
		seller.GetInventory().Add(item.Clone())
	}

	msg(fmt.Sprintf("You bought [:darkText]%s[:white] for [:darkText]%d[:white] gold", item.Name(), cost))
	u.OpenShopMenu(seller)
}

func (u *UIController) showFlags() {
	state := u.context.GameState()
	asStrings := state.GetFlags().AsStrings()
	u.OpenTextWindow(util.ToLabelText(asStrings))
	//logWindow.ScrollToBottom()
}

func (u *UIController) setTileScale(scale float64) {
	u.tileScale = scale
	u.tileRenderer.SetTileScale(scale)
}

func (u *UIController) UpdateMiniMap() {
	u.mapRenderer.UpdateMiniMap()
}

func (u *UIController) SetPlayerMessage(message func(message string)) {
	u.msg = message
}

func (u *UIController) drawHealthBar(actor *common.Actor) {
	health := actor.GetStats().GetResource(stats.Health).Value()
	maxHealth := actor.GetStats().GetAttribute(stats.MaxHealth)
	healthPercent := float64(health) / float64(maxHealth)

	gridSize := u.worldTexture.GetTileSize()
	tileScale := u.GetTileScaleFactor()
	pos := actor.DrawPosition()
	screenPos := u.mapRenderer.MapFloatToScreen(pos)
	healthBarHeight := 2
	healthBarWidth := float64(gridSize.X) * tileScale

	healthBarX := screenPos.X
	healthBarY := float64(screenPos.Y) + (float64(gridSize.Y) * tileScale) - float64(healthBarHeight)

	healthBarColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	backgroundColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	if actor != u.context.GameState().GetPlayer() {
		backgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
		healthBarColor = redToGreen(healthPercent)
	}

	topLeft := geometry.Point{X: int(healthBarX), Y: int(healthBarY)}
	size := geometry.Point{X: int(healthBarWidth), Y: healthBarHeight}
	barSize := geometry.Point{X: int(healthBarWidth * healthPercent), Y: healthBarHeight}
	u.tileRenderer.DrawColoredRect(topLeft, size, backgroundColor)
	u.tileRenderer.DrawColoredRect(topLeft, barSize, healthBarColor)
}

func (u *UIController) GetCurrentAtlas() renderer.TextureAtlas {
	switch u.renderingMode {
	case UnicodeRendering:
		fallthrough
	case CodePage437Rendering:
		fallthrough
	case AsciiRendering:
		return u.asciiTexture
	default:
		return u.worldTexture
	}
}

func (u *UIController) SetAsciiAtlas(atlas renderer.TextureAtlas) {
	u.asciiTexture = atlas
}

func redToGreen(percent float64) color.RGBA {
	// 0 -> red, 0.5 -> yellow, 1 -> green
	red := uint8(255 * (1 - percent))
	green := uint8(255 * percent)
	return color.RGBA{R: red, G: green, B: 0, A: 255}
}
