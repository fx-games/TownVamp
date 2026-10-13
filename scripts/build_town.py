#!/usr/bin/env python3
"""Paint the PoC town described in docs/specs/town-design-handover.md into the
Tiles layer of files/assets/maps.ldtk.

The Tiles layer, the level size and the Meta layer's Location entities are
written from here: terrain and the names the data files address it by are one
decision, so they live in one file. NPCs, Doors and Signs on the Entities layer
are not touched — people are authored in files/assets/data/townspeople.rec.
Re-running is safe: the tile layer and the places are rebuilt from scratch every
time.

Tile id = row * 28 + col in files/assets/classic_roguelike.png. Names for the
ids used here are Go constants in defs/definedIcons.go.
"""
import json
import os
import uuid

LDTK = os.path.join(os.path.dirname(__file__), "..", "files", "assets", "maps.ldtk")
CELL = 8
W, H = 72, 40

# --- tiles -----------------------------------------------------------------
GRASS, DIRT = 0, 1
PINE, BARE_TREE, SHRUB = 4, 5, 6
MOUNTAIN, CAVE = 7, 8
BOULDER, PEBBLES = 10, 11
WATER = 28
CROP_A, CROP_B, CROP_C, CROP_D, GREENS = 30, 31, 45, 46, 138
FIRE = 32
PLANKS = 63
FLOOR, FLOOR_B = 64, 65
SHELF_A, SHELF_B = 76, 77
GRAVE, GRAVE_PALE, GRAVE_PALE_B, GRAVE_PALE_C, TOMB = 78, 166, 194, 222, 350
BRAZIER = 109
BOOK, SCROLL = 127, 139
BARREL, BARREL_B = 122, 125
BASKET, CRATE = 62, 155
SACK = 163
COAL = 160
HAY = 181
GRAIN = 183
RED_GRASS = 180
DOORWAY = 234
CARPET_V, ALTAR = 237, 239
TABLE, BENCH = 353, 505
FLOWERS = 274
SHRINE = 75
IRON_SCRAP = 692
WALL_CORNER, WALL_V, WALL_H = 56, 57, 58
BED_A, BED_B, BED_C, BED_D = 756, 757, 758, 759
ANVIL, GRINDSTONE, TOOL_RACK, TONGS, HORSESHOE = 728, 729, 730, 731, 732
WHEEL_TL, WHEEL_TR, WHEEL_BL, WHEEL_BR = 733, 734, 761, 762
SAWHORSE, SAW, LOG_PILE, TOY = 735, 736, 737, 738
CHAIR, COUNTER, MUG, BUCKET = 739, 740, 741, 742
LECTERN, BELL, CANDLE = 743, 744, 745
FENCE_H, FENCE_V, GATE = 746, 747, 748
WHEELBARROW, BREAD, CHUTE, FLOUR, BELL_ROPE = 749, 750, 751, 752, 753
WINDOW_H, WINDOW_V = 763, 764

grid = [[GRASS] * W for _ in range(H)]


def put(x, y, t):
    if 0 <= x < W and 0 <= y < H:
        grid[y][x] = t


def rect(x, y, w, h, t):
    for cy in range(y, y + h):
        for cx in range(x, x + w):
            put(cx, cy, t)


def hline(x, y, w, t):
    rect(x, y, w, 1, t)


def vline(x, y, h, t):
    rect(x, y, 1, h, t)


BUILDINGS = []


def building(x, y, w, h, floor=FLOOR, windows=True):
    """Walls around a flagstone interior, with windows in the outer walls."""
    rect(x, y, w, h, floor)
    hline(x, y, w, WALL_H)
    hline(x, y + h - 1, w, WALL_H)
    vline(x, y, h, WALL_V)
    vline(x + w - 1, y, h, WALL_V)
    for cx, cy in ((x, y), (x + w - 1, y), (x, y + h - 1), (x + w - 1, y + h - 1)):
        put(cx, cy, WALL_CORNER)
    if windows:
        pierce_windows(x, y, w, h)
    BUILDINGS.append((x, y, w, h))


