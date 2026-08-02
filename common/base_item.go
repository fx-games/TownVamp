package common

import (
	"cmp"
	"fmt"
	"image/color"
	"rpg/defs"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
	"slices"
)

type ItemRarity int

const (
	ItemRarityCommon ItemRarity = iota
	ItemRarityMagic
	ItemRarityUnique
)

type ArmorStats struct {
	BaseArmor  int
	Type       defs.ArmorType
	Durability *stats.Resource
	Plus       int
}

func (s ArmorStats) ToString(player *Actor) []util.LabelText {
	return []util.LabelText{
		util.LabelText{
			Text:      fmt.Sprintf("Armor Type: %s", s.Type.ToString()),
			TextColor: defs.OffWhite,
		},
		util.LabelText{
			Text:      fmt.Sprintf("Armor Class: %d", s.BaseArmor),
			TextColor: defs.OffWhite,
		},
		util.LabelText{
			Text:      fmt.Sprintf("Enchantment: +%d", s.Plus),
			TextColor: defs.OffWhite,
		},
		util.LabelText{
			Text:      fmt.Sprintf("Durability: %s", s.Durability.ToString()),
			TextColor: defs.OffWhite,
		},
	}
}

func (s ArmorStats) NameParts() string {
	// angband style [3, +3]
	return fmt.Sprintf("[%d, +%d]", s.BaseArmor, s.Plus)
}

func (s ArmorStats) GetTotalArmorClass() int {
	return s.BaseArmor + s.Plus
}

func (s ArmorStats) Clone() *ArmorStats {
	return &ArmorStats{
		BaseArmor:  s.BaseArmor,
		Type:       s.Type,
		Durability: stats.NewResource(s.Durability.Min(), s.Durability.GetMax, s.Durability.Value()),
		Plus:       s.Plus,
	}
}

func (s ArmorStats) CanBeFurtherEnchanted() bool {
	return s.Plus < 3
}

type LightStats struct {
	LightColor        color.RGBA
	FlickerPattern    []float64
	FlickerFrameDelay float64
}

func (s LightStats) Clone() *LightStats {
	return &LightStats{
		LightColor:        s.LightColor,
		FlickerPattern:    s.FlickerPattern,
		FlickerFrameDelay: s.FlickerFrameDelay,
	}
}

type Item struct {
	name string

	pos  geometry.Point
	icon int32

	armorComponent  *ArmorStats
	lightComponent  *LightStats

	qualityLevel      int
	equippedModifiers []*stats.Modifier
	getPlayer         func() *Actor
	category          defs.ItemCategory
	rarity            ItemRarity

	droppedByPlayer bool
	isHidden        bool
	shouldBeDrawn   bool
	baseCost        int

	forcedTooltip     []util.LabelText
	itemHooks         []ItemActionHook
	actionCharges     int
	destroyAfterUse   bool
	nameOnlyInTooltip bool
	internalName      string
	isQuestItem       bool
}

func NewItem(getPlayerRef func() *Actor, name string, icon int32) *Item {
	return &Item{
		getPlayer:     getPlayerRef,
		name:          name,
		icon:          icon,
		shouldBeDrawn: true,
	}
}

func (i *Item) SetDrawnToScreen(shouldBeDrawn bool) {
	i.shouldBeDrawn = shouldBeDrawn
}

func (i *Item) IsDrawn() bool {
	return i.shouldBeDrawn
}
func (i *Item) IsModifyingLight() bool {
	for _, modifier := range i.equippedModifiers {
		if modifier.Attribute == stats.LightRadius {
			return true
		}
	}
	return false
}

func (i *Item) SetIcon(icon int32) {
	i.icon = icon
}
func (i *Item) SetName(name string) {
	i.name = name
}

func (i *Item) Hide() {
	i.isHidden = true
}
func (i *Item) Show() {
	i.isHidden = false
}
func (i *Item) IsHidden() bool {
	return i.isHidden
}
func (i *Item) DroppedByPlayer() {
	i.droppedByPlayer = true
}

func (i *Item) IsDroppedByPlayer() bool {
	return i.droppedByPlayer
}
func (i *Item) GetRarity() ItemRarity {
	return i.rarity
}
func (i *Item) SetRarity(rarity ItemRarity) {
	i.rarity = rarity
}
func (i *Item) GetItemType() defs.ItemCategory {
	return i.category
}
func (i *Item) GetEquippedModifiers() []*stats.Modifier {
	return i.equippedModifiers
}
func (i *Item) Counter() int {
	return -1
}
func (i *Item) BuyPrice() int {
	itemPrice := i.baseCost
	if i.IsArmor() {
		pricePerPlus := 35
		plusValue := i.armorComponent.Plus * pricePerPlus
		itemPrice += plusValue
	}
	chargeMultiplier := 20 // for infinite charges
	if !i.HasInfiniteCharges() {
		chargeMultiplier = i.GetCharges()
	}
	_ = chargeMultiplier
	if i.HasModifiers() {
		for _, m := range i.equippedModifiers {
			itemPrice += m.GetMonetaryValue()
		}
	}
	return itemPrice
}
func (i *Item) InventoryIcon() int32 {
	return i.icon
}

