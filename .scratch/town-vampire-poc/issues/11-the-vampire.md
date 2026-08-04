# 11 — The vampire

**What to build:** the player becomes a vampire with a hunger and a deadline. Blood drains
continuously and is refilled only by Killing. Killing leaves a body that can be hidden.
Dawn outdoors is fatal.

Blood is the only thing forcing the player to act — without it they could watch the town
forever and never choose anyone. It drains continuously and never stops, and the player
can see how much is left so they can judge how many more nights of observation they can
afford.

Killing sets the victim's Presence to Absent, not to dead. **The town cannot tell a murder
from someone having simply gone away**, and that uncertainty is the vampire's operating
room. The body remains, and can be hidden; what happens when it is found belongs to a
later ticket.

Sunlight kills a vampire caught outdoors at dawn, building on the daylight and lighting
state that already exists. This is the hard deadline that structures every night: how far
from shelter it is safe to go, and whether there is time to get back. The player needs
enough information to plan around it rather than be ambushed by it.

Feeding without killing is deliberately out of scope. Killing is the only way to refill
Blood in the proof of concept.

**Blocked by:** 07 — Basic test town. 04 — Real-time clock.

**Status:** done

- [x] Blood drains continuously at a configurable rate and is shown to the player
- [x] Killing a Townsperson refills Blood by a configurable amount
- [x] A killed Townsperson's Presence becomes Absent, indistinguishable from having gone
      away
- [x] A killed Townsperson leaves a body, and the player can hide it
- [x] Running out of Blood ends the run
- [x] Being outdoors at dawn ends the run
      *sunlight burns Blood away fast rather than killing on the instant; the run still ends*
- [x] The player can judge how long remains before dawn
- [x] A test asserts Blood drains over a Day and is restored by a Kill
