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

func (thisTile *Tile) IsIdentical(otherTile *Tile) bool {
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

func (t *Tile) getChrBin() ([]byte, []byte) {
	plane1 := []byte{}
	plane2 := []byte{}
	for row := 0; row < 8; row++ {
		var p1 uint8 = 0
		var p2 uint8 = 0
		for col := 0; col < 8; col++ {
			color := t.Pix[col + (row * 8)]
			switch color {
			case 1:
				p1 = (p1 << 1) | 1
				p2 = (p2 << 1)

			case 2:
				p1 = (p1 << 1)
				p2 = (p2 << 1) | 1

			case 3:
				p1 = (p1 << 1) | 1
				p2 = (p2 << 1) | 1

			default:
				p1 = (p1 << 1)
				p2 = (p2 << 1)
			}
		}
		plane1 = append(plane1, p1)
		plane2 = append(plane2, p2)
	}
	return plane1, plane2
}

// Asm takes two parameters:
// Setting `half` to true will only return the first bitplane of the tile.
// Setting `binary` to true will use the binary notation, otherwise it will
// use the hexidecimal notation.
func (t *Tile) Asm(half, binary bool) string {
	plane1, plane2 := getChrBin

	p1 := []string{}
	p2 := []string{}

	formatString := "$%02X"
	if binary {
		formatString = "%%%08b"
	}

	for _, b := range plane1 {
		p1 = append(p1, fmt.Sprintf(formatString, b))
	}

	if !half {
		for _, b := range plane2 {
			p2 = append(p2, fmt.Sprintf(formatString, b))
		}
	}

	if half {
		return ".byte " + strings.Join(p1, ", ")
	}
	return ".byte " strings.Join(p1, ", ") + "\n.byte" + strings.Join(p2, ", ")
}

func (t *Tile) Chr() []byte {
	plane1, plane2 := t.getChrBin()
	return append(plane1, plane2...)
}