def pierce_windows(x, y, w, h):
    """Put windows in a building's outer walls.

    Windows block movement but not sight, so they are what lets the vampire watch
    a Private Action performed indoors from a street he cannot be reached from.
    A house without one is a box the player can never see into.

    Doors are punched after the walls are drawn, and they go in the middle of a
    horizontal wall — so windows sit two tiles in from each corner, where a door
    will not land on them, and on both side walls, which never carry doors here.
    """
    if h >= 5:
        cy = y + h // 2
        put(x, cy, WINDOW_V)
        put(x + w - 1, cy, WINDOW_V)
    if w >= 7:
        for cx in (x + 2, x + w - 3):
            put(cx, y, WINDOW_H)
            put(cx, y + h - 1, WINDOW_H)


def fenced(x, y, w, h):
    hline(x, y, w, FENCE_H)
    hline(x, y + h - 1, w, FENCE_H)
    vline(x, y, h, FENCE_V)
    vline(x + w - 1, y, h, FENCE_V)


def scatter(points, t):
    for x, y in points:
        put(x, y, t)


# --- terrain ---------------------------------------------------------------
hline(0, 0, W, MOUNTAIN)
hline(0, H - 1, W, MOUNTAIN)
vline(0, 0, H, MOUNTAIN)
vline(W - 1, 0, H, MOUNTAIN)

# main road east-west, and the spine north-south through the square
hline(1, 20, W - 2, DIRT)
vline(36, 4, 33, DIRT)
rect(32, 16, 10, 9, DIRT)  # town square
put(34, 18, SHRINE)

# The market stall, on the north side of the road where it crosses the square.
# The counter blocks, so the trader stands behind it and everyone else deals
# with her across it from the road — which is why the stall is a place people
# are seen at rather than a building they disappear into.
hline(33, 19, 3, COUNTER)
put(32, 19, CRATE)
put(36, 19, BASKET)
put(33, 18, SACK)
put(35, 18, BARREL)

# river down the east side, bridge where the main road crosses it
rect(64, 1, 4, H - 2, WATER)
hline(62, 20, 8, PLANKS)
rect(62, 30, 2, 1, PLANKS)  # dock
scatter([(68, 29), (70, 28), (69, 32), (68, 33)], SHRUB)  # quiet riverbank
scatter([(69, 27), (70, 31)], PINE)

# the vampire's way in: a cave mouth south-west of the fields
put(5, 36, CAVE)

