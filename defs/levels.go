package defs

type LevelName string

// The town is the only level. Dungeon levels, the hell town and their generated
// layouts were removed along with the dungeon crawler.
const LevelNameTown LevelName = "Town"

func IsTownLevel(levelName string) bool {
	return levelName == string(LevelNameTown)
}
