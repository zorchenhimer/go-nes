package image

import (
	"fmt"
	"os"
)

/*
CHR

	each byte is a single row of pixels
	each tile is 16 bytes long
	first 8 bytes are the first half of the plane	(color 0 & 1)
	second 8 bytes are the second half of the plane	(color 2 & 3)
*/
func (t Tile) ToChr() []byte {
	// These are a max of 8 bytes each
	planeA := []byte{}
	planeB := []byte{}

	// Foreach row
	for rowNum := 0; rowNum < 8; rowNum++ {
		a := byte(0)
		b := byte(0)

		// Get the byte for the given row
		// The 8 here isn't row, it's column
		for _, d := range t.Pix[rowNum*8 : ((rowNum + 1) * 8)] {
			// Normalize index to be between 0 and 3, inclusively
			d = d % 4

			// Get the bit for each plane and shift it onto their bytes
			a = a<<0x1 | byte(d)&0x1
			b = b<<0x1 | byte(d)>>1
		}

		// Add the bytes to their respective planes
		planeA = append(planeA, a)
		planeB = append(planeB, b)
	}

	// return the tiles two planes
	return append(planeA, planeB...)
}

func LoadCHR(filename string) (*PatternTable, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open CHR file for reading: %v", err)
	}

	// The data length needs to be a power of 16.
	if len(raw)%16 != 0 {
		return nil, fmt.Errorf("Invalid size of CHR data: %d", len(raw))
	}

	pt := NewPatternTable()

	// Loop through each tile
	for i := 0; i < len(raw); i += 16 {
		tile := NewTile(i / 16)
		plane1 := raw[i+0 : i+8]
		plane2 := raw[i+8 : i+16]

		for row := 0; row < 8; row++ {
			p1 := plane1[row]
			p2 := plane2[row]

			// %x1xx xxxx
			// %x0xx xxxx
			// %xxxx xx01
			for col := 0; col < 8; col++ {
				a := ((p1 >> uint(7-col)) & 1)
				b := ((p2 >> uint(7-col)) & 1)
				px := (a<<1 | b)
				tile.SetPaletteIndex(row, col, px)
			}
		}

		pt.AddTile(tile)
	}

	return pt, nil
}

// These are some the functions for implementing the image.Image
// interface for CHR files.

// DecodeConfig reads the CHR binary and returns four colors and the dimensions of the image.
//func DecodeConfig(r io.Reader) (image.Config, error) {
//}
//
//func Decode(r io.Reader) (image.Image, error) {
//}
//
//func init() {
//	image.RegisterFormat("chr", "", Decode, DecodeConfig)
//}
