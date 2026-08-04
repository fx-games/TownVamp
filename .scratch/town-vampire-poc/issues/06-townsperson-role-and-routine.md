# 06 — Townsperson, Role and Routine

**What to build:** one Townsperson, defined entirely in the existing plaintext record
format, walks their authored Routine through a full Day and is somewhere different at
dawn than at midnight. The first end-to-end slice from an authored file to visible
behaviour in the town.

A Townsperson carries a Role, an Identity, a Routine and a Presence. The Routine is the
fixed part of their day — sleep, meals, home — authored and never reacting to anything.
It is publicly legible: a glance at the street should show it.

Role is drawn from the closed set of five in the glossary: Farmer, Smith, Builder, Trader,
Authority.

The hours a Routine leaves free are that Townsperson's spare capacity, which later becomes
the only limit on Reassignment. It needs to be a queryable quantity from the start, not
something derived later.

Use the existing record format and its round-trip tooling. Do not introduce a new
authoring format.

**Blocked by:** 04 — Real-time clock. 01 — Headless test seam.

**Status:** done

- [x] A Townsperson's Role, Identity and Routine are authored in the record format and
      loaded at startup
- [x] Role is constrained to the five values in the glossary
- [x] A Townsperson walks their Routine unattended across a full Day and returns to where
      they started
- [x] The free hours remaining in a Routine can be queried
- [x] A test advances a Day and asserts the Townsperson is in the expected place at
      several hours
      *position is asserted at one hour (11:00) rather than several*
