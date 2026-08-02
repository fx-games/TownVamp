package common

import (
	"fmt"
	"io"
	"rpg/geometry"
	"rpg/recfile"
)

// ReadTownspeople loads Townspeople from the plaintext record format the rest of
// the game data already uses. Nothing about a Townsperson is hard-coded: Role,
// Identity and every Routine slot come from the file.
func ReadTownspeople(file io.Reader, namedLocation func(string) (geometry.Point, bool), logErr func(string)) map[string]*Townsperson {
	people := make(map[string]*Townsperson)
	for _, record := range recfile.Read(file) {
		person := &Townsperson{}
		for _, field := range record {
			switch field.Name {
			case "internal_name":
				person.InternalName = field.Value
			case "role":
				role, err := RoleFromString(field.Value)
				if err != nil {
					logErr(fmt.Sprintf("townsperson %s: %v", person.InternalName, err))
					continue
				}
				person.Role = role
			case "identity":
				person.Identity = field.Value
			case "routine":
				slot, err := ParseRoutineSlot(field.Value, namedLocation)
				if err != nil {
					logErr(fmt.Sprintf("townsperson %s: %v", person.InternalName, err))
					continue
				}
				person.Routine = append(person.Routine, slot)
			}
		}
		if person.InternalName == "" {
			logErr("townsperson record without an internal_name")
			continue
		}
		if person.Role == "" {
			logErr(fmt.Sprintf("townsperson %s has no role", person.InternalName))
			continue
		}
		SortRoutine(person.Routine)
		people[person.InternalName] = person
	}
	return people
}
