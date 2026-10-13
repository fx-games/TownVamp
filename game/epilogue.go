package game

import (
	"fmt"
	"rpg/common"
	"rpg/util"
	"strings"
)

// EpilogueSlide is one card of the ending: a Townsperson's end state and what
// followed from it, or the town as a whole.
type EpilogueSlide struct {
	Title string
	Lines []string
}

// Epilogue is what is shown once a run ends: one slide per Townsperson, plus one
// summarising the town.
//
// These are raw data readouts rather than prose. The simulation is what generates
// the story; the Epilogue only reports it. It says nothing about whether the
// player chose well — the game has no notion of a right or wrong victim, and the
// ending must not smuggle one back in.
func (g *GameInstance) Epilogue() []EpilogueSlide {
	var slides []EpilogueSlide

	for _, person := range g.townspeople {
		slides = append(slides, g.slideFor(person))
	}
	slides = append(slides, g.townSlide())
	return slides
}

func (g *GameInstance) slideFor(person *common.Townsperson) EpilogueSlide {
	lines := []string{
		fmt.Sprintf("%s, %s", titleCase(string(person.Role)), person.Identity),
		"",
	}

	switch person.Presence {
	case common.PresenceAlive:
		lines = append(lines, "Still here.")
	case common.PresenceAbsent:
		if person.BodyHidden {
			lines = append(lines, "Vanished. Nobody ever found out where to.")
		} else {
			lines = append(lines, "Vanished. Nobody could say when they had last been seen.")
		}
	case common.PresenceConfirmedDead:
		lines = append(lines, "Found dead. The town knew there was something among them.")
	}

	if len(person.Obligations) > 0 {
		lines = append(lines, "", "Still doing:")
		for _, obligation := range person.Obligations {
			marker := "  "
			if obligation.IsInherited() {
				marker = "  taken on after " + displayName(obligation.OriginalOwner) + " went: "
			}
			lines = append(lines, marker+obligation.Performed)
		}
	}

	if person.HasWithdrawn {
		lines = append(lines, "", "Kept indoors after something they saw and never spoke of.")
	}

	inherited := 0
	for _, obligation := range person.Obligations {
		if obligation.IsInherited() {
			inherited++
		}
	}
	if inherited > 0 {
		lines = append(lines, "", fmt.Sprintf("Carried %d duty that was not theirs.", inherited))
	}

	return EpilogueSlide{Title: displayName(person.InternalName), Lines: lines}
}

func (g *GameInstance) townSlide() EpilogueSlide {
	alive, absent, confirmed, withdrawn := 0, 0, 0, 0
	for _, person := range g.townspeople {
		switch person.Presence {
		case common.PresenceAlive:
			alive++
		case common.PresenceAbsent:
			absent++
		case common.PresenceConfirmedDead:
			confirmed++
		}
		if person.HasWithdrawn {
			withdrawn++
		}
	}

	clock := g.state.GetClock()
	lines := []string{
		fmt.Sprintf("The vampire %s, on day %d.", g.endingDescription(), clock.Day()+1),
		"",
		fmt.Sprintf("Of %d: %d still living, %d vanished, %d found dead.", len(g.townspeople), alive, absent, confirmed),
	}
	if withdrawn > 0 {
		lines = append(lines, fmt.Sprintf("%d had stopped going out.", withdrawn))
	}

	if len(g.unmetObligations) == 0 {
		lines = append(lines, "", "Everything the town needed doing still got done.")
	} else {
		lines = append(lines, "", "What stopped being done, and was never taken up again:")
		for _, obligation := range g.unmetObligations {
			lines = append(lines, "  "+obligation.Absent)
		}
	}

	return EpilogueSlide{Title: "The town", Lines: lines}
}

func (g *GameInstance) endingDescription() string {
	switch g.ending {
	case RunLeftTown:
		return "moved on"
	case RunStarved:
		return "starved"
	case RunBurned:
		return "burned"
	case RunCaught:
		return "was caught"
	}
	return "is still there"
}

// ShowEpilogue puts the ending on screen. It fires once, when the run ends.
func (g *GameInstance) ShowEpilogue() {
	var text []util.LabelText
	for _, slide := range g.Epilogue() {
		text = append(text, util.LabelTextFromString("— "+slide.Title+" —"))
		for _, line := range slide.Lines {
			text = append(text, util.LabelTextFromString(line))
		}
		text = append(text, util.LabelTextFromString(""))
	}
	g.uiController.OpenTextWindow(text)
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
