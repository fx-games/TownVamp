package defs

import (
	"cmp"
	"fmt"
	"math/rand"
	"path"
	"rpg/files"
	"rpg/stats"
	"slices"
)

type DataContainer struct {
	ItemDefinitions       map[string]ItemDefinition
	UniqueItemDefinitions map[string]ItemDefinition

	Modifiers map[string]func(source *rand.Rand, level int) *stats.Modifier

	ClassDefinitions  []LoadoutDefinition
	VendorDefinitions map[string]VendorDefinition
}

func (c DataContainer) GetItemCategories() []ItemCategory {
	result := make(map[ItemCategory]bool, 0)
	for _, def := range c.ItemDefinitions {
		result[def.InventoryCategory] = true
	}
	var categories []ItemCategory
	for cat := range result {
		categories = append(categories, cat)
	}
	slices.SortStableFunc(categories, func(i, j ItemCategory) int {
		return cmp.Compare(i, j)
	})

	return categories
}
func LoadGameData(logErr func(err string)) DataContainer {
	itemDefs := make(map[string]ItemDefinition)
	itemDir := path.Join("assets", "data", "items")
	otherDir := path.Join("assets", "data")

	join := path.Join(itemDir, "armor.rec")
	file, _ := files.Open(join)
	armorDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d armor definitions", len(armorDefs)))
	itemDefs = merge(itemDefs, armorDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "weapons.rec"))
	weaponDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d weapon definitions", len(weaponDefs)))
	itemDefs = merge(itemDefs, weaponDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "misc.rec"))
	throwableDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d misc. definitions", len(throwableDefs)))
	itemDefs = merge(itemDefs, throwableDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "accessory.rec"))
	accessoryDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d accessory definitions", len(accessoryDefs)))
	itemDefs = merge(itemDefs, accessoryDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "light.rec"))
	lightDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d light definitions", len(lightDefs)))
	itemDefs = merge(itemDefs, lightDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "potions.rec"))
	potionDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d potion definitions", len(potionDefs)))
	itemDefs = merge(itemDefs, potionDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "scrolls.rec"))
	scrollDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d scroll definitions", len(scrollDefs)))
	itemDefs = merge(itemDefs, scrollDefs)
	file.Close()

	file, _ = files.Open(path.Join(otherDir, "stores.rec"))
	vendorDefs := ReadVendorDefinitions(file)
	println(fmt.Sprintf("Loaded %d store definitions", len(vendorDefs)))
	file.Close()

	file, _ = files.Open(path.Join(otherDir, "classes.rec"))
	classDefs := ReadClassDefinitions(file)
	println(fmt.Sprintf("Loaded %d class definitions", len(classDefs)))
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "unique_items.rec"))
	uniqueItemDefs := ReadItemDefinitions(file, logErr)
	println(fmt.Sprintf("Loaded %d unique item definitions", len(uniqueItemDefs)))
	file.Close()

	return DataContainer{
		ItemDefinitions:        itemDefs,
		UniqueItemDefinitions:  uniqueItemDefs,
		VendorDefinitions:      vendorDefs,
		ClassDefinitions:       classDefs,
		Modifiers:              GetAllModifiers(),
	}
}

type BuiltInDataContainer struct {
	Modifiers         map[string]bool
	HitZones          map[string]bool
	HitEffects        map[string]bool
	TargetingRules    map[string]bool
	UntargetedEffects map[string]bool
}

func merge(defs map[string]ItemDefinition, other map[string]ItemDefinition) map[string]ItemDefinition {
	for key, value := range other {
		if _, ok := defs[key]; ok {
			println("duplicate key", key)
		}
		defs[key] = value
	}
	return defs
}
