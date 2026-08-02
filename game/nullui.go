package game

import (
	"rpg/common"
	"rpg/geometry"
	"rpg/util"
)

// NullUI satisfies the user interface without drawing anything, so the game can
// be constructed and advanced with no graphical subsystem — by the headless
// tests, and by the tools that run the game without a window.
type NullUI struct {
}

func (d NullUI) FadeToBlack() {

}

func (d NullUI) FadeFromBlack() {

}

func (d NullUI) SetDebugActions(actions func() common.DebugActions) {

}

func (d NullUI) GetVisibleMap() geometry.Rect {
	//TODO implement me
	return geometry.NewRect(0, 0, 0, 0)
}

func (d NullUI) HasOverlays() bool {
	//TODO implement me
	return true
}

func (d NullUI) ShowLogMessage(message string) {
	//TODO implement me
	println(message)
}

func (d NullUI) Cancel() {
	//TODO implement me

}

func (d NullUI) PopAll() {
	//TODO implement me

}

func (d NullUI) ShowGameOverMenu() {
	//TODO implement me

}

func (d NullUI) OpenMenu(options []util.MenuItem) {
	//TODO implement me

}

func (d NullUI) OpenTextWindow(msg []util.LabelText) {
	//TODO implement me

}

func (d NullUI) ShowFullStats() {
	//TODO implement me

}

func (d NullUI) UpdateStatsLabel() {
	//TODO implement me

}

func (d NullUI) OpenStash(inv *common.Inventory, town *common.Inventory, drop func(item *common.Item)) {
	//TODO implement me

}

func (d NullUI) OnMapChanged() {
	//TODO implement me

}

func (d NullUI) OpenClassSelectionMenu() {
	//TODO implement me

}

func (d NullUI) Say(actor *common.Actor, linesOfText []string) {
	//TODO implement me

}

func (d NullUI) OpenShopMenu(npc *common.Actor) {
	//TODO implement me

}

func (d NullUI) BeginPlayerSellingTo(vendor *common.Actor) {
	//TODO implement me

}

func (d NullUI) OpenListInventoryForSelection(filter func(element *common.Item) bool, selected func(itemKey rune)) {
	//TODO implement me

}

func (d NullUI) UpdateMiniMap() {
	//TODO implement me

}

func (d NullUI) OpenPlayerStashInTown() {
	//TODO implement me

}

func (d NullUI) OnPlayerTurnTaken() {
	//TODO implement me

}

func (d NullUI) SetPlayerMessage(message func(message string)) {
	//TODO implement me

}

func NewNullUI() *NullUI {
	return &NullUI{}
}
