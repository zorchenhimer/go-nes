package rom

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// For visualizing the code/data breakdown in a rom

// type PRGUsage byte
// type CHRUsage byte
const (
	// PRG ROM
	UP_CODE    byte = 0x01
	UP_DATA    byte = 0x02
	UP_JMPCODE byte = 0x10 // Indirect code; via JMP instruction
	UP_LDADATA byte = 0x20 // Indirect data; via LDA instruction
	UP_AUDIO   byte = 0x40
	UP_UNKNOWN byte = 0x00

	// CHR ROM
	UC_DRAWN  byte = 0x01 // rendered by PPU
	UC_READ   byte = 0x02 // Read via port $2007
	UC_UNKOWN byte = 0x00
)

func ImgFromBin(filename string, data []byte) error {
	width := 256
	height := int(math.Ceil(float64(len(data)) / float64(width)))

	fmt.Printf("width: %d\nheight: %d\n", width, height)

	img := image.NewGray(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{width, height}})

	for i := 0; i < len(data); i++ {
		x := int(math.Mod(float64(i), float64(width)))
		y := int(math.Floor(float64(i) / float64(width)))

		img.Set(x, y, color.Gray{Y: uint8(data[i]) + 255})
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create %q: %s", filename, err)
	}

	if err = png.Encode(file, img); err != nil {
		return fmt.Errorf("Unable to write %q: %s", filename, err)
	}

	return nil
}
