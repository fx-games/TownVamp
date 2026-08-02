package game

import (
	"os"
	"path/filepath"
	"rpg/common"
	"rpg/files"
	"rpg/geometry"
	"testing"
)

// This file is the seam every test in this project drives the game through:
// construct the game with the null user interface, advance it by some amount of
// simulated time, then assert on state a player could observe.
//
// Tests assert on what is visible — where an actor stands, what the map holds,
// what the world's clock has done. They do not reach into how that state was
// reached. A test that would have to know the order the timekeeper visited
// actors in, or which branch a schedule builder took, is testing the
// implementation and will break on every refactor without catching a defect.

const testSeed = 1

// newTestGame builds a game with no graphical subsystem, no audio playback and no
// text servers, seeded so that every run of a test sees the same world.
func newTestGame(t *testing.T) *GameInstance {
	t.Helper()
	useRepositoryAssets(t)

	startInfo := common.StartInfo{
		Title:       "test",
		ScreenSize:  geometry.Point{X: 640, Y: 480},
		TileScale:   1.0,
		LDTKMapFile: filepath.Join("assets", "maps.ldtk"),
		FileSystem:  files.GetFileSystem(),
		RandomSeed:  testSeed,
		// Zero value: audio off, no text servers, no automatic pickup. Anything a
		// test depends on it should turn on for itself, so the default stays quiet.
		UserConfiguration: common.Configuration{},
	}

	game := NewGameInstance(startInfo, NewNullUI())
	if game == nil {
		t.Fatal("NewGameInstance returned nil")
	}
	game.SetConsoleMode(true) // silences audio playback
	return game
}

// useRepositoryAssets points the asset filesystem at the real asset tree. The
// build-tagged default resolves relative to the working directory, which during a
// test is the package directory rather than the repository root.
func useRepositoryAssets(t *testing.T) {
	t.Helper()
	assetRoot, err := filepath.Abs(filepath.Join("..", "files"))
	if err != nil {
		t.Fatalf("resolving asset root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(assetRoot, "assets")); err != nil {
		t.Fatalf("asset tree not found at %s: %v", assetRoot, err)
	}
	dirFS, ok := os.DirFS(assetRoot).(files.FileSystem)
	if !ok {
		t.Fatal("os.DirFS does not satisfy files.FileSystem")
	}
	files.SetFileSystem(dirFS)
}

func TestGameConstructsHeadlessly(t *testing.T) {
	game := newTestGame(t)

	state := game.GetWorldState()
	if state == nil {
		t.Fatal("game has no world state")
	}
	if state.GetMap() == nil {
		t.Fatal("game loaded no map")
	}
	if player := state.GetPlayer(); player == nil {
		t.Fatal("game has no player")
	}
}

func TestPlayerStartsOnTheMap(t *testing.T) {
	game := newTestGame(t)
	state := game.GetWorldState()

	position := state.GetPlayer().MapPosition()
	if !state.GetMap().Contains(position) {
		t.Errorf("player starts at %v, which is outside the map", position)
	}
}

func TestAdvancingTimeLeavesTheWorldIntact(t *testing.T) {
	game := newTestGame(t)
	state := game.GetWorldState()

	// A move costs 10 AUT, so this is a hundred moves' worth of simulated time.
	for i := 0; i < 100; i++ {
		game.Advance(10)
	}

	if !state.GetPlayer().IsAlive() {
		t.Error("player died while nothing was happening")
	}
	position := state.GetPlayer().MapPosition()
	if !state.GetMap().Contains(position) {
		t.Errorf("player ended up at %v, outside the map", position)
	}
}

func TestTheSameSeedProducesTheSameWorld(t *testing.T) {
	first := newTestGame(t)
	second := newTestGame(t)

	for i := 0; i < 50; i++ {
		first.Advance(10)
		second.Advance(10)
	}

	firstActors := actorPositions(first)
	secondActors := actorPositions(second)

	if len(firstActors) != len(secondActors) {
		t.Fatalf("same seed produced %d actors and %d actors", len(firstActors), len(secondActors))
	}
	for name, position := range firstActors {
		if other, ok := secondActors[name]; !ok {
			t.Errorf("%q is missing from the second world", name)
		} else if other != position {
			t.Errorf("%q stands at %v in one world and %v in the other", name, position, other)
		}
	}
}

func actorPositions(game *GameInstance) map[string]geometry.Point {
	positions := make(map[string]geometry.Point)
	for _, actor := range game.GetWorldState().GetMap().Actors() {
		positions[actor.Name()] = actor.MapPosition()
	}
	return positions
}
