# Tileset coverage for the PoC town

What [Town Design Handover](specs/town-design-handover.md) asks for versus what `files/assets/classic_roguelike.png`
(28×28 tiles of 8×8 px, tile id = `row*28 + col`) can already draw. Names for existing ids live in
[defs/definedIcons.go](../defs/definedIcons.go); the ids used by the current map are listed in
[AGENTS.md](../AGENTS.md).

**Verdict:** terrain, buildings, fields, church, cemetery, people and most carried items were
already drawable. The gaps were almost entirely *work-place props* — the objects that make a mill
read as a mill and a forge as a forge — plus fences. Those 24 tiles are now drawn (ids 728–753,
761–762); nothing on the design doc's scenery or item lists is missing art any more.

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

## Was missing — now drawn

All the props the design needs and the sheet lacked have been painted into empty slots by
[scripts/add_sprites.py](../scripts/add_sprites.py), which holds the art as ASCII and is safe to
re-run. Names are Go constants in [defs/definedIcons.go](../defs/definedIcons.go); the solid ones
are tagged `IsBlockingMovement` in `maps.ldtk`.

| id | sprite | id | sprite |
| --- | --- | --- | --- |
| 728 | anvil | 742 | bucket |
| 729 | grindstone | 743 | lectern |
| 730 | tool rack | 744 | bell |
| 731 | tongs | 745 | candle |
| 732 | horseshoe | 746 | fence, horizontal |
| 733/734/761/762 | **waterwheel** (2×2) | 747 | fence, vertical |
| 735 | sawhorse | 748 | gate |
| 736 | saw | 749 | wheelbarrow |
| 737 | log pile | 750 | bread |
| 738 | wooden toy | 751 | grain chute |
| 739 | chair | 752 | flour sack |
| 740 | counter / bar | 753 | bell rope |
| 741 | mug | 763 | **window, horizontal wall** |
| | | 764 | **window, vertical wall** |

Blocking (movement): 728, 729, 730, 733, 734, 735, 737, 740, 743, 746, 747, 748, 749, 751, 761,
762, 763, 764. The waterwheel also blocks view. Chairs are deliberately walkable so a Townsperson
can sit on one.

**Windows (763, 764) block movement and nothing else** — they are the only tiles in the sheet that
do. `IsWalkable` and `IsTransparent` have always been independent bools in `gridmap`, and until now
nothing used the gap between them. A window is what lets a vampire stand in a dark street and watch
a Private Action happen in a lit room he cannot walk into. If either ever gets tagged
`IsBlockingView`, every house becomes opaque again and the secrets inside them become
unobservable — so check that tag before blaming the perception code.

763 sits in a horizontal (top/bottom) wall and 764 in a vertical (side) wall; both keep the brick
rhythm of the wall they interrupt, so a window reads as part of the building rather than a hole in
it.

Still standing in for something else, and fine that way: **iron scrap** → 692 (scattered metal
bits), **hoe** → 119 (pickaxe), **basket** → 62 or 378 (wicker), **prayer book** → 127,
**prayer notes** → 139 (scroll), **coin purse** → 165, **grain sack** → 163.

## Where new art goes

The new sprites took 728–753 and 761–764. Around 160 slots are still fully transparent for
whatever comes next — largest free runs: **511–531** (21), **765–783** (19), **700–717** (18),
**672–681** (10). Ids must stay stable: only ever add art into empty slots, never reflow existing
rows, because ids are hard-coded in [defs/definedIcons.go](../defs/definedIcons.go) and baked into
`maps.ldtk`.

After adding art, tag anything solid in LDtk's tileset editor (`IsBlockingMovement`,
`IsBlockingView`) — the game reads collision from those tags, not from the sprite.

## Open question for the engine, not the tileset

The design has people **carrying** things (basket, sack, flowers, bucket, toy). The map currently
gives an NPC a single `Tile` field, so a carried item has no place to render unless the engine
draws a second sprite next to the NPC, or swaps the NPC tile for a carrying variant. If the answer
is carrying variants, that multiplies sprite work by the number of carried items — worth settling
before any art is drawn.