func (i *Item) GetLightStats() *LightStats {
	return i.lightComponent
}

func (i *Item) TintColor() color.RGBA {
	switch i.rarity {
	case ItemRarityCommon:
		return i.categoryColor()
/*	case ItemRarityMagic:
		return defs.MagicHighlightColor
 */
	case ItemRarityUnique:
		return defs.InvColorUnique
	}
	return defs.InvColorMisc
}

func (i *Item) categoryColor() color.RGBA {
	if i.IsLightSource() {
		return defs.InvColorLightSource
	}

	if i.IsRing() {
		return defs.InvColorRing
	}

	if i.IsAmulet() {
		return defs.InvColorAmulet
	}

	if i.IsPotion() {
		return defs.InvColorPotion
	}

	if i.IsScroll() {
		return defs.InvColorScroll
	}

	if i.IsWand() {
		return defs.InvColorWand
	}

	if i.IsArmor() {
		return defs.InvColorArmor
	}

	if i.IsRangedWeapon() {
		return defs.InvColorRangedWeapon
	}

	if i.IsMeleeWeapon() {
		return defs.InvColorMeleeWeapon
	}

	return defs.InvColorMisc
}

func (i *Item) Name() string {
	name := i.name
	return name
}
func (i *Item) String() string {
	return i.Name()
}

// The weapon, throwable and skill components of an item went with the combat
// systems. These report false so the equipment and inventory code — which is
// mothballed rather than deleted — still compiles and behaves sanely.
func (i *Item) IsWeapon() bool {
	return false
}

func (i *Item) IsMeleeWeapon() bool {
	return false
}

func (i *Item) IsRangedWeapon() bool {
	return false
}

func (i *Item) IsThrowable() bool {
	return false
}

func (i *Item) IsArmor() bool {
	return i.armorComponent != nil
}

func (i *Item) SlotName() defs.SlotName {
	if i.IsLightSource() {
		return defs.SlotNameLightSource
	}
	if i.IsArmor() {
		return i.armorComponent.Type.GetInventorySlot()
	}
	if i.IsRing() {
		return defs.SlotNameRing
	}
	if i.IsAmulet() {
		return defs.SlotNameAmulet
	}
	return defs.SlotNameNotEquippable
}

func (i *Item) MapPosition() geometry.Point {
	return i.pos
}

func (i *Item) Icon(tick uint64) int32 {
	return i.icon
}

func (i *Item) SetMapPosition(pos geometry.Point) {
	i.pos = pos
}

func (i *Item) GetArmorStats() *ArmorStats {
	return i.armorComponent
}

func (i *Item) SetQualityLevel(level int) {
	i.qualityLevel = level
}

func (i *Item) SetEquippedModifiers(modifiers []*stats.Modifier) {
	i.equippedModifiers = modifiers
	i.sortModifiers()
}
func (i *Item) sortModifiers() {
	slices.SortStableFunc(i.equippedModifiers, func(i, j *stats.Modifier) int {
		return cmp.Compare(i.Attribute, j.Attribute)
	})
}

func (i *Item) RemoveEquippedModifier(modifier *stats.Modifier) {
	for index, m := range i.equippedModifiers {
		if m == modifier {
			i.equippedModifiers = append(i.equippedModifiers[:index], i.equippedModifiers[index+1:]...)
			return
		}
	}
}

func (i *Item) AddEquippedModifier(modifier *stats.Modifier) {
	i.equippedModifiers = append(i.equippedModifiers, modifier)
	i.sortModifiers()
}
func (i *Item) SetLightComponent(component *LightStats) {
	i.lightComponent = component
}
func (i *Item) SetArmorComponent(component *ArmorStats) {
	i.armorComponent = component
}

func (i *Item) GetArmorStatsAsString() []util.LabelText {
	return i.GetArmorStats().ToString(i.getPlayer())
}

func (i *Item) SetCategory(category defs.ItemCategory) {
	i.category = category
}

func (i *Item) IsLightSource() bool {
	return i.lightComponent != nil
}

func (i *Item) SetBaseCost(cost int) {
	i.baseCost = cost
}

