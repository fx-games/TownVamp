# 10 — Zones, trespass and Withdrawal

**What to build:** the town divides into places the vampire is free to be and places he is
not, and being caught somewhere he does not belong costs him access to that person for the
rest of the run.

Zone permissions reuse the map's existing public, private and high-security zone types,
which are already loaded from the map format and currently unused. A Zone's permitted
occupants are the same predicate as a Private Action's Witness Whitelist — may this person
be here, seeing this — asked of a place rather than an act, and authored the same way.

A Townsperson who sees the vampire where he does not belong runs away and postpones any
Private Action. **They do not challenge him and they do not raise an alarm.** Instead they
Withdraw: their Routine biases toward private Zones and indoor hours for the remainder of
the run. Withdrawal is a Routine bias and not a new system, and it is the only thing in
the game that modifies a Routine.

Enough Withdrawal across enough Townspeople and the town starves the vampire out without
anyone ever raising a hand against him or understanding what they have done. That is the
intended shape of the pressure — a slow tightening, never a loud event.

**Blocked by:** 07 — Basic test town. 09 — Private Actions and seeking privacy.

**Status:** ready-for-agent

- [ ] Zone permissions are read from the existing map format and authored on the map
- [ ] A Townsperson who sees the vampire in a Zone where he does not belong flees
- [ ] That Townsperson postpones any Private Action they were about to perform
- [ ] That Townsperson thereafter spends measurably more of their day indoors and in
      private Zones
- [ ] No alarm is raised and no other Townsperson is informed
- [ ] Withdrawal persists for the rest of the run
- [ ] A test asserts a Townsperson's indoor hours increase after witnessing a trespass
