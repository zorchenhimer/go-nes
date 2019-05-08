package image

import (
	"fmt"
	"image"
	"image/color"
	//"strings"
)

// This is that ugly palette from YYCHR
var DefaultPalette color.Palette = color.Palette{
	color.RGBA{R: 0x00, G: 0x39, B: 0x73, A: 0xFF},
	color.RGBA{R: 0x84, G: 0x5E, B: 0x21, A: 0xFF},
	color.RGBA{R: 0xAD, G: 0xB5, B: 0x31, A: 0xFF},
	color.RGBA{R: 0xC6, G: 0xE7, B: 0x9C, A: 0xFF},
}

// Assembled meta sprites and tiles.  These will be unwrapped to the specified
// layout (eg, 8*16 vs 8x8)
type MetaTile struct {
	Tiles []*Tile

	// Width and Hight in tiles, not pixels
	Width  int
	Height int

	// Layout of tiles in the destination CHR
	Layout TileLayout
}

// Data is a list of palette indexes.  One ID per pixel.  A single tile is
// always 8x8 pixels.  Larger meta tiles (eg, 8*16) will be made up of multiple
// tiles of 64 total pixels.
//type Tile [64]byte
type Tile struct {
	image.Paletted
	OrigId int
}

func NewTile(id int) *Tile {
	return &Tile{
		Paletted: image.Paletted{
			Pix:     make([]uint8, 64),
			Stride:  8,
			Rect:    image.Rect(0, 0, 8, 8),
			Palette: DefaultPalette,
		},
		OrigId: id,
	}
}

func (thisTile *Tile) IsIdentical(otherTile Tile) bool {
	for i := 0; i < 64; i++ {
		if thisTile.Pix[i] != otherTile.Pix[i] {
			return false
		}
	}
	return true
}

// Ideally, each tile or object will be in its own input file and is assembled
// into the final CHR layout during assemble time.
type TileLayout int

const (
	TL_SINGLE = iota // Default.  A single 8x8 tile.
	TL_8X16          // 8x16 sprites.
	TL_ROW           // Row sequential
	TL_COLUMN        // Column sequential
	TL_ASIS          // Don't transform.  This will break things if there's meta tiles that are not the same size.
)

func (t *Tile) ASCII() string {
	//chars := [64]rune{}
	chars := ""
	for i, t := range t.Pix {
		c := ""
		switch t {
		case 0:
			c = "_"
		case 1:
			c = "|"
		case 2:
			c = "X"
		case 3:
			c = "O"
		}
		if i%8 == 0 {
			chars = fmt.Sprintf("%s\n", chars)
		}
		chars = fmt.Sprintf("%s%s", chars, c)
	}

	return fmt.Sprintf("%s", chars)
}

func (t *Tile) Asm() string {
	//chars := [64]rune{}
	charsA := ""
	charsB := ""
	for i, t := range t.Pix {
		ca := ""
		cb := ""
		switch t {
		case 0:
			ca = "0"
			cb = "0"
		case 1:
			ca = "1"
			cb = "0"
		case 2:
			ca = "0"
			cb = "1"
		case 3:
			ca = "1"
			cb = "1"
		}
		if i%8 == 0 {
			charsA = fmt.Sprintf("%s\n    .byte %%", charsA)
			charsB = fmt.Sprintf("%s\n    .byte %%", charsB)
		}
		charsA = fmt.Sprintf("%s%s", charsA, ca)
		charsB = fmt.Sprintf("%s%s", charsB, cb)
	}

	//return fmt.Sprintf("    .byte %%%s", chars)
	return charsA + "\n" + charsB
}
