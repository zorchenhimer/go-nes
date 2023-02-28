package image

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

type Arrangement uint

const (
	ARR_SINGLE  Arrangement = iota
	ARR_DBLHIGH             // 8x16 sprite mode
	// TODO: more?
)

// A PatternTable is the data that will be written as the CHR.  It can only be in
// 1k, 2k, 4k, or 8k sizes; or 64, 128, 256, or 512 tiles; or 128x32, 128x64,
// 128x128, or 128x256 pixels; or 4, 8, 16, or 32 rows of tiles.
type PatternTable struct {
	Patterns     []*Tile
	Layout       Arrangement
	ReducedIds   []int

	// Dimensions are in pixels
	SourceWidth  int
	SourceHeight int
}

//type TableSize int
//const (
//	TS_64	TableSize = TableSize(64)
//	TS_128	TableSize = TableSize(128)
//	TS_256	TableSize = TableSize(256)
//	TS_512	TableSize = TableSize(512)
//)

//func (ts TableSize) String() string {
//	switch ts {
//	case TS_64:
//		return "1k (64 tiles)"
//	case TS_128:
//		return "2k (128 tiles)"
//	case TS_256:
//		return "4k (256 tiles)"
//	case TS_512:
//		return "8k (512 tiles)"
//	}
//}

func NewPatternTable() *PatternTable {
	return &PatternTable{
		Patterns: []*Tile{},
	}
}

func (pt *PatternTable) Debug() string {
	//return fmt.Sprintf("Size: %s", pt.Size)
	l := float64(len(pt.Patterns))
	row := int(math.Ceil(l / 16.0))
	return fmt.Sprintf("%d tiles; %d rows", int(l), row)
}

func (pt *PatternTable) AddTile(tile *Tile) {
	pt.Patterns = append(pt.Patterns, tile)
}

func (pt *PatternTable) AddPatternTable(newPt *PatternTable) {
	for _, tile := range newPt.Patterns {
		pt.AddTile(tile)
	}
}

// Returns before/after count
func (pt *PatternTable) RemoveEmpty() (int, int) {
	tiles := []*Tile{}
	pt.ReducedIds = []int{}
	for idx, tile := range pt.Patterns {
		if tile.IsEmpty() {
			continue
		}

		pt.ReducedIds = append(pt.ReducedIds, idx)
		tiles = append(tiles, tile)
	}

	pt.Patterns = tiles
	return len(pt.ReducedIds), len(pt.Patterns)
}

// Returns before/after count
func (pt *PatternTable) RemoveDuplicates(removeEmpty bool) (int, int) {
	tiles := []*Tile{}      // unique tiles
	pt.ReducedIds = []int{} // idx is orig ID, value is new tile's old id

OUTER:
	for _, tile := range pt.Patterns {
		if removeEmpty && tile.IsEmpty() {
			continue
		}

		for i, t := range tiles {
			if t.IsIdentical(tile) {
				pt.ReducedIds = append(pt.ReducedIds, i)
				continue OUTER
			}
		}

		pt.ReducedIds = append(pt.ReducedIds, len(tiles))
		tiles = append(tiles, tile)
	}

	pt.Patterns = tiles
	return len(pt.ReducedIds), len(pt.Patterns)
}

// Chr returns the pattern table data as bytes in the CHR format.
func (pt *PatternTable) Chr(firstPlane bool) []byte {
	chr := []byte{}

	for _, t := range pt.Patterns {
		chr = append(chr, t.Chr(firstPlane)...)
	}

	return chr
}

// PadTiles ensures that all rows have 16 tiles in them.
func (pt *PatternTable) PadTiles() {
	emptyTile := NewTile(0)
	for len(pt.Patterns)%16 != 0 {
		pt.Patterns = append(pt.Patterns, emptyTile)
	}
}

// PadTileCount ensures that at least count number of tiles exist
// in the pattern table
func (pt *PatternTable) PadTileCount(count int) {
	emptyTile := NewTile(0)
	for len(pt.Patterns) < count {
		pt.Patterns = append(pt.Patterns, emptyTile)
	}
}

func (pt *PatternTable) SetPalette(pal color.Palette) {
	for _, t := range pt.Patterns {
		t.Palette = pal
	}
}

// WriteFile writes the CHR data of all the tiles in assembly format to
// the given file.
func (pt *PatternTable) Asm(firstPlane bool) string {
	var chrStr strings.Builder
	var idsStr strings.Builder

	for _, t := range pt.Patterns {
		//chr = append(chr, t.Chr()...)
		// TODO: pass --first-plane here and add a --hex or --binary flag for other param
		chrStr.WriteString(t.Asm(firstPlane, true) + "\n")
	}

	// Only write tile IDs if duplicates have been removed
	//if pt.ReducedIds != nil {
	//	line := []string{}
	//	for i := 0; i < len(pt.ReducedIds); i++ {
	//		line = append(line, fmt.Sprintf("$%02X", pt.ReducedIds[i]))
	//		if i%32 == 0 && i != 0 {
	//			fmt.Fprintf(&idsStr, ".byte %s\n", strings.Join(line, ", "))
	//			line = []string{}
	//		}
	//	}
	//	fmt.Fprintf(&idsStr, ".byte %s\n", strings.Join(line, ", "))
	//}

	if idsStr.Len() > 0 {
		return chrStr.String() + "\n" + idsStr.String()
	}
	return chrStr.String()
}

// Implement the image.Image interface
func (pt *PatternTable) ColorModel() color.Model {
	return NESModel
}

func (pt *PatternTable) Bounds() image.Rectangle {
	// TODO
	width := 128
	if len(pt.Patterns) < 16 {
		width = len(pt.Patterns) * 8
	}
	height := int(math.Ceil(float64(len(pt.Patterns))/16.0)) * 8
	return image.Rect(0, 0, width, height)
}

// TODO: Verify this actually works
func (pt *PatternTable) At(x, y int) color.Color {
	tile := pt.getTileAtCoord(x, y)
	// Get the pixel in the tile
	x = x % 8
	y = y % 8
	return tile.At(x, y)
}

// Implement image.draw.Drawer and image.draw.Image
//func (pt *PatternTable) Draw(dst image.Image, r image.Rectangle, src image.Image, sp image.Point) {
//	// This would require splitting the source image into tiles, then drawing
//	// on each affected tile.
//	panic("PatternTable.Draw() is not implemented")
//}

// returns the tile that contains the pixel at the X/Y coordinate
func (pt *PatternTable) getTileAtCoord(x, y int) *Tile {
	// This is integer division.  The results are
	// automatically floor()'d.
	// Row and column of the Tile, given the x, y of the whole table
	col := x / 8
	row := y / 8

	// Tile index
	idx := (row * 16) + col

	// Get the tile
	if idx >= len(pt.Patterns) || idx == -1 {
		panic(fmt.Sprintf("getTileAtCoord() index too large: %d/%d (%d, %d) col/row:(%d, %d) rect: %v",
			idx,
			len(pt.Patterns)-1, // last idx of this slice
			x, y,
			col, row,
			pt.Bounds(),
		))
	}
	return pt.Patterns[idx]
}

// TODO: Verify this works
func (pt *PatternTable) Set(x, y int, c color.Color) {
	tile := pt.getTileAtCoord(x, y)

	x = x % 8
	y = y % 8
	tile.Set(x, y, c)
}
