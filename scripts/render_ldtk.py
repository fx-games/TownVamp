#!/usr/bin/env python3
"""Render an LDtk level's tile layers to a PNG so an agent can look at the map.

    python3 scripts/render_ldtk.py [out.png] [level_index] [scale]
"""
import json
import sys

from PIL import Image

MAP = "files/assets/maps.ldtk"

out_path = sys.argv[1] if len(sys.argv) > 1 else "/tmp/town.png"
level_index = int(sys.argv[2]) if len(sys.argv) > 2 else 0
scale = int(sys.argv[3]) if len(sys.argv) > 3 else 4

project = json.load(open(MAP))
level = project["levels"][level_index]
tileset = Image.open("files/assets/" + project["defs"]["tilesets"][0]["relPath"]).convert("RGBA")
grid = project["defs"]["tilesets"][0]["tileGridSize"]

out = Image.new("RGBA", (level["pxWid"], level["pxHei"]), (255, 255, 255, 255))
for layer in reversed(level["layerInstances"]):  # LDtk stores topmost layer first
    for tile in layer.get("gridTiles", []) + layer.get("autoLayerTiles", []):
        sx, sy = tile["src"]
        out.paste(tileset.crop((sx, sy, sx + grid, sy + grid)), tuple(tile["px"]))
out.resize((out.width * scale, out.height * scale), Image.NEAREST).save(out_path)
print(out_path, out.width, out.height, level["identifier"])
