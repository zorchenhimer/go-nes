package image

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

// Data is a list of palette indexes.  One ID per pixel.  A single tile is
// always 8x8 pixels.  Larger meta tiles (eg, 8*16) will be made up of multiple
// tiles of 64 total pixels.
// type Tile [64]byte
type Tile struct {
	image.Paletted
	OrigId    int
	charWidth int
	PaletteId int // 0-3
	bgIndex   uint8
}

func NewTile(id int) *Tile {
	return &Tile{
		Paletted: image.Paletted{
			Pix:     make([]uint8, 64),
			Stride:  8,
			Rect:    image.Rect(0, 0, 8, 8),
			Palette: DefaultPalette,
		},
		OrigId:    id,
		charWidth: -1,
	}
}

func (t *Tile) FillBackground(index uint8) {
	index = index % 4
	for i := 0; i < 64; i++ {
		t.Pix[i] = index
	}
}

func (t *Tile) SetBackgroundIndex(index uint8) {
	t.bgIndex = index
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

const ( // currently unused
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

func (t *Tile) IsEmpty() bool {
	return !t.IsNotEmpty()
}

func (t *Tile) IsNotEmpty() bool {
	for _, p := range t.Pix {
		if p > 0 {
			return true
		}
	}
	return false
}

func (t *Tile) CharacterWidth() int {
	if t.charWidth > -1 {
		return t.charWidth
	}

LOOP:
	for col := 7; col > -1; col-- {
		for row := 0; row < 8; row++ {
			pix := t.Paletted.ColorIndexAt(col, row)
			//fmt.Printf("%s\n(%d,%d)\n%d\n\n", t.ASCII(), row, col, pix);
			if pix != t.bgIndex {
				t.charWidth = col + 1
				break LOOP
			}
		}
	}

	return t.charWidth
}

func (t *Tile) getChrBin() ([]byte, []byte) {
	plane1 := []byte{}
	plane2 := []byte{}
	for row := 0; row < 8; row++ {
		var p1 uint8 = 0
		var p2 uint8 = 0
		for col := 0; col < 8; col++ {
			color := t.Pix[col+(row*8)]
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
	plane1, plane2 := t.getChrBin()

	p1 := []string{}
	p2 := []string{}

	formatString := "$%02X"
	if binary {
		formatString = "%%%08b"
	}

	for _, b := range plane1 {
		p1 = append(p1, fmt.Sprintf(formatString+"; $%02X", b, b))
	}

	if !half {
		for _, b := range plane2 {
			p2 = append(p2, fmt.Sprintf(formatString+"; $%02X", b, b))
		}
	}

	if half {
		return ".byte " + strings.Join(p1, "\n.byte ") + "\n"
	}
	return ".byte " + strings.Join(p1, ", ") + "\n.byte " + strings.Join(p2, ", ")
}

// Chr returns a slice of bytes that contain both the bit planes of
// a tile's CHR data.
func (t *Tile) Chr(firstPlane bool) []byte {
	plane1, plane2 := t.getChrBin()
	if firstPlane {
		return plane1
	}
	return append(plane1, plane2...)
}

// image.Image implementation
func (t *Tile) ColorModel() color.Model {
	return NESModel
}

func (t *Tile) Bounds() image.Rectangle {
	return image.Rect(0, 0, 8, 8)
}

func (t *Tile) At(x, y int) color.Color {
	// x = 0
	// y = 1
	// idx = (0 * 8) + 1
	val := t.Pix[(y*8)+x]
	return t.Palette[val]
}

// drawer.Image implementation
func (t *Tile) Set(x, y int, c color.Color) {
	palcolor := t.Palette.Convert(c)

	idx, err := getColorIndex(t.Palette, palcolor)
	if err != nil {
		// Don't panic, just use the first color.
		fmt.Printf("WARNING: Set() trying to use a color not in the palette!")
		idx = 0
	}

	t.SetPaletteIndex(x, y, uint8(idx))
}

// SetPaletteIndex works like Set(), but uses an index instead of a color as input.
func (t *Tile) SetPaletteIndex(x, y int, idx uint8) {
	if int(idx) > len(t.Palette) {
		// Don't panic, just use the first color.
		fmt.Printf("WARNING: SetPaletteIndex() trying to use a color not in the palette!")
		idx = 0
	}
	t.Pix[(x*8)+y] = uint8(idx)
}
