# 02 — Demolish the RPG systems

**What to build:** the game builds, loads the town, and lets the player walk around, with
every system belonging to the old dungeon crawler removed. Roughly six and a half
thousand lines go.

Deleted: the dungeon generator, the standalone combat simulator, melee, ranged, spell and
knockback actions, magic definitions, status effects, hit zones, targeting and the
targeted and untargeted action bases, weapon definitions, combat hooks, unique items,
quest dialogue and quest flags, and the RPG-specific interface panels for targeting,
status effects and enchantment transfer.

Also deleted: the existing two-state AI action selector. Its idle-and-attacking model has
nothing to contribute and behaviour will come from Schedules later. Actors may stand
still until then.

This is the largest single ticket by line count but not by thought — most of it is
deletion and fixing the call sites that break.

**Blocked by:** None — can start immediately.

**Status:** done

- [x] Dungeon generation and the combat simulator are gone
- [x] Combat, magic, status effects, hit zones and targeting are gone
- [x] Loot generation, unique items, quests and quest dialogue are gone
- [x] The two-state AI action selector is gone
      *the `ActorState` enum and its setters survive in `common/base_actor.go`, but nothing reads `GetState()` — no selector remains*
- [x] RPG-specific interface panels are gone
- [x] The project builds, the town map loads, and the player can move around it
- [x] The command-line subcommands that only served deleted systems are gone too