# --- farm ------------------------------------------------------------------
# vegetable field: crop rows with walkable lanes between them
fenced(2, 3, 14, 10)
for i, row in enumerate(range(4, 12)):
    if row % 2 == 0:
        hline(3, row, 12, [CROP_A, CROP_B, CROP_C, CROP_D][(i // 2) % 4])
    else:
        hline(3, row, 12, DIRT)
scatter([(6, 6), (11, 8)], GREENS)
put(3, 11, BUCKET)
put(14, 11, WHEELBARROW)

# grain field
fenced(2, 23, 14, 10)
for row in range(24, 32):
    hline(3, row, 12, GRAIN if row % 2 == 0 else DIRT)
vline(8, 24, 8, DIRT)  # grain blocks movement, so leave a spine to reach the lanes
put(14, 31, HAY)

# paths off the main road into both fields, past the farmhouse and the barn
vline(8, 12, 12, DIRT)
vline(23, 12, 9, DIRT)

building(20, 5, 8, 7)  # farmhouse
put(23, 11, DOORWAY)
put(21, 6, BED_D)          # farmer bedroom
put(22, 6, FLOWERS)
put(25, 7, TABLE)
put(25, 8, CHAIR)
put(26, 9, BASKET)
put(21, 9, BUCKET)

# The barn sits north of the road between the vegetable field and the farmhouse.
# It used to be at (20,25), underneath the cemetery — fenced(19,27,11,11) and its
# grass fill are drawn afterwards and erased everything below y=27, leaving two
# rows of wall and no building. Nothing addresses the barn by name in the routine
# or obligation data, so moving it costs only its zone, its Location entity and
# the waypoint below.
building(9, 13, 8, 7)  # barn
put(12, 19, DOORWAY)   # opens straight onto the main road at y=20
scatter([(10, 14), (11, 14)], HAY)
scatter([(13, 14), (14, 14)], SACK)
put(15, 15, CRATE)
put(14, 17, WHEELBARROW)

put(10, 18, BED_B)     # the farmhand sleeps in the barn with the animals

# The pasture below the grain field. Fenced on the south and east only — the
# herd is kept off the graves and off the road, and the herdsman walks his
# animals in from the north without a gate to open.
hline(9, 37, 10, FENCE_H)
vline(18, 33, 5, FENCE_V)
scatter([(10, 34), (15, 36), (11, 36)], HAY)
put(14, 34, BUCKET)

scatter([(17, 15), (17, 3), (18, 14), (16, 34), (29, 14), (30, 4)], PINE)  # (17,16) is the lunch spot
scatter([(18, 22), (29, 34)], BARE_TREE)
scatter([(19, 18), (31, 13)], BOULDER)
scatter([(17, 19), (30, 25), (44, 22)], PEBBLES)

# --- inn -------------------------------------------------------------------
building(30, 5, 11, 9)
put(36, 13, DOORWAY)
vline(34, 6, 7, WALL_V)     # innkeeper's rooms off the common room
hline(31, 9, 3, WALL_H)
put(34, 7, DOORWAY)         # bedroom
put(34, 11, DOORWAY)        # cellar
put(31, 7, BED_A)           # innkeeper bedroom
put(33, 6, CANDLE)
scatter([(31, 11), (33, 11), (32, 12)], BARREL)   # inn cellar
put(32, 11, BUCKET)
hline(35, 6, 3, COUNTER)
put(38, 6, BREAD)
put(39, 6, FIRE)
put(39, 7, BRAZIER)
scatter([(36, 9), (38, 9)], TABLE)
scatter([(35, 9), (37, 9), (36, 10), (38, 10)], CHAIR)
put(37, 8, MUG)
put(39, 11, BARREL_B)

# --- blacksmith ------------------------------------------------------------
building(44, 5, 9, 8)
put(48, 12, DOORWAY)
vline(48, 6, 3, WALL_V)     # smith storage in the back
hline(49, 9, 3, WALL_H)
put(50, 9, DOORWAY)
scatter([(49, 7), (51, 6)], CRATE)
put(50, 6, IRON_SCRAP)
put(51, 7, SHELF_A)
put(49, 6, TOOL_RACK)
put(45, 6, FIRE)
put(45, 7, COAL)
put(46, 6, BRAZIER)
put(47, 6, TOOL_RACK)
put(46, 8, ANVIL)
put(45, 9, GRINDSTONE)
put(47, 8, TONGS)
put(47, 9, HORSESHOE)
put(51, 11, BED_B)          # blacksmith bedroom
put(50, 11, TABLE)
vline(48, 13, 7, DIRT)

# --- carpenter -------------------------------------------------------------
building(44, 26, 9, 9)
put(48, 26, DOORWAY)
hline(45, 30, 3, WALL_H)    # carpenter bedroom behind the workshop
vline(48, 31, 3, WALL_V)
put(46, 30, DOORWAY)
put(45, 32, BED_C)
put(46, 33, TABLE)
scatter([(46, 28), (47, 28)], BENCH)   # workbench
put(49, 28, SAWHORSE)
put(49, 29, SAW)
scatter([(50, 32), (51, 32)], LOG_PILE)
put(51, 27, SHELF_A)
put(51, 28, SHELF_B)
put(50, 31, TOY)
vline(48, 21, 5, DIRT)

# --- the trader's house ----------------------------------------------------
# Hard against the square, with its door on the main road: the one house in
# town whose business is other people's goods wants to be where they pass.
building(42, 14, 6, 6)
put(44, 19, DOORWAY)
put(43, 15, BED_A)          # trader bedroom
put(46, 15, SHELF_A)
put(46, 16, SHELF_B)
put(45, 17, TABLE)
put(43, 18, CRATE)
put(46, 18, BASKET)

# --- the fisherman's cottage -----------------------------------------------
# Downstream of the mill, with its door opening straight onto the dock.
building(56, 28, 6, 6)
put(61, 30, DOORWAY)
put(57, 29, BED_C)          # fisher bedroom
put(58, 29, TABLE)
put(57, 32, BARREL)         # salted fish
put(60, 32, CRATE)
put(59, 31, BASKET)
scatter([(62, 32), (61, 27)], CRATE)

# --- church and cemetery ---------------------------------------------------
rect(26, 25, 18, 13, RED_GRASS)
building(30, 27, 10, 10)
put(36, 27, DOORWAY)
vline(36, 28, 6, CARPET_V)
put(36, 34, ALTAR)
scatter([(33, 29), (34, 29), (33, 31), (34, 31)], BENCH)
scatter([(38, 29), (38, 31)], BENCH)
scatter([(35, 34), (37, 34)], CANDLE)
put(37, 33, LECTERN)
put(31, 28, BELL)
put(32, 28, BELL_ROPE)
vline(34, 32, 4, WALL_V)    # priest's study and room
hline(31, 31, 3, WALL_H)
put(34, 33, DOORWAY)
put(32, 33, SCROLL)         # the rehearsed sermon
put(31, 33, BOOK)
put(33, 32, CANDLE)
put(32, 35, BED_A)          # priest bedroom

fenced(19, 27, 11, 11)
rect(20, 28, 9, 9, GRASS)
scatter([(21, 29), (23, 29), (26, 29), (28, 30)], GRAVE)
scatter([(21, 32), (27, 32)], GRAVE_PALE)
scatter([(22, 34), (26, 35)], GRAVE_PALE_B)
scatter([(28, 33), (20, 35)], GRAVE_PALE_C)
put(27, 36, TOMB)
put(24, 33, GRAVE)          # the child's grave
put(24, 34, FLOWERS)
vline(24, 21, 8, DIRT)      # path from the road in through the cemetery gate
put(23, 27, GATE)           # the gate itself is solid, so it stands beside the gap
put(25, 27, GATE)

# --- mill ------------------------------------------------------------------
building(54, 8, 8, 8)
put(57, 15, DOORWAY)
scatter([(55, 10), (56, 10)], SACK)
scatter([(55, 12), (56, 12)], FLOUR)
put(60, 10, CHUTE)
put(59, 13, SHELF_A)
put(60, 13, SHELF_B)
put(55, 14, BED_D)          # miller bedroom
put(56, 14, TABLE)
vline(57, 16, 4, DIRT)
put(62, 11, WHEEL_TL)
put(63, 11, WHEEL_TR)
put(62, 12, WHEEL_BL)
put(63, 12, WHEEL_BR)
scatter([(53, 4), (53, 24), (56, 34), (44, 3), (60, 27)], PINE)
scatter([(59, 22), (52, 36)], SHRUB)

# --- named places ----------------------------------------------------------
# Every place the data files can address by name. Routines in townspeople.rec
# and `at:` in obligations.rec take one of these names, and `spawn` is both the
# player's start and the vampire's lair. This dict is the only source: the
# Location entities on the Meta layer are written from it below, and the
# reachability check at the bottom walks to every one of them.
WAYPOINTS = {
    "spawn": (5, 37),
    "vegetable_field": (8, 8), "grain_field": (8, 27), "farmhouse": (24, 8),
    "farmer_bedroom": (22, 7), "barn": (12, 16), "hayloft": (11, 18),
    "pasture": (13, 35), "lunch_tree": (17, 16),
    "inn": (37, 9), "innkeeper_bedroom": (32, 7), "inn_cellar": (33, 11),
    "blacksmith": (46, 10), "smith_storage": (50, 7),
    "blacksmith_bedroom": (50, 11),
    "mill": (58, 11), "mill_interior": (58, 13), "miller_bedroom": (56, 14),
    "carpenter_workshop": (47, 29), "carpenter_interior": (49, 30),
    "carpenter_bedroom": (46, 32),
    "church": (36, 31), "priest_study": (32, 33), "priest_bedroom": (32, 35),
    "cemetery": (22, 31), "childs_grave": (24, 34),
    "town_square": (34, 21), "main_road": (20, 20), "market_stall": (34, 20),
    "trader_house": (44, 17), "trader_bedroom": (43, 15),
    "fisher_cottage": (58, 31), "fisher_bedroom": (57, 29),
    "fishing_spot": (63, 30),
    "bridge": (66, 20), "river": (62, 25), "quiet_riverbank": (69, 30),
}

# --- write -----------------------------------------------------------------
with open(LDTK) as handle:
    project = json.load(handle)

level = next(lvl for lvl in project["levels"] if lvl["identifier"] == "Town")
level["pxWid"], level["pxHei"] = W * CELL, H * CELL

tileset = next(ts for ts in project["defs"]["tilesets"] if ts["identifier"] == "World")
per_row = tileset["pxWid"] // CELL

for layer in level["layerInstances"]:
    layer["__cWid"], layer["__cHei"] = W, H
    if layer["__identifier"] != "Tiles":
        continue
    tiles = []
    for y in range(H):
        for x in range(W):
            t = grid[y][x]
            tiles.append({
                "px": [x * CELL, y * CELL],
                "src": [(t % per_row) * CELL, (t // per_row) * CELL],
                "f": 0,
                "t": t,
                "d": [x + y * W],
                "a": 1,
            })
    layer["gridTiles"] = tiles

# Location entities are rewritten from WAYPOINTS. A place that is already on the
# Meta layer keeps its iid so the LDtk editor and its backups stay pointed at the
# same entity; a new one gets a fresh uuid. Anything else on the Meta layer is
# left alone.
meta = next(l for l in level["layerInstances"] if l["__identifier"] == "Meta")
existing = {}
for entity in meta["entityInstances"]:
    if entity["__identifier"] != "Location":
        continue
    for field in entity["fieldInstances"]:
        if field["__identifier"] == "Name":
            existing[field["__value"]] = entity["iid"]

others = [e for e in meta["entityInstances"] if e["__identifier"] != "Location"]
locations = []
for name, (x, y) in sorted(WAYPOINTS.items()):
    locations.append({
        "__identifier": "Location",
        "__grid": [x, y],
        "__pivot": [0.0, 0.0],
        "__tags": ["META"],
        "__tile": None,
        "__smartColor": "#63C74D",
        "iid": existing.get(name, str(uuid.uuid4())),
        "width": CELL,
        "height": CELL,
        "defUid": 46,
        "px": [x * CELL, y * CELL],
        "fieldInstances": [{
            "__identifier": "Name",
            "__type": "String",
            "__value": name,
            "__tile": None,
            "defUid": 47,
            "realEditorValues": [{"id": "V_String", "params": [name]}],
        }],
    })
meta["entityInstances"] = others + locations

with open(LDTK, "w") as handle:
    json.dump(project, handle, indent=1)

print(f"painted {W}x{H} tiles and {len(locations)} places into {os.path.normpath(LDTK)}")

# --- check: every place a Routine names can be walked to from the lair -------
blocking = set()
for tag in tileset.get("enumTags", []):
    if tag["enumValueId"] in ("IsBlockingMovement", "IsWater"):
        blocking.update(tag["tileIds"])

reached, frontier = set(), [(5, 37)]  # spawn, at the cave mouth
while frontier:
    x, y = frontier.pop()
    if (x, y) in reached or not (0 <= x < W and 0 <= y < H) or grid[y][x] in blocking:
        continue
    reached.add((x, y))
    frontier += [(x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)]

unreachable = sorted(name for name, at in WAYPOINTS.items() if at not in reached)
assert not unreachable, f"unreachable from the lair: {unreachable}"
print(f"all {len(WAYPOINTS)} waypoints reachable")

# --- check: every building still has its walls ------------------------------
# The barn used to sit under the cemetery, which is drawn later and erased all
# but two rows of it — a whole building quietly reduced to a sliver, and nothing
# said so. Anything painted over a wall after the fact now fails here.
WALLISH = {WALL_CORNER, WALL_V, WALL_H, WINDOW_H, WINDOW_V, DOORWAY}
breached = []
for bx, by, bw, bh in BUILDINGS:
    ring = (
        [(cx, by) for cx in range(bx, bx + bw)]
        + [(cx, by + bh - 1) for cx in range(bx, bx + bw)]
        + [(bx, cy) for cy in range(by, by + bh)]
        + [(bx + bw - 1, cy) for cy in range(by, by + bh)]
    )
    holes = [(cx, cy) for cx, cy in ring if grid[cy][cx] not in WALLISH]
    if holes:
        breached.append(((bx, by, bw, bh), holes[:6]))
assert not breached, f"buildings with walls painted over: {breached}"
print(f"all {len(BUILDINGS)} buildings have intact walls")

# --- check: every building still has a window -------------------------------
# Doors, interior walls and props are all painted over the shell after it is
# drawn, so a window can be overwritten after the fact. Without this check that
# happens silently and a house quietly becomes unwatchable.
windowless = []
for bx, by, bw, bh in BUILDINGS:
    panes = sum(
        1
        for cy in range(by, by + bh)
        for cx in range(bx, bx + bw)
        if grid[cy][cx] in (WINDOW_H, WINDOW_V)
    )
    if panes == 0:
        windowless.append((bx, by, bw, bh))
assert not windowless, f"buildings with no window left: {windowless}"
print(f"all {len(BUILDINGS)} buildings have windows")
