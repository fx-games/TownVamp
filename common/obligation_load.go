package common

import (
	"fmt"
	"io"
	"rpg/geometry"
	recfile "github.com/memmaker/rec-files"
	"strings"
)

// ReadObligations loads Obligations from the record format and hands each one to
// the Townsperson who owes it. An Obligation naming an owner nobody recognises
// is reported rather than dropped quietly — a duty owed by no one would degrade
// the town for reasons the player could never trace.
func ReadObligations(file io.Reader, people map[string]*Townsperson, namedLocation func(string) (geometry.Point, bool), logErr func(string)) []*Obligation {
	var all []*Obligation

	for _, record := range recfile.Read(file) {
		obligation := &Obligation{}
		for _, field := range record {
			switch field.Name {
			case "id":
				obligation.ID = field.Value
			case "owner":
				obligation.Owner = field.Value
			case "towards":
				obligation.Towards = field.Value
			case "role":
				role, err := RoleFromString(field.Value)
				if err != nil {
					logErr(fmt.Sprintf("obligation %s: %v", obligation.ID, err))
					continue
				}
				obligation.SuitableRole = role
			case "at":
				place, err := parsePlace(field.Value, namedLocation)
				if err != nil {
					logErr(fmt.Sprintf("obligation %s: %v", obligation.ID, err))
					continue
				}
				obligation.At = place
			case "window":
				span := strings.Split(field.Value, "-")
				if len(span) != 2 {
					logErr(fmt.Sprintf("obligation %s: expected 'HH:MM-HH:MM', got %q", obligation.ID, field.Value))
					continue
				}
				start, startErr := ParseClockTime(span[0])
				end, endErr := ParseClockTime(span[1])
				if startErr != nil || endErr != nil {
					logErr(fmt.Sprintf("obligation %s: bad window %q", obligation.ID, field.Value))
					continue
				}
				obligation.StartSecond, obligation.EndSecond = start, end
			case "performed":
				obligation.Performed = field.Value
			case "absent":
				obligation.Absent = field.Value
			case "private":
				obligation.Private = strings.EqualFold(field.Value, "true") || field.Value == "1"
			case "witness":
				obligation.Witnesses = append(obligation.Witnesses, field.Value)
			}
		}

		if obligation.ID == "" {
			logErr("obligation record without an id")
			continue
		}
		if obligation.Performed == "" || obligation.Absent == "" {
			logErr(fmt.Sprintf("obligation %s must be authored twice: what it looks like performed, and what its absence looks like", obligation.ID))
			continue
		}
		owner, found := people[obligation.Owner]
		if !found {
			logErr(fmt.Sprintf("obligation %s is owed by %q, who does not live here", obligation.ID, obligation.Owner))
			continue
		}
		if obligation.Towards != "" {
			if _, known := people[obligation.Towards]; !known {
				logErr(fmt.Sprintf("obligation %s points at %q, who does not live here", obligation.ID, obligation.Towards))
				continue
			}
		}

		obligation.OriginalOwner = obligation.Owner
		owner.TakeObligation(obligation)
		all = append(all, obligation)
	}

	for _, person := range people {
		for _, problem := range person.RecomputeSchedule() {
			logErr(fmt.Sprintf("%s: %v", person.InternalName, problem))
		}
	}
	return all
}
