package image

import (
//"image"
//"image/color"
//"io"
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
