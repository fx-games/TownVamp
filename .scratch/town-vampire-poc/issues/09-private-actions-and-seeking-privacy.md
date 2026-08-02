# 09 — Private Actions and seeking privacy

**What to build:** Townspeople begin keeping secrets. Some of what they do is a Private
Action carrying a Witness Whitelist naming who may see it, and anyone else in view makes
them break off and go elsewhere. This is what gives the player something worth catching
people at, and it is the first ticket where being hidden matters rather than merely being
nearby.

A group secret is a Private Action whose whitelist holds several names, so watching who is
permitted to be present is itself informative.

**Keep the perception narrow.** Build one query — who can currently see me — and run it
only for Townspeople attempting or performing a Private Action. No general NPC vision.
Cost stays bounded by the number of Private Actions in flight rather than by population.
The check runs while the action is in progress, not only at its start, so someone
wandering in halfway aborts it too.

The existing field of view already stops a viewer standing in a lit tile from seeing into
unlit ones. That asymmetry is the whole reason a vampire can watch without being watched,
and it is used exactly as it stands.

Per ADR 0002, nothing is recorded about what the player saw. There is no dossier and no
discovered-state. What a Private Action looks like should be abstract and open to
interpretation — a door, an object carried, a thing used, a conversation within earshot.

**Blocked by:** 08 — Obligations and derived Schedule.

**Status:** ready-for-agent

- [ ] An action can be marked Private and given a Witness Whitelist in the record format
- [ ] A Townsperson performs a Private Action only when no unwhitelisted observer can see
      them
- [ ] A Private Action is abandoned when an unwhitelisted observer arrives partway through
- [ ] An interrupted Townsperson moves somewhere unobserved and attempts the action again
      if such a place is reachable
- [ ] The perception query runs only for in-flight Private Actions, not for every actor
      every tick
- [ ] A vampire standing in darkness can watch a Private Action performed in light without
      interrupting it
- [ ] Nothing is stored about what the player has observed
