package defs

import (
	"fmt"
	"io"
	"rpg/recfile"
)

func ReadItemDefinitions(file io.Reader, logErr func(errString string)) map[string]ItemDefinition {
	records := recfile.Read(file)
	itemDefinitions := make(map[string]ItemDefinition, len(records))
	for _, record := range records {
		definition := NewItemDefinitionFromRecord(record)
		if definition.InternalName == "" {
			logErr(fmt.Sprintf("Item definition has no internal name: %v", definition))
			continue
		}
		if _, exists := itemDefinitions[definition.InternalName]; exists {
			logErr(fmt.Sprintf("Duplicate item definition: %s", definition.InternalName))
			continue
		}
		if definition.Armor.ArmorClass.NotZero() && definition.Armor.ArmorType == ArmorTypeNone {
			logErr(fmt.Sprintf("Armor definition has invalid armor type: %v", definition))
			continue
		}
		if definition.Weapon.Damage.NotZero() && definition.Weapon.WeaponType == WeaponTypeNone {
			logErr(fmt.Sprintf("Weapon definition has invalid weapon type: %v", definition))
			continue
		}
		itemDefinitions[definition.InternalName] = definition
	}
	return itemDefinitions
}
