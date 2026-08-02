package game

import (
	"fmt"
	"path"
	"rpg/common"
	"rpg/defs"
	"rpg/files"
	"rpg/geometry"
	"rpg/util"
)

// populateTown gives the town its inhabitants. Townspeople already placed on the
// map by the map file are matched by internal name and given their Role and
// Routine; any authored Townsperson the map does not contain is spawned at the
// first place their Routine puts them.
//
// This runs after the map is loaded because Routines may name locations on it.
func (g *GameInstance) populateTown() {
	currentMap := g.state.GetMap()
	if currentMap == nil {
		return
	}

	namedLocation := func(name string) (geometry.Point, bool) {
		locations := currentMap.GetNamedLocations()
		pos, ok := locations[name]
		return pos, ok
	}

	file, err := files.Open(path.Join("assets", "data", "townspeople.rec"))
	if err != nil {
		g.ErrLog(fmt.Sprintf("no townspeople file: %v", err))
		return
	}
	defer file.Close()

	people := common.ReadTownspeople(file, namedLocation, g.ErrLog)

	existing := make(map[string]*common.Actor)
	for _, actor := range currentMap.Actors() {
		if name := actor.GetInternalName(); name != "" {
			existing[name] = actor
		}
	}

	for name, person := range people {
		if actor, found := existing[name]; found {
			actor.SetTownsperson(person)
			// Their old life as a shopkeeper is over.
			actor.SetVendorType(common.VendorTypeNone)
			continue
		}
		g.spawnTownsperson(person)
	}

	obligations := g.loadObligations(people, namedLocation)
	g.checkSchedulesAreWalkable()
	println(fmt.Sprintf("Loaded %d townspeople owing %d obligations", len(people), len(obligations)))
}

func (g *GameInstance) loadObligations(people map[string]*common.Townsperson, namedLocation func(string) (geometry.Point, bool)) []*common.Obligation {
	file, err := files.Open(path.Join("assets", "data", "obligations.rec"))
	if err != nil {
		g.ErrLog(fmt.Sprintf("no obligations file: %v", err))
		return nil
	}
	defer file.Close()
	return common.ReadObligations(file, people, namedLocation, g.ErrLog)
}

// checkSchedulesAreWalkable is the cost of deriving Schedules rather than
// authoring them: nothing otherwise guarantees a Townsperson can physically get
// from one slot to the next in the hours allowed. Reported at load, not
// discovered at runtime.
func (g *GameInstance) checkSchedulesAreWalkable() {
	currentMap := g.state.GetMap()
	travelSeconds := func(from, to geometry.Point) (float64, bool) {
		path := currentMap.GetJPSPath(from, to, func(p geometry.Point) bool {
			return currentMap.Contains(p) && currentMap.IsWalkable(p)
		})
		if len(path) == 0 {
			return 0, false
		}
		// A step costs 10 AUT, and an AUT is SecondsPerAut in-game seconds.
		return float64(len(path)) * 10 * common.SecondsPerAut, true
	}

	for _, actor := range currentMap.Actors() {
		person := actor.GetTownsperson()
		if person == nil {
			continue
		}
		for _, problem := range person.Schedule().CheckWalkable(travelSeconds) {
			g.ErrLog(fmt.Sprintf("%s: %v", person.InternalName, problem))
		}
	}
}

func (g *GameInstance) spawnTownsperson(person *common.Townsperson) {
	currentMap := g.state.GetMap()

	spawnAt := geometry.Point{}
	if len(person.Routine) > 0 {
		spawnAt = person.Routine[0].At
	}
	if !currentMap.Contains(spawnAt) || !currentMap.IsWalkable(spawnAt) {
		g.ErrLog(fmt.Sprintf("townsperson %s has nowhere to spawn (%v)", person.InternalName, spawnAt))
		return
	}
	if currentMap.IsActorAt(spawnAt) {
		free := currentMap.GetFreeCardinalNeighbors(spawnAt)
		if len(free) == 0 {
			g.ErrLog(fmt.Sprintf("townsperson %s has no free tile at %v", person.InternalName, spawnAt))
			return
		}
		spawnAt = free[0]
	}

	icon := int32('@')
	actor := common.NewActor(icon)
	actor.SetInternalName(person.InternalName)
	actor.SetName(displayName(person.InternalName))
	actor.SetAsciiIcon(rune(displayName(person.InternalName)[0]))
	actor.SetCP437Icon(int32(util.UnicodeToCP437Byte(rune(displayName(person.InternalName)[0]))))
	actor.SetColor(defs.White)
	actor.SetNeutral()
	actor.SetTownsperson(person)
	actor.SetDrawPosition(spawnAt.ToPointF())

	currentMap.AddActor(actor, spawnAt)
	g.state.GetTimeKeeper().AddTimed(actor)
}

func displayName(internalName string) string {
	if internalName == "" {
		return "Someone"
	}
	return string(internalName[0]-32) + internalName[1:]
}
