package image

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
)

func LoadBitmap(filename string) (*PatternTable, error) {

	// Read input file
	rawBmp, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open input bitmap file: %s", err)
	}

	// Parse some headers
	fileHeader, err := ParseFileHeader(rawBmp)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse bitmap file header: %s", err)
	}

	imageHeader, err := ParseImageHeader(rawBmp)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse bitmap image header: %s", err)
	}

	// Validate image dimensions
	if imageHeader.Width%8 != 0 {
		return nil, fmt.Errorf("Image width must be a multiple of 8")
	}

	if imageHeader.Height%8 != 0 {
		return nil, fmt.Errorf("Image height must be a multiple of 8")
	}

	if imageHeader.BitDepth != 8 {
		return nil, fmt.Errorf("Image has incorrect bit depth of %d", imageHeader.BitDepth)
	}

	rect := image.Rect(0, 0, imageHeader.Width, imageHeader.Height)
	data := rawBmp[fileHeader.Offset:len(rawBmp)]

	// Invert row order. They're stored top to bottom in BMP.
	uprightRows := []byte{}
	for row := (len(data) / rect.Max.X) - 1; row > -1; row-- {
		// Get the row
		rawRow := data[row*rect.Max.X : row*rect.Max.X+rect.Max.X]

		// normalize each pixel's palette index
		for _, b := range rawRow {
			uprightRows = append(uprightRows, byte(int(b)%4))
		}
	}

	// Cut out the 8x8 tiles
	tileID := 0
	table := NewPatternTable()

	tilesPerRow := rect.Max.X / 8

	for tileID < (len(uprightRows) / 64) {
		// The first pixel offset in the current tile

		// tile row * tile row length in pixels + offset in tile
		startOffset := (tileID/tilesPerRow)*(64*tilesPerRow) + (tileID%tilesPerRow)*8

		var nt *Tile
		nt = NewTile(tileID)
		for y := 0; y < 8; y++ {
			tileY := y

			// Wrap rows at 8 pixels
			if tileY >= 8 {
				tileY -= 8
			}

			// Get the pixels for the row.
			for x := 0; x < 8; x++ {
				nt.Pix[x+(8*tileY)] = uprightRows[startOffset+x+rect.Max.X*y]
			}
		}

		//tiles = append(tiles, nt)
		table.AddTile(nt)
		tileID++
	}

	return table, nil
}

type FileHeader struct {
	Size   int // size of file in bytes
	Offset int // offset to start of pixel data
}

func (f FileHeader) String() string {
	return fmt.Sprintf("Size: %d Offset: %d", f.Size, f.Offset)
}

// Size, offset, error
func ParseFileHeader(input []byte) (*FileHeader, error) {
	if len(input) < 4 {
		return nil, fmt.Errorf("Data too short for header")
	}
	header := input[0:14]

	size := binary.LittleEndian.Uint32(header[2:6])
	offset := binary.LittleEndian.Uint32(header[10:14])
	return &FileHeader{Size: int(size), Offset: int(offset)}, nil
}

type ImageHeader struct {
	headerSize  int
	Width       int
	Height      int
	BitDepth    int
	Compression int
	Size        int // image size

	// "Pixels per meter"
	ppmX int
	ppmY int

	ColorMapEntries   int
	SignificantColors int
}

func (i ImageHeader) String() string {
	return fmt.Sprintf("(%d, %d) %d bpp @ %d bytes", i.Width, i.Height, i.BitDepth, i.Size)
}

func ParseImageHeader(input []byte) (*ImageHeader, error) {
	if len(input) < (14 + 12) {
		return nil, fmt.Errorf("Data too short for image header")
	}

	header := &ImageHeader{}
	header.headerSize = int(binary.LittleEndian.Uint32(input[14:18]))

	//headerRaw := input[14 : 14+header.Size]

	header.Width = int(binary.LittleEndian.Uint32(input[18:22]))
	header.Height = int(binary.LittleEndian.Uint32(input[22:26]))
	header.BitDepth = int(binary.LittleEndian.Uint16(input[28:30]))

	header.Size = int(binary.LittleEndian.Uint32(input[38:42]))

	return header, nil
}
