package common

// Presence is what the town can tell about a Townsperson. It is deliberately
// coarse: the town cannot distinguish a murder from someone having gone away,
// and that uncertainty is the vampire's operating room.
type Presence int

const (
	// PresenceAlive: present and meeting their Obligations.
	PresenceAlive Presence = iota

	// PresenceAbsent: not where they should be. Covers the murdered, the
	// vanished and the merely away. After one full Day of Absence their
	// Obligations are offered for Reassignment.
	PresenceAbsent

	// PresenceConfirmedDead: the town has proof. Reassignment has already
	// happened by then; what changes is that the town now knows.
	PresenceConfirmedDead
)

func (p Presence) String() string {
	switch p {
	case PresenceAlive:
		return "alive"
	case PresenceAbsent:
		return "absent"
	case PresenceConfirmedDead:
		return "confirmed dead"
	}
	return "unknown"
}