func (i *Item) Clone() *Item {
	clone := &Item{
		name:              i.name,
		internalName: 	i.internalName,
		pos:               i.pos,
		icon:              i.icon,
		armorComponent:    nil,
		lightComponent:    nil,
		qualityLevel:      i.qualityLevel,
		equippedModifiers: i.equippedModifiers,
		getPlayer:         i.getPlayer,
		category:          i.category,
		rarity:            i.rarity,
		droppedByPlayer:   false,
		isHidden:          i.isHidden,
		shouldBeDrawn:     i.shouldBeDrawn,
		baseCost:          i.baseCost,
	}
	if i.armorComponent != nil {
		clone.armorComponent = i.armorComponent.Clone()
	}
	if i.lightComponent != nil {
		clone.lightComponent = i.lightComponent.Clone()
	}

	clone.forcedTooltip = i.forcedTooltip
	clone.itemHooks = i.itemHooks
	clone.actionCharges = i.actionCharges
	clone.destroyAfterUse = i.destroyAfterUse
	clone.nameOnlyInTooltip = i.nameOnlyInTooltip

	return clone
}

func (i *Item) SetCharges(charges int) {
	i.actionCharges = charges
}
func (i *Item) CanStackWith(other *Item) bool {
	if i.SlotName() != other.SlotName() {
		return false
	}

	if i.SlotName() != defs.SlotNameNotEquippable && i.SlotName() != defs.SlotNameQuiver {
		return false
	}

	if i.GetItemType() != other.GetItemType() {
		return false
	}

	if i.Name() != other.Name() {
		return false
	}

	if i.InventoryIcon() != other.InventoryIcon() {
		return false
	}

	if i.GetCharges() != other.GetCharges() {
		return false
	}

	return true
}

func (i *Item) SetUnlimitedCharges() {
	i.actionCharges = -1
}
func (i *Item) GetCharges() int {
	return i.actionCharges
}

func (i *Item) HasAction() bool {
	return false
}

func (i *Item) GetItemHooks() []ItemActionHook {
	return i.itemHooks
}
func (i *Item) GetTooltipLines(func(string) []util.ColoredTextPart) []util.LabelText {
	info := []util.LabelText{util.LabelTextFromStringWithColor(i.Name(), defs.OffWhite)}

	if i.nameOnlyInTooltip {
		return info
	}

	if len(i.forcedTooltip) > 0 {
		return append(info, i.forcedTooltip...)
	}

	if i.IsWeapon() {
		handedNess := "One-Handed"
		if i.SlotName() == defs.SlotNameTwoHandedWeapon {
			handedNess = "Two-Handed"
		}
		mrType := "melee"
		useText := []util.LabelText{util.LabelTextFromString("Move into an enemy to attack")}
		if i.IsRangedWeapon() {
			mrType = "ranged"
			useText = []util.LabelText{util.LabelTextFromString("Press \"f\" to aim and fire"), util.LabelTextFromString("Press \"q\" for quickfire")}
		}
		info = append(info, util.LabelTextFromString(fmt.Sprintf("%s %s weapon", handedNess, mrType)))
		info = append(info, util.LabelTextFromString(""))
		info = append(info, useText...)
		info = append(info, util.LabelTextFromString(""))
	}

	if i.IsArmor() {
		info = append(info, util.LabelTextFromString(""))
		info = append(info, i.GetArmorStatsAsString()...)
	}

	if len(i.GetEquippedModifiers()) > 0 {
		info = append(info, util.LabelTextFromString(""))
		for _, m := range i.GetEquippedModifiers() {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("%s", m.ToString())))
		}
	}
	if len(i.GetItemHooks()) > 0 {
		info = append(info, util.LabelTextFromString(""), util.LabelTextFromString("Effects"))
		for _, h := range i.GetItemHooks() {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("%s", h.ToString())))
		}
	}
	if i.HasAction() { // wands
		if i.HasInfiniteCharges() {
			//info = append(info, common.LabelTextFromString(fmt.Sprintf("Has infinite charges")))
		} else {
			if i.SelfDestructs() {
				info = append(info, util.LabelTextFromString(fmt.Sprintf("Is destroyed after use")))
			}

		}

	}

	if i.IsLightSource() {
		info = append(info, util.LabelTextFromString(""))
		if i.HasInfiniteCharges() {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("Provides light")))
		} else {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("Provides light for %d turns", i.GetCharges())))
		}
	}

	if i.IsThrowable() {
		info = append(info, util.LabelTextFromString(""))
		info = append(info, util.LabelTextFromString(fmt.Sprintf("Can be thrown")))
		info = append(info, util.LabelTextFromString(fmt.Sprintf("Press \"v\" to throw an item")))
		info = append(info, util.LabelTextFromString(""))
	}

	return info
}
func (i *Item) GetAllEnchantments() []Enchantment {
	var result []Enchantment
	for _, m := range i.GetEquippedModifiers() {
		result = append(result, m)
	}
	for _, h := range i.GetItemHooks() {
		result = append(result, h)
	}
	return result
}

