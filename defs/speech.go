package defs

// ActorSpeechTrigger says what made an actor speak. It came from the enemy
// definitions, which went with the combat systems, but Townspeople still talk —
// and overhearing them is how the player learns anything.
type ActorSpeechTrigger int

const (
	SpeechTriggerSight ActorSpeechTrigger = iota
	SpeechTriggerHurt
)
