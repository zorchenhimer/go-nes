package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	//"image/png"
	//"bytes"

	"github.com/alexflint/go-arg"

	"github.com/zorchenhimer/go-nes/image"
)

type options struct {
	// TODO: file or command line?
	Input string `arg:"--input,required" help:"Input data"`

	OutputChr string `arg:"--chr,required" help:"CHR output file"`
	Metadata  string `arg:"--metadata,required" help:"File to write metadata info to"`
	FontImage string `arg:"--font,required" help:"Font CHR/Bitmap to use"`
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

	//buf := bytes.NewBuffer([]byte{})
	//err = png.Encode(buf, font)
	//if err != nil {
	//	return fmt.Errorf("error encoding loaded-font.png: %w", err)
	//}
	//err = os.WriteFile("loaded-font.png", buf.Bytes(), 0777)
	//if err != nil {
	//	return fmt.Errorf("error saving loaded-font.png: %w", err)
	//}

	// IDs into the paterren tables
	chr := NewChrString(3)
	//runes := *image.Tile[]
	for _, r := range opts.Input {
		if r == ' ' {
			chr.WriteSpace()
			continue
		}
		i := int(r)
		if i >= len(font.Patterns) {
			return fmt.Errorf("%q [0x%X] does not exist in font", r, r)
		}
		//runes = append(runes, font.Patterns[i]
		//t := font.Patterns[i]
		//c := t.Chr(false)
		//for j := 0; j < 8; j++ {
		//	fmt.Println(strings.ReplaceAll(fmt.Sprintf("%08b", c[j]), "0", "_"))
		//}
		chr.WriteTile(font.Patterns[i])
		//fmt.Println("")
	}

	//buf.Reset()
	//err = png.Encode(buf, chr.Tiles[1])
	//if err != nil {
	//	return fmt.Errorf("error encoding first tile: %w", err)
	//}

	//err = os.WriteFile("first-tile.png", buf.Bytes(), 0777)
	//if err != nil {
	//	return fmt.Errorf("error saving first-tile.png: %w", err)
	//}

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

	fmt.Fprintf(out,  "  .byte %d\n", len(str))
	fmt.Fprintln(out, "  .byte", strings.Join(str, ", "))

	//for i := 0; i < len(pt.Patterns); i++ {
	//for i := 0; i < 2; i++ {
	//	t := pt.Patterns[i]
	//	c := t.Chr(false)
	//	for j := 0; j < 8; j++ {
	//		fmt.Println(strings.ReplaceAll(fmt.Sprintf("%08b", c[j]), "0", "_"))
	//	}
	//	fmt.Println("")
	//}
	return nil
}

type ChrString struct {
	// Tile order is different than standard pattern tables.  Goes vertically down column before across row.
	Tiles []*image.Tile

	// Dimensions in tiles
	Height int
	Width  int

	SpaceSize int
	LastCol   int
}

func NewChrString(spaceSize int) *ChrString {
	return &ChrString{
		Tiles:     []*image.Tile{},
		SpaceSize: spaceSize,
	}
}

func (cs *ChrString) WriteTile(tile *image.Tile) {
	// ????? If it's tile.At(x, y) shit comes out mirrored.
	//       but with y,x and tile.CharacterWidth(), the bottom
	//       is chopped off.  wtf?
	//for x := 0; x < tile.CharacterWidth(); x++ {
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			c := tile.At(x, y)
			cs.Set(x+cs.LastCol, y, c)
		}
	}

	cs.LastCol += tile.CharacterWidth() + 1
	//fmt.Printf("WriteTile() char width: %d\n", tile.CharacterWidth())
}

func (cs *ChrString) WriteSpace() {
	cs.LastCol += cs.SpaceSize
}

func (cs *ChrString) At(x, y int) color.Color {
	return cs.TileAt(x, y).At(x%8, y%8)
}

func (cs *ChrString) Set(x, y int, c color.Color) {
	//fmt.Printf("     Set(%02d, %02d, %v)\n", x, y, c)
	// Add tile colums when needed
	//fmt.Printf("ChrString.Set(%d, %d, %v) len(tiles): %d\n", x ,y ,c, len(cs.Tiles))
	//fmt.Print(".")
	for (x / 8) >= len(cs.Tiles) {
		//fmt.Printf("Set(%d, %d) %d (%d) >= %d\n", x, y, (x/8), (x%8), len(cs.Tiles))
		cs.Tiles = append(cs.Tiles, image.NewTile(len(cs.Tiles)))
	}
	//cs.TileAt(x, y).Set(x%8, y%8, c)
	cs.Tiles[x/8].Set(y%8, x%8, c)
}

func (cs *ChrString) TileAt(x, y int) *image.Tile {
	// Out of bounds
	//if y >= 8 {
	//	panic("Single-height fonts only, plz")
	//}
	//if x >= cs.Width*8 || y >= cs.Height*8 || x < 0 || y < 0 {
	//	return nil
	//}

	tile := x / 8
	if tile >= len(cs.Tiles) {
		panic(fmt.Sprintf("TileAt(%d, %d): Something went wrong. Tile ID %d >= len %d", x, y, tile, len(cs.Tiles)))
	}

	//fmt.Printf("ChrString.TileAt(%d, %d) [%d] %v\n", x, y, tile, cs.Tiles[tile])
	return cs.Tiles[tile]
}

func (cs *ChrString) PatternTable() *image.PatternTable {
	pt := &image.PatternTable{
		Patterns: cs.Tiles,
	}
	pt.RemoveDuplicates(false)
	return pt
}
