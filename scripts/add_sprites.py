#!/usr/bin/env python3
"""Paint the town-PoC sprites into empty slots of files/assets/classic_roguelike.png.

The art is the ASCII below; run this to (re)generate the tiles. Idempotent — it always writes
the same pixels to the same tile ids. Palette indices only, so the PNG stays 8-bit with index 0
transparent, exactly as the rest of the sheet.

    python3 scripts/add_sprites.py
"""
import sys

from PIL import Image

SHEET = "files/assets/classic_roguelike.png"
COLS = 28

# palette index per art character (see the sheet's own palette)
INK = {
    ".": 0,  # transparent
    "w": 19,  # wood
    "l": 18,  # wood, light
    "W": 21,  # wood, near black — grooves and outlines
    "k": 20,  # wood, dark — iron rims, wheels
    "i": 6,  # iron
    "I": 5,  # iron, light
    "d": 7,  # iron, shadow
    "y": 25,  # brass / grain
    "o": 24,  # flame
    "f": 15,  # flour, cloth
    "s": 17,  # rope
    "g": 3,  # glass
    "G": 4,  # glass, glint
}

SPRITES = {
    728: ("anvil", [
        "........",
        ".IIIIII.",
        "IIIIIIII",
        "..dddd..",
        "..ddd...",
        ".ddddd..",
        ".dddddd.",
        "........",
    ]),
    729: ("grindstone", [
        "..iiii..",
        ".iiiiii.",
        ".ii..ii.",
        ".ii..ii.",
        ".iiiiii.",
        "..iiii..",
        ".w.ww.w.",
        "w..ww..w",
    ]),
    730: ("tool rack", [
        "wwwwwwww",
        ".k....k.",
        ".III.i.i",
        "..w..i.i",
        "..w..ii.",
        "..w...i.",
        "......i.",
        "........",
    ]),
    731: ("tongs", [
        ".i....i.",
        ".i....i.",
        "..i..i..",
        "...ii...",
        "...dd...",
        "..i..i..",
        ".i....i.",
        ".i....i.",
    ]),
    732: ("horseshoe", [
        "........",
        ".ii..ii.",
        ".i....i.",
        ".i....i.",
        ".i....i.",
        "..i..i..",
        "..iiii..",
        "........",
    ]),
    735: ("sawhorse", [
        "........",
        "wwwwwwww",
        "..w..w..",
        "..w..w..",
        ".w....w.",
        ".w....w.",
        "w......w",
        "........",
    ]),
    736: ("saw", [
        "......II",
        ".....II.",
        "....II..",
        "...II...",
        "..II....",
        ".wI.....",
        "ww......",
        ".w......",
    ]),
    737: ("log pile", [
        "........",
        ".ll..ll.",
        ".lWl.lWl",
        ".ll..ll.",
        "...ll...",
        "..lWl...",
        "...ll...",
        "........",
    ]),
    738: ("wooden toy", [
        "........",
        "..ww....",
        ".wwww...",
        "..wwwww.",
        ".wwwwww.",
        "..w..w..",
        "..w..w..",
        "........",
    ]),
    739: ("chair", [
        "..ww....",
        "..ww....",
        "..ww....",
        "..wwwww.",
        "..lllll.",
        "..w..w..",
        "..w..w..",
        "........",
    ]),
    740: ("counter", [
        "........",
        "llllllll",
        "wwwwwwww",
        "w.ww.ww.",
        "wwwwwwww",
        "w.ww.ww.",
        "wwwwwwww",
        "kkkkkkkk",
    ]),
    741: ("mug", [
        "........",
        ".ffff.i.",
        ".IIII.i.",
        ".IIIIii.",
        ".IIII.i.",
        ".IIII...",
        ".IIII...",
        "........",
    ]),
    742: ("bucket", [
        "..iiii..",
        ".i....i.",
        "llllllll",
        ".wwwwww.",
        ".wwwwww.",
        ".wwwwww.",
        "..wwww..",
        "........",
    ]),
    743: ("lectern", [
        "....lll.",
        "...llll.",
        "..lllll.",
        ".llllll.",
        "...ww...",
        "...ww...",
        "..wwww..",
        ".wwwwww.",
    ]),
    744: ("bell", [
        "...kk...",
        "..yyyy..",
        ".yyyyyy.",
        ".yyyyyy.",
        "yyyyyyyy",
        "y.yyyy.y",
        "...WW...",
        "........",
    ]),
    # Windows. A wall you can see through but not walk through — the one thing
    # the tile flags model that nothing else in the town uses. Tag these
    # IsBlockingMovement and NOT IsBlockingView in the LDtk tileset editor;
    # getting that backwards makes them ordinary walls and the whole point of
    # them is lost. Both keep the brick rhythm of the wall they sit in (58
    # horizontal, 57 vertical) so a window reads as part of the building.
    763: ("window, horizontal wall", [
        "........",
        ".IIIII.I",
        "..kkkk..",
        ".IkggkII",
        "..kgGk..",
        ".dkkkkdd",
        "........",
        ".d.ddddd",
    ]),
    764: ("window, vertical wall", [
        "........",
        ".I.....I",
        "..kkkk..",
        ".Ikggk.I",
        "..kgGk..",
        ".Ikkkk.I",
        "........",
        ".I.I...I",
    ]),
    753: ("bell rope", [
        "...ss...",
        "...ss...",
        "..ssss..",
        "...ss...",
        "...ss...",
        "..ssss..",
        "...ss...",
        "...ss...",
    ]),
    745: ("candle", [
        "...o....",
        "..ooo...",
        "...y....",
        "...f....",
        "..fff...",
        "..fff...",
        "..fff...",
        ".lllll..",
    ]),
    746: ("fence, horizontal", [
        "........",
        "w..w..w.",
        "wwwwwwww",
        "w..w..w.",
        "wwwwwwww",
        "w..w..w.",
        "w..w..w.",
        "........",
    ]),
    747: ("fence, vertical", [
        "..w.w...",
        "..w.w...",
        ".wwwww..",
        "..w.w...",
        "..w.w...",
        ".wwwww..",
        "..w.w...",
        "..w.w...",
    ]),
    748: ("gate", [
        "........",
        "ww....ww",
        "wwwwwwww",
        "wWw..wWw",
        "wwwwwwww",
        "w.Ww.W.w",
        "ww....ww",
        "........",
    ]),
    749: ("wheelbarrow", [
        "........",
        "llllll..",
        "wwwwww.w",
        ".wwww..w",
        "..ww....",
        ".kk.....",
        "kkkk....",
        ".kk.....",
    ]),
    750: ("bread", [
        "........",
        "..llll..",
        ".llllll.",
        "lWlWlWll",
        "llllllll",
        ".llllll.",
        "..llll..",
        "........",
    ]),
    751: ("grain chute", [
        "wwww....",
        "w.ww....",
        ".w.ww...",
        "..w.ww..",
        "...w.ww.",
        "....y.ww",
        "...yy...",
        "..y.y...",
    ]),
    752: ("flour sack", [
        "...kk...",
        "..ffff..",
        ".ffffff.",
        "ffffffff",
        "ffffffff",
        "ffffffff",
        ".ffffff.",
        "........",
    ]),
}

