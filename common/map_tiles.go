package common

import (
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/ldtk"
)

// MapTileFromLDTK translates a tile from an LDtk layer into a map tile. It is the
// only part of the old map generator the town loader needed; everything else in
// that file generated dungeons.
func MapTileFromLDTK(tileSet *ldtk.Tileset, layer *ldtk.Layer, tile *ldtk.Tile) (geometry.Point, gridmap.Tile) {
	posX, posY := layer.ToGridPosition(tile.Position[0], tile.Position[1])
	pos := geometry.Point{X: posX, Y: posY}
	enums := tileSet.EnumsForTile(tile.ID)

	walkable := !enums.Contains("IsBlockingMovement")
	transparent := !enums.Contains("IsBlockingView")
	isDamaging := enums.Contains("IsDamaging")
	isWater := enums.Contains("IsWater")
	specialTile := gridmap.SpecialTileNone

	if isWater {
		specialTile = gridmap.SpecialTileWater
	} else if isDamaging {
		specialTile = gridmap.SpecialTileDamaging
	}

	tileType := gridmap.Tile{
		DefinedIcon:   int32(tile.ID),
		IsWalkable:    walkable,
		IsTransparent: transparent,
		Special:       specialTile,
	}
	return pos, tileType
}
