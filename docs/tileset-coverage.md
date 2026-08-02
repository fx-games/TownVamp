# Tileset coverage for the PoC town

What [Town Design Handover](specs/town-design-handover.md) asks for versus what `files/assets/classic_roguelike.png`
(28×28 tiles of 8×8 px, tile id = `row*28 + col`) can already draw. Names for existing ids live in
[defs/definedIcons.go](../defs/definedIcons.go); the ids used by the current map are listed in
[AGENTS.md](../AGENTS.md).

**Verdict:** terrain, buildings, fields, church, cemetery, people and most carried items are
already drawable. The gaps are almost entirely *work-place props* — the objects that make a mill
read as a mill and a forge as a forge — plus fences. Roughly 20 new 8×8 sprites cover everything,
and there are 190 empty tile slots to put them in.

## Covered

| Need | Tiles |
| --- | --- |
| Town square, main road | 1 (dirt), 0 (grass) |
| River, riverbank | 28, 29 (tagged `IsWater`) |
| Bridge, footbridge, dock, lumber | 63 (planks), 111 (ladder/boards), 188 (wood floor) |
| Building shells (inn, mill, smithy, workshop, farmhouse, barn) | walls 56/57/58, floors 12, 64–67, doors 69 (closed), 70 (open), 234 (doorway) |
| Church | altar 239, aisle carpet 236/237, doors 234, braziers 109, floor 64–67 |
| Cemetery graves | 78 (cross-topped stone), 166/194/222 (pale crosses), 350 (tomb mouth) |
| Flowers (for the grave, for blessing fields) | 274, plus 180 as a red-flower ground |
| Vegetable field / crop rows | 30, 31, 45, 46, 138 (leafy greens) |
| Grain field, hay, straw | 183 (grain), 181/182/184 (hay bands) |
| Trees, bushes | 4, 5, 6, 2, 275–278 (mushrooms) |
| Rocks, gravel, coal-ish ground | 10, 11, 36, 159 (embers, tagged `IsDamaging`), 160 |
| Fire, hearth, forge fire | 32, 33, 109, 597/598, 351 (lit torch) |
| Barrels (ale, storage) | 122, 125, 126 |
| Sacks and pouches (grain, flour, coin purse) | 162, 163, 164, 165, 409 (coins) |
| Beds | 756, 757, 758, 759 |
| Tables / benches / workbench surface | 353, 354, 505 |
| Shelves, racks (as bookshelf-ish blocks) | 76, 77 |
| Crates, baskets (woven) | 155, 62, 378 |
| Jug, flask, bottle (mug, bucket stand-ins) | 152, 153, 113 |
| Lantern, candle-light | 247, 246, 351 |
| Scroll, notes, book (prayer notes, sermon) | 139, 127 |
| Hammers, axes, blades, tools | 117, 168–173, 695–699, 119 (pickaxe) |
| Shop / trade signs on a wall | 114 (sword), 116 (shield), 117 (hammer), 119 (pickaxe), 154 (hanging sign) |
| Townspeople sprites, all six distinct | 19–27, 47–55, 79–83, 286–288, 314, 342, 370, 454–465 |

## Weak reuse — works, but reads as something else

- **Fence** — 158 (lattice) and 59 (checker) are dungeon wall art; they fence a plot but do not
  read as farm or cemetery fencing.
- **Gate** — 60 (dark double doors) or 69/70; reads as a house door, not a gate in a fence.
- **Grindstone / millstone** — 35 (brown ring) is the only round stone shape.
- **Counter (inn bar), lectern, workbench** — all fall back to 353/505, the same table block.
- **Chair** — nothing chair-shaped; a bench block has to stand in.
- **Bucket** — 152/153 read as jug or flask; fine for carrying, wrong for "pour water into barrel".
- **Iron scrap** — 692 (scattered brown/metal bits) is the nearest thing.
- **Bread** — 163 (small sack) or 165; no bread shape exists.

## Gaps — new sprites needed

Grouped by what the design doc needs them for. One 8×8 tile each unless noted.

**Mill (nothing here exists)**
1. Waterwheel / mill wheel — the single most identifying prop in the design; ideally 2×2 tiles
2. Grain chute
3. Flour sack distinct from grain sack (colour variant is enough)

**Blacksmith**
4. Anvil
5. Grindstone (a proper wheel-on-frame)
6. Tool rack
7. Tongs
8. Horseshoe

**Carpenter**
9. Sawhorse
10. Saw
11. Log / cut lumber pile (63 is flat planks, not a stack)
12. Wooden toy

**Inn**
13. Chair
14. Counter / bar front
15. Mug
16. Bucket

**Church**
17. Lectern
18. Bell rope (and a bell, if the bell is ever visible)
19. Candle (upright, distinct from a torch)

**Farm and cemetery**
20. Wooden fence — straight and corner, so 2 tiles
21. Gate
22. Wheelbarrow
23. Bread

That is 20 sprites, 24 tiles counting the fence pieces and a 2×2 waterwheel.

## Where to put them

190 tile slots in the sheet are fully transparent, so new art can be appended without moving
anything. Largest free runs: **511–531** (21), **726–755** (30), **761–783** (23), **700–717** (18),
**672–681** (10). Ids stay stable as long as sprites are only added into empty slots — never
reflow existing rows, the ids are hard-coded in [defs/definedIcons.go](../defs/definedIcons.go)
and baked into `maps.ldtk`.

After adding art, tag anything solid in LDtk's tileset editor (`IsBlockingMovement`,
`IsBlockingView`) — the game reads collision from those tags, not from the sprite.

## Open question for the engine, not the tileset

The design has people **carrying** things (basket, sack, flowers, bucket, toy). The map currently
gives an NPC a single `Tile` field, so a carried item has no place to render unless the engine
draws a second sprite next to the NPC, or swaps the NPC tile for a carrying variant. If the answer
is carrying variants, that multiplies sprite work by the number of carried items — worth settling
before any art is drawn.
