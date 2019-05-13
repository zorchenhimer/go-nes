package image

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Arrangement uint
const (
	ARR_SINGLE Arrangement = iota
	ARR_DBLHIGH	// 8x16 sprite mode
	// TODO: more?
)

// A PatternTable is the data tha will be written as the CHR.  It can only be in
// 1k, 2k, 4k, or 8k sizes; or 64, 128, 256, or 512 tiles; or 128x32, 128x64,
// 128x128, or 128x256 pixels; or 4, 8, 16, or 32 rows of tiles.
type PatternTable struct {
	Patterns	[]*Tile
	Layout		Arrangement
	ReducedIds	[]int
}

//type TableSize int
//const (
//	TS_64	TableSize = TableSize(64)
//	TS_128	TableSize = TableSize(128)
//	TS_256	TableSize = TableSize(256)
//	TS_512	TableSize = TableSize(512)
//)

//func (ts TableSize) String() string {
//	switch ts {
//	case TS_64:
//		return "1k (64 tiles)"
//	case TS_128:
//		return "2k (128 tiles)"
//	case TS_256:
//		return "4k (256 tiles)"
//	case TS_512:
//		return "8k (512 tiles)"
//	}
//}

func NewPatternTable() *PatternTable {
	return &PatternTable{
		Patterns:	[]*Tile{},
	}
}

func (pt *PatternTable) Debug() string {
	//return fmt.Sprintf("Size: %s", pt.Size)
	l := float64(len(pt.Patterns))
	row := int(math.Ceil(l / 16.0))
	return fmt.Sprintf("%d tiles; %d rows", int(l), row)
}

func (pt *PatternTable) AddTile(tile *Tile) {
	pt.Patterns = append(pt.Patterns, tile)
}

// Returns before/after count
func (pt *PatternTable) RemoveDuplicates() (int, int) {
	tiles := []*Tile{}
	pt.ReducedIds = []int{}

OUTER:
	for idx, tile := range pt.Patterns {
		for id, t := range tiles {
			if t.IsIdentical(tile) {
				pt.ReducedIds = append(pt.ReducedIds, id)
				continue OUTER
			}
		}
		pt.ReducedIds = append(pt.ReducedIds, idx)
		tiles = append(tiles, tile)
	}

	pt.Patterns = tiles
	return len(pt.ReducedIds), len(pt.Patterns)
}

func (pt *PatternTable) WriteFile(filename string) error {
	chr := []byte{}

	for _, t := range pt.Patterns {
		chr = append(chr, t.Chr()...)
	}

	// Only write tile IDs if duplicates have been removed
	if pt.ReducedIds != nil {
		name := filename[:len(filename) - len(filepath.Ext(filename))] + ".ids.asm"
		file, err := os.Create(name)
		if err != nil {
			return err
		}
		defer file.Close()

		line := []string{}
		for i := 0; i < len(pt.ReducedIds); i++ {
			line = append(line, fmt.Sprintf("$%02X", pt.ReducedIds[i]))
			if i % 32 == 0 {
				fmt.Fprintf(file, ".byte %s\n", strings.Join(line, ", "))
				line = []string{}
			}
		}
		fmt.Fprintf(file, ".byte %s\n", strings.Join(line, ", "))
	}

	return ioutil.WriteFile(filename, chr, 0777)
}

//func (pt *PatternTable) Add(img *image.Paletted, arr Arrangement) error {
//	if img.Bounds().Max.X % 8 != 0 || img.Bounds().Max.Y % 8 != 0 {
//		return fmt.Errorf("Dimensions must be multiples of 8")
//	}
//
//	mt, err := NewMetaTileFromPaletted(img)
//	if err != nil {
//		return err
//	}
//	pt.Patterns = append(pt.Patterns, mt)
//	return nil
//}

// Implement the image.Image interface
func (pt *PatternTable) ColorModel() color.Model {
	// TODO
	return nil
}

func (pt *PatternTable) Bounds() image.Rectangle {
	// TODO
	width := 128
	if len(pt.Patterns) < 16 {
		width = len(pt.Patterns) * 8
	}
	return image.Rectangle{}
}

func (pt *PatternTable) At(x, y int) color.Color {
	// TODO
	return nil
}

// Implement image.draw.Drawer and image.draw.Image
func (pt *PatternTable) Draw(dst image.Image, r image.Rectangle, src image.Image, sp image.Point) {
	// TODO
}

func (pt *PatternTable) Set(x, y int, c color.Color) {
	// TODO
}
