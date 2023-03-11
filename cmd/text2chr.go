package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/zorchenhimer/go-nes/image"
)

type options struct {
	// TODO: input file?
	Input string `arg:"--input,required" help:"Input data"`

	OutputChr string `arg:"--chr,required" help:"CHR output file"`
	Metadata  string `arg:"--metadata,required" help:"File to write metadata info to"`
	FontImage string `arg:"--font,required" help:"Font CHR/Bitmap to use"`

	BackgroundColor int `arg:"--background-color" default:"0" help:"Color to use as the background in each tile."`
}

func main() {
	opts := &options{}
	arg.MustParse(opts)

	err := run(opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(opts *options) error {
	var font *image.PatternTable
	var err error

	switch filepath.Ext(opts.FontImage) {
	case ".bmp":
		font, err = image.LoadBitmap(opts.FontImage)
	case ".chr":
		font, err = image.LoadCHR(opts.FontImage)
	default:
		err = fmt.Errorf("Unsupported image format")
	}

	if err != nil {
		return err
	}

	chr := NewChrString(3)
	if opts.BackgroundColor != 0 {
		chr.BackgroundIndex = uint8(opts.BackgroundColor % 4)
		for _, t := range font.Patterns {
			t.SetBackgroundIndex(chr.BackgroundIndex)
		}
	}

	for _, r := range opts.Input {
		if r == ' ' {
			chr.WriteSpace()
			continue
		}
		i := int(r)
		if i >= len(font.Patterns) {
			return fmt.Errorf("%q [0x%X] does not exist in font", r, r)
		}
		chr.WriteTile(font.Patterns[i])
	}

	pt := chr.PatternTable()

	err = os.WriteFile(opts.OutputChr, pt.Chr(false), 0644)
	if err != nil {
		return fmt.Errorf("Unable to write CHR data: %w", err)
	}

	out, err := os.Create(opts.Metadata)
	if err != nil {
		return fmt.Errorf("Unable to create metadata file: %w", err)
	}
	defer out.Close()

	str := []string{}
	for _, b := range pt.ReducedIds {
		str = append(str, fmt.Sprintf("$%02X", b))
	}

	fmt.Fprintf(out, "  .byte %d\n", len(str))
	fmt.Fprintln(out, "  .byte", strings.Join(str, ", "))

	return nil
}

type ChrString struct {
	// There is only one row of tiles.
	Tiles []*image.Tile

	// Dimensions in tiles
	Height int
	Width  int

	SpaceSize int
	LastCol   int

	BackgroundIndex uint8
}

func NewChrString(spaceSize int) *ChrString {
	return &ChrString{
		Tiles:     []*image.Tile{},
		SpaceSize: spaceSize,
	}
}

func (cs *ChrString) WriteTile(tile *image.Tile) {
	// Go column by column
	for x := 0; x < tile.CharacterWidth(); x++ {
		for y := 0; y < 8; y++ {
			c := tile.At(x, y)
			cs.Set(x+cs.LastCol, y, c)
		}
	}

	cs.LastCol += tile.CharacterWidth() + 1
}

func (cs *ChrString) WriteSpace() {
	cs.LastCol += cs.SpaceSize
}

func (cs *ChrString) At(x, y int) color.Color {
	return cs.TileAt(x, y).At(x%8, y%8)
}

func (cs *ChrString) Set(x, y int, c color.Color) {
	// Add tile colums when needed
	for (x / 8) >= len(cs.Tiles) {
		nt := image.NewTile(len(cs.Tiles))
		if cs.BackgroundIndex != 0 {
			nt.FillBackground(cs.BackgroundIndex)
		}
		cs.Tiles = append(cs.Tiles, nt)
	}

	// Swapping the x and y around breaks things.  IDK why.
	//cs.TileAt(x, y).Set(x%8, y%8, c)
	cs.Tiles[x/8].Set(y%8, x%8, c)
}

func (cs *ChrString) TileAt(x, y int) *image.Tile {
	if y > 8 {
		panic("TileAt(): Y cannot be greater than 8")
	}

	tile := x / 8
	if tile >= len(cs.Tiles) {
		panic(fmt.Sprintf("TileAt(%d, %d): Something went wrong. Tile ID %d >= len %d", x, y, tile, len(cs.Tiles)))
	}

	return cs.Tiles[tile]
}

func (cs *ChrString) PatternTable() *image.PatternTable {
	pt := &image.PatternTable{
		Patterns:   cs.Tiles,
		TableWidth: len(cs.Tiles),
	}
	pt.RemoveDuplicates(false)
	return pt
}
