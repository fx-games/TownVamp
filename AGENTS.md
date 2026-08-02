# Agent notes

## LDtk MCP server (`ldtk`)

Edits `files/assets/maps.ldtk` (LDtk JSON v1.5.3) directly. Registered in
[.mcp.json](.mcp.json); binary built from https://github.com/gazure/ldtk-mcp at
`/Users/felix/Projects/ldtk-mcp/target/release/ldtk-mcp` (rebuild: `cargo build --release`).
Needs a session restart to appear.

### Workflow

`open_project` (path `files/assets/maps.ldtk`) → `describe_defs` → read/edit → `validate_project`
→ `save_project`. Edits stay in memory until `save_project`. LDtk GUI must reload the file after.

### This project's map

- One world, `Free` layout, one level: `Town` (464×200 px at world -912,-1104), no external levels.
- Layers: `Meta` (Entities), `Entities` (Entities), `Tiles` (Tiles). **No IntGrid and no AutoLayer**
  — so `set_intgrid` / `flood_fill_intgrid` do not apply; tiles are painted directly with
  `paint_tiles` / `clear_tiles`.
- Tileset `World` = `classic_roguelike.png`, 8×8 tiles. Grid size 8 everywhere; entity grid
  coordinates are 8 px cells.
- Entity defs and fields:
  - `NPC` (tag NPC) — `Name`, `InternalName`, `UniqueName`, `ASCII` (String), `Tile` (Tile),
    `Type` (enum `NPCCategory`), `FlipLookDirection` (Bool). NPCs live on the `Entities` layer.
  - `Location` (tag META) — `Name`. On the `Meta` layer.
  - `Sign` (tag OBJECT) — `Text`. `Shrine`, `Door` (tag OBJECT) — no fields.
- Enums: `NPCCategory`, `TilesetMeta` (IsBlockingMovement / IsBlockingView / IsWater / IsDamaging),
  `TemplateAreas`.

### Tile ids and names

Tile id = `row * 28 + col` (tileset is 28×28 tiles = 784, 8 px each). Names for ~200 ids already
exist as Go constants in [defs/definedIcons.go](defs/definedIcons.go) — reuse those names, they are
what the game code refers to. Blocking behaviour is not in the name: it comes from the tileset's
`enumTags` in the `.ldtk` (`IsBlockingMovement`, `IsBlockingView`, `IsWater`, `IsDamaging`), so
check the tag before painting something as a wall.

The tiles currently used in `Town`, with what they actually look like:

| id | name | look / use | blocks |
| --- | --- | --- | --- |
| 0 | `IconGrass` | grass | no |
| 1 | `IconDirt` | dirt, road, square | no |
| 4 | `IconTreeLight` | pine tree | move+view |
| 5 | `IconTreeDark` | bare tree | move+view |
| 6 | — shrub | small green sprout | move |
| 7 | `IconMountain` | mountain, map border | move+view |
| 8 | `IconMountainWithEntrance` | cave mouth | no |
| 10 | — boulder | big round rock | no |
| 11 | — pebbles | two small rocks | no |
| 12 | — floor block | stone block, shop interiors | no |
| 28 | `IconWater` | solid block, tagged `IsWater` | move |
| 56 | `IconCathedralWall` | wall corner | move+view |
| 57 | — wall vertical | left/right building edge | move+view |
| 58 | `IconCathedralWallWithShading` | wall horizontal, top/bottom edge | move+view |
| 63 | `IconWoodenBridge` | planks — bridge, dock | no |
| 64–67 | `IconCathedralFloor` + variants | flagstone floor | no |
| 73 | `IconStairsDown` | stairs down | no |
| 74 | `IconStairsUp` | stairs up | no |
| 75 | `IconShrine` | shrine / font | no |
| 109 | `IconEffectFireSmall` | flame, used as brazier | no |
| 114 | — sword sign | shop sign in wall (weapons) | move+view |
| 116 | — shield sign | shop sign in wall (armour) | move+view |
| 117 | — hammer sign | shop sign in wall (smith) | move+view |
| 119 | — pickaxe sign | shop sign in wall | move+view |
| 180 | — red grass | flower field around the church | no |
| 234 | — doorway with carpet | church door | no |
| 235 | — bare tree on red grass | single decorative tile | no |
| 236 | — floor, carpet horizontal | church aisle | no |
| 237 | — floor, carpet vertical | church aisle | no |
| 239 | — altar | red altar with cross | no |

Wider inventory of what the tileset can and cannot depict for the PoC:
[docs/tileset-coverage.md](docs/tileset-coverage.md).

To look at a map after editing it:

```bash
python3 scripts/render_ldtk.py /tmp/town.png
```

### Gotchas

- The Go side loads this file at `assets/maps.ldtk` ([main.go:147](main.go:147), and
  [game/harness_test.go:34](game/harness_test.go:34)) — a broken save breaks the game and the tests.
  Run `go test ./game/...` after structural edits.
- LDtk writes its own backups to `files/assets/maps/backups/`; those are LDtk's, not a safety net
  for MCP edits. `maps.ldtk` is git-tracked — check `git diff` before committing a generated edit.
