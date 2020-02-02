package image

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// This is that ugly palette from YYCHR.  Currently only used when
// saving a pattern table to a PNG image.  Override this with the
// --palette command line option.
//
// Note: this variable isn't actually used as the default colors are
// defined for the --palette option.
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

// ParseHexPalette takes a palette from a string in hex in the
// from of "#003973,#ADB531,#845E21,#C6E79C" and returns a color
// palette.
//
// The input must have exactly four colors defined or an error will
// be returned.
func ParseHexPalette(input string) (color.Palette, error) {
	pal := color.Palette{}

	for _, str := range strings.Split(input, ",") {
		c, err := ParseHexColor(str)
		if err != nil {
			return nil, err
		}
		pal = append(pal, c)
	}

	if len(pal) != 4 {
		return nil, fmt.Errorf("Palette must have exactly four colors. Found %d", len(pal))
	}

	return pal, nil
}

// ParseHexColor takes a single color represented in hexidecimal format
// and returns a color.RGBA object on success, and an error otherwise.
//
// Shorthand notation can be used (eg, #000 is expanded to #000000).
func ParseHexColor(input string) (color.RGBA, error) {
	if input[0] == '#' {
		input = input[1:len(input)]
	}

	length := len(input)
	if length != 6 && length != 3 {
		return color.RGBA{}, fmt.Errorf("Invalid length for hex color %q", input)
	}

	nums := []uint8{}
	if length == 6 {
		for i := 0; i < length; i += 2 {
			i64, err := strconv.ParseInt(input[i:i+2], 16, 9)
			if err != nil {
				return color.RGBA{}, err
			}

			nums = append(nums, uint8(i64))
		}

	} else if length == 3 {
		for i := 0; i < length; i++ {
			i64, err := strconv.ParseInt(fmt.Sprintf("%c%c", input[i], input[i]), 16, 9)
			if err != nil {
				return color.RGBA{}, err
			}

			nums = append(nums, uint8(i64))
		}
	}

	if len(nums) != 3 {
		return color.RGBA{}, fmt.Errorf("Failed parsing hex numbers")
	}

	return color.RGBA{R: nums[0], G: nums[1], B: nums[2], A: 0xFF}, nil

}
