package game

import (
	"fmt"
	"path"
	"rpg/common"
	"rpg/files"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/recfile"
	"strconv"
	"strings"
)

// TownZone is a region with a permission rule attached. The map's own zone
// layer is empty, so these are authored alongside the Townspeople and painted
// onto the map at load.
type TownZone struct {
	Name       string
	Private    bool
	Rect       geometry.Rect
	BelongsTo  []string
	zoneOnMap  *gridmap.ZoneInfo
	Withdrawal bool
}

// MayEnter is the Zone half of the question secrecy turns on: may this person be
// here? Public Zones are open to everyone; a private one belongs to the people
// named on it.
func (z *TownZone) MayEnter(internalName string) bool {
	if !z.Private {
		return true
	}
	for _, resident := range z.BelongsTo {
		if resident == internalName {
			return true
		}
	}
	return false
}

func (g *GameInstance) loadZones() {
	file, err := files.Open(path.Join("assets", "data", "zones.rec"))
	if err != nil {
		g.ErrLog(fmt.Sprintf("no zones file: %v", err))
		return
	}
	defer file.Close()

	currentMap := g.state.GetMap()
	g.zones = nil

	for _, record := range recfile.Read(file) {
		zone := &TownZone{}
		for _, field := range record {
			switch field.Name {
			case "name":
				zone.Name = field.Value
			case "type":
				zone.Private = strings.EqualFold(field.Value, "private")
			case "belongs_to":
				zone.BelongsTo = append(zone.BelongsTo, field.Value)
			case "rect":
				rect, err := parseRect(field.Value)
				if err != nil {
					g.ErrLog(fmt.Sprintf("zone %s: %v", zone.Name, err))
					continue
				}
				zone.Rect = rect
			}
		}
		if zone.Name == "" {
			g.ErrLog("zone record without a name")
			continue
		}

		zoneType := gridmap.ZoneTypePublic
		if zone.Private {
			zoneType = gridmap.ZoneTypePrivate
		}
		zone.zoneOnMap = &gridmap.ZoneInfo{Name: zone.Name, Type: zoneType}
		currentMap.AddZone(zone.zoneOnMap)
		for x := zone.Rect.Min.X; x < zone.Rect.Max.X; x++ {
			for y := zone.Rect.Min.Y; y < zone.Rect.Max.Y; y++ {
				at := geometry.Point{X: x, Y: y}
				if currentMap.Contains(at) {
					currentMap.SetZone(at, zone.zoneOnMap)
				}
			}
		}
		g.zones = append(g.zones, zone)
	}
	println(fmt.Sprintf("Loaded %d zones", len(g.zones)))
}

func parseRect(value string) (geometry.Rect, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return geometry.Rect{}, fmt.Errorf("expected x,y,width,height, got %q", value)
	}
	numbers := make([]int, 4)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return geometry.Rect{}, fmt.Errorf("bad number %q in %q", part, value)
		}
		numbers[i] = n
	}
	return geometry.NewRect(numbers[0], numbers[1], numbers[0]+numbers[2], numbers[1]+numbers[3]), nil
}

// zoneAt finds the authored Zone covering a tile, if any.
func (g *GameInstance) zoneAt(at geometry.Point) *TownZone {
	for _, zone := range g.zones {
		if at.X >= zone.Rect.Min.X && at.X < zone.Rect.Max.X &&
			at.Y >= zone.Rect.Min.Y && at.Y < zone.Rect.Max.Y {
			return zone
		}
	}
	return nil
}

// checkForTrespass is what happens when a Townsperson sees the vampire somewhere
// he does not belong. They run, they postpone whatever they were about to do,
// and they Withdraw — spending more of their day indoors from then on.
//
// Nothing is reported and no alarm is raised. The town's response is a slow
// tightening rather than a loud event: enough Withdrawal and it starves the
// vampire out without anyone ever understanding what they have done.
func (g *GameInstance) checkForTrespass() {
	player := g.state.GetPlayer()
	if player == nil || !player.IsAlive() {
		return
	}
	zone := g.zoneAt(player.MapPosition())
	if zone == nil || !zone.Private || zone.MayEnter("") {
		return
	}

	for _, actor := range g.state.GetMap().Actors() {
		person := actor.GetTownsperson()
		if person == nil || person.IsAbsent() {
			continue
		}
		if !g.canSee(actor, player.MapPosition()) {
			continue
		}
		g.witnessTrespass(actor, person, zone)
	}
}

func (g *GameInstance) witnessTrespass(actor *common.Actor, person *common.Townsperson, zone *TownZone) {
	if g.withdrawn[person.InternalName] {
		return
	}
	g.withdrawn[person.InternalName] = true
	person.Withdraw()

	g.fleeFrom(actor, g.state.GetPlayer().MapPosition())
	g.PlayerMessage(fmt.Sprintf("%s sees you in %s and hurries away.", actor.Name(), zone.Name))
}

// fleeFrom sends an actor one step away from something. Ordinary Townspeople
// only ever flee; a Hunter is the single exception that pursues.
func (g *GameInstance) fleeFrom(actor *common.Actor, from geometry.Point) {
	currentMap := g.state.GetMap()
	here := actor.MapPosition()
	best := here
	bestDistance := currentMap.MoveDistance(here, from)
	for _, candidate := range currentMap.GetFilteredNeighbors(here, func(p geometry.Point) bool {
		return currentMap.Contains(p) && currentMap.IsWalkable(p) && !currentMap.IsActorAt(p)
	}) {
		if distance := currentMap.MoveDistance(candidate, from); distance > bestDistance {
			best, bestDistance = candidate, distance
		}
	}
	if best != here {
		g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Flee"), actor, best, false))
	}
}