# The waterwheel is 2x2 tiles (16x16 px): tiles 733/734 on top, 761/762 below.
WHEEL_TOPLEFT = 733


def wheel_pixels():
    """16x16 wheel: rim, eight paddles, eight spokes, hub."""
    import math

    px = [[0] * 16 for _ in range(16)]
    cx = cy = 7.5
    for y in range(16):
        for x in range(16):
            dx, dy = x - cx, y - cy
            r = math.hypot(dx, dy)
            angle = math.atan2(dy, dx)
            spoke = min(abs(((angle + math.pi / 8) % (math.pi / 4)) - math.pi / 8), 1.0)
            if 6.0 <= r <= 7.6:  # rim
                px[y][x] = INK["k"]
            elif r > 7.6 and spoke * r < 0.9:  # paddles poking out past the rim
                px[y][x] = INK["l"]
            elif r < 6.0 and spoke * r < 0.7:  # spokes
                px[y][x] = INK["w"]
            elif r <= 1.6:  # hub
                px[y][x] = INK["W"]
    return px


def blit(image, tile_id, pixels):
    ox, oy = (tile_id % COLS) * 8, (tile_id // COLS) * 8
    for y, row in enumerate(pixels):
        for x, value in enumerate(row):
            image.putpixel((ox + x, oy + y), value)


def main():
    sheet = Image.open(SHEET)
    assert sheet.mode == "P", f"expected a palette image, got {sheet.mode}"
    before = sheet.size

    for tile_id, (_name, art) in SPRITES.items():
        assert len(art) == 8 and all(len(r) == 8 for r in art), f"tile {tile_id} is not 8x8"
        blit(sheet, tile_id, [[INK[c] for c in row] for row in art])

    wheel = wheel_pixels()
    for dy in (0, 1):
        for dx in (0, 1):
            tile_id = WHEEL_TOPLEFT + dx + dy * COLS
            blit(sheet, tile_id, [row[dx * 8:dx * 8 + 8] for row in wheel[dy * 8:dy * 8 + 8]])

    assert sheet.size == before, "sheet must not change size — tile ids are hard-coded"
    for tile_id in list(SPRITES) + [WHEEL_TOPLEFT + d for d in (0, 1, COLS, COLS + 1)]:
        ox, oy = (tile_id % COLS) * 8, (tile_id // COLS) * 8
        assert len(sheet.crop((ox, oy, ox + 8, oy + 8)).getcolors(64)) > 1, f"tile {tile_id} is blank"

    sheet.save(SHEET, transparency=0, optimize=False)
    print(f"wrote {len(SPRITES) + 4} tiles to {SHEET}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