func (i *Item) RemoveEnchantment(enchantment Enchantment) {
	switch e := enchantment.(type) {
	case *stats.Modifier:
		i.RemoveEquippedModifier(e)
	case ItemActionHook:
		i.RemoveItemHook(e)
	}
}

func (i *Item) AddEnchantment(enchantment Enchantment) {
	switch e := enchantment.(type) {
	case *stats.Modifier:
		i.AddEquippedModifier(e)
	case ItemActionHook:
		i.AddItemhook(e)
	}
}

func (i *Item) RemoveItemHook(hookToRemove ItemActionHook) {
	for index, hook := range i.itemHooks {
		if hook == hookToRemove {
			i.itemHooks = append(i.itemHooks[:index], i.itemHooks[index+1:]...)
			return
		}
	}
}

func (i *Item) AddItemhook(hookToAdd ItemActionHook) {
	i.itemHooks = append(i.itemHooks, hookToAdd)
}

func (i *Item) IsEquippable() bool {
	return i.IsArmor() || i.IsWeapon() || i.IsLightSource() || i.IsAccessory() || i.IsThrowable()
}

func (i *Item) HasTargetedAction() bool {
	return false
}

func (i *Item) HasUntargetedAction() bool {
	return false
}

func (i *Item) UseCharge() {
	if i.actionCharges > 0 {
		i.actionCharges--
	}
}

func (i *Item) HasInfiniteCharges() bool {
	return i.actionCharges == -1
}

func (i *Item) SelfDestructs() bool {
	return i.destroyAfterUse
}

func (i *Item) SetItemHooks(hooks []ItemActionHook) {
	i.itemHooks = hooks
}

func (i *Item) ShowNameOnlyInTooltip() {
	i.nameOnlyInTooltip = true
}
func (i *Item) SetTooltip(forcedTooltip []util.LabelText) {
	i.forcedTooltip = forcedTooltip
}

func (i *Item) HasChargesLeft() bool {
	return i.actionCharges != 0
}

func flatFunc(i int) func() int {
	return func() int {
		return i
	}
}

func (i *Item) BandStyleName() string {
	name := i.Name()
	if i.IsArmor() {
		armorValues := i.GetArmorStats().NameParts()
		name = fmt.Sprintf("%s %s", name, armorValues)
	}

	if i.IsLightSource() && !i.HasInfiniteCharges() {
		name = fmt.Sprintf("%s (%d turns)", name, i.GetCharges())
	}

	return name
}

func (i *Item) PlusName() string {
	name := i.Name()
	if i.IsArmor() {
		armorPlus := fmt.Sprintf("+%d", i.GetArmorStats().Plus)
		return fmt.Sprintf("%s %s", name, armorPlus)
	}
	return name
}

func (i *Item) SetDestroyAfterUse(value bool) {
	i.destroyAfterUse = value
}

func (i *Item) IsRing() bool {
	return i.GetItemType() == defs.ItemCategoryRing
}

func (i *Item) IsAmulet() bool {
	return i.GetItemType() == defs.ItemCategoryAmulet
}

func (i *Item) IsAccessory() bool {
	return i.IsRing() || i.IsAmulet()
}
func (i *Item) IsPotion() bool {
	return i.GetItemType() == defs.ItemCategoryPotion
}

func (i *Item) IsScroll() bool {
	return i.GetItemType() == defs.ItemCategoryScroll
}

func (i *Item) IsWand() bool {
	return i.GetItemType() == defs.ItemCategoryWand
}

func (i *Item) IsTwoHandedWeapon() bool {
	return i.IsWeapon() && i.SlotName() == defs.SlotNameTwoHandedWeapon
}

func (i *Item) GetInternalName() string {
	return i.internalName
}

func (i *Item) SetInternalName(name string) {
	i.internalName = name
}

func (i *Item) IsShield() bool {
	return i.IsArmor() && i.GetArmorStats().Type == defs.ArmorTypeShield
}

func (i *Item) Repair() {
	if i.IsArmor() {
		i.GetArmorStats().Durability.SetToMax()
	}
}

func (i *Item) IsFood() bool {
	return i.GetItemType() == defs.ItemCategoryFood
}

func (i *Item) HasModifiers() bool {
	return len(i.equippedModifiers) > 0
}

func (i *Item) IsQuestItem() bool {
	return i.isQuestItem
}

type Enchantment interface {
	ToString() string
}
