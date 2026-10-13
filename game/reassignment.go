package game

import (
	"fmt"
	"rpg/common"
)

// reassignVoidedObligations is the town trying to repair itself. It runs at the
// Day boundary, a full Day after someone stopped showing up — the town cannot
// react to a disappearance it has not yet noticed.
//
// Candidates are tried in order: whoever the Obligation pointed at, then anyone
// whose Role suits it. Relationship before competence. A candidate needs free
// hours and nothing else; there is no displacement of their existing duties and
// no degraded performance. If nobody has capacity the Obligation goes Unmet,
// permanently, and nobody ever picks it up.
//
// That single constraint is what makes the town absorb early deaths and then
// fail everything at once later. The difficulty curve is emergent rather than
// tuned.
func (g *GameInstance) reassignVoidedObligations() {
	today := g.state.GetClock().Day()

	for _, person := range g.townspeople {
		if !person.HasBeenMissedForADay(today) {
			continue
		}
		voided := person.ReleaseObligations()
		if len(voided) == 0 {
			continue
		}
		g.PlayerMessage(fmt.Sprintf("Nobody has seen %s since yesterday.", displayName(person.InternalName)))

		for _, obligation := range voided {
			if taker := g.findTakerFor(obligation); taker != nil {
				taker.TakeObligation(obligation)
				if problems := taker.RecomputeSchedule(); len(problems) == 0 {
					g.PlayerMessage(fmt.Sprintf("%s takes on %s.", displayName(taker.InternalName), obligation.ID))
					continue
				}
				// Their day could not hold it after all. Hand it back: the town
				// reporting a loss is worth more than the town announcing a repair
				// that never happens, which is what a dropped Schedule entry is.
				taker.ReleaseObligation(obligation)
				taker.RecomputeSchedule()
			}
			g.unmetObligations = append(g.unmetObligations, obligation)
			g.PlayerMessage(obligation.Absent + ".")
		}
	}
}

// findTakerFor tries the person the Obligation pointed at first — the widow being
// fed can fetch her own bread — then anyone whose Role suits it. Capacity is the
// only bar.
func (g *GameInstance) findTakerFor(obligation *common.Obligation) *common.Townsperson {
	if obligation.Towards != "" {
		if pointedAt := g.townspersonByName(obligation.Towards); g.canTake(pointedAt, obligation) {
			return pointedAt
		}
	}

	if obligation.SuitableRole == "" {
		return nil // dies with its holder
	}
	for _, candidate := range g.townspeople {
		if candidate.Role != obligation.SuitableRole {
			continue
		}
		if g.canTake(candidate, obligation) {
			return candidate
		}
	}
	return nil
}

// canTake asks only about capacity, never competence — but about the capacity
// that matters, which is the Obligation's own hours rather than the day's total.
func (g *GameInstance) canTake(candidate *common.Townsperson, obligation *common.Obligation) bool {
	if candidate == nil || candidate.IsAbsent() {
		return false
	}
	if candidate.FreeSecondsPerDay() < obligation.DurationSeconds() {
		return false
	}
	return candidate.CanFitObligation(obligation)
}

// UnmetObligations is the ledger of what the town has permanently stopped doing.
// It is the unit of consequence, and the raw material of the Epilogue.
func (g *GameInstance) UnmetObligations() []*common.Obligation {
	return g.unmetObligations
}
