package image

import (
	"fmt"
	"image/color"
)

// This is that ugly palette from YYCHR
var DefaultPalette color.Palette = color.Palette{
	color.RGBA{R: 0x00, G: 0x39, B: 0x73, A: 0xFF},
	color.RGBA{R: 0xAD, G: 0xB5, B: 0x31, A: 0xFF},
	color.RGBA{R: 0x84, G: 0x5E, B: 0x21, A: 0xFF},
	color.RGBA{R: 0xC6, G: 0xE7, B: 0x9C, A: 0xFF},
}

var NESModel color.Model = color.ModelFunc(nesModel)

func nesModel(c color.Color) color.Color {
	return DefaultPalette.Convert(c)
}

func getColorIndex(pal color.Palette, c color.Color) (int, error) {
	r, g, b, a := c.RGBA()
	for idx, palcolor := range pal {
		pr, pg, pb, pa := palcolor.RGBA()
		if r == pr && g == pg && b == pb && a == pa {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("Color not in palette")
}
