# TownVamp

A real-time town simulation in Go (Ebitengine) where you play a vampire feeding on a town of nine Townspeople, unseen. Every death breaks the duties (Obligations) the town runs on, and the Epilogue reports what followed.

Status: proof of concept, work in progress. See `docs/specs/town-vampire-poc.md` and `CONTEXT.md` for the design and glossary.

## Build and run

Requires Go (see `go.mod`) and the Ebitengine platform dependencies (CGO).

```sh
./build_n_run.sh          # builds builds/dev/LuxAeterna and runs it
# or
go build -tags "ebitensinglethread development" -o builds/dev/LuxAeterna .
./builds/dev/LuxAeterna
```

Tests: `go test ./...`

Other scripts: `build_release.sh` (mac/win/linux zips; uses a local mingw path), `build_embedded.sh` / `build_linux_embedded.sh` (assets embedded via the `embedded` tag), `build_web.sh [outdir]` (WASM build).

## Controls

- W/A/S/D: walk (hold)
- Space: pause / resume
- E: kill; H: hide body
- 1-6: ask about the Townsperson with that number
- `,` / `.`: wait until dusk / dawn
- L: log; C: stats; X: center on player; B: flags
- End: leave town (ends the run)
- Esc: cancel; F1: help; F11: fullscreen

## Layout

- `game/` town simulation: routines, obligations, reassignment, vampire, hunters, epilogue (plus tests)
- `common/`, `defs/`, `stats/`, `container/` shared engine types, actions and data definitions
- `gridmap/`, `geometry/`, `ldtk/` map, field of view, pathfinding, LDtk loading
- `renderer/`, `gui/`, `console/`, `audio/`, `util/` rendering, UI, input, sound
- `files/` assets (`files/assets/maps.ldtk` is the town map)
- `scripts/` Python helpers for building and rendering the map
- `docs/` specs, ADRs, tileset notes; `CONTEXT.md` glossary; `AGENTS.md` LDtk MCP notes
