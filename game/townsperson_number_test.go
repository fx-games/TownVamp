package game

import "testing"

// The Number is the only handle the player has on a Townsperson in the watch
// list, so it has to be dense, unique and the same for everyone every game.
func TestEveryTownspersonHasADistinctNumber(t *testing.T) {
	game := newTestGame(t)

	seen := make(map[int]string)
	for _, person := range game.townspeople {
		if person.Number < 1 || person.Number > len(game.townspeople) {
			t.Errorf("%s has number %d, outside 1..%d", person.InternalName, person.Number, len(game.townspeople))
		}
		if other, taken := seen[person.Number]; taken {
			t.Errorf("%s and %s share number %d", other, person.InternalName, person.Number)
		}
		seen[person.Number] = person.InternalName
	}

	// A number that moved between actors on a second load would mean the player
	// watched someone else under the same digit.
	for _, person := range newTestGame(t).townspeople {
		if seen[person.Number] != person.InternalName {
			t.Errorf("number %d went to %s, was %s", person.Number, person.InternalName, seen[person.Number])
		}
	}
}
