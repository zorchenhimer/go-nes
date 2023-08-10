package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	nesimg "github.com/zorchenhimer/go-nes/image"
)

type options struct {
	Input      string `arg:"positional,required" help:"Input bitmap file"`
	OutputChr  string `arg:"positional,required" help:"Output CHR file"`
	OutputData string `arg:"positional,required" help:"Output tile metadata assembly code"`
	TileSize   string `arg:"-s,--tile-size" default:"2" help:"Meta tile size in number of CHR tiles. Format is either a single number for a square or WxH for a rectangle (eg 1x2 or 2).  The source file must be the proper dimensions for the given tile size."`
	Count      int    `arg:"-c,--count" default:"0" help:"Number of meta tiles to process.  A value of zero processes all available given the image and metatile dimensions."`
	Offset     int    `arg:"-o,--offset" default"0" help:"Offset to start the tile IDs"`
	PadTiles   int    `arg:"-p,--pad" default:"0" help:"Pad the output to contain at least this many tiles"`
	sizeWidth  int
	sizeHeight int
}

type MetaTile struct {
	Tiles   []int // IDs in the pattern table
	Palette int
	Width   int
	Height  int
}

func (mt MetaTile) String() string {
	s := []string{}
	for _, i := range mt.Tiles {
		s = append(s, strconv.Itoa(i))
	}
	return fmt.Sprintf("[%s]", strings.Join(s, " "))
}

func (mt MetaTile) Asm(offset int) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(":   .byte %d, %d\n", mt.Width, mt.Height))
	sb.WriteString(fmt.Sprintf("    .byte %d, %d\n", mt.Palette, mt.Width*mt.Height))
	s := []string{}
	for _, i := range mt.Tiles {
		s = append(s, strconv.Itoa(i+offset))
	}
	sb.WriteString(fmt.Sprintf("    .byte %s\n", strings.Join(s, ", ")))
	return sb.String()
}

var re_tileformat = regexp.MustCompile(`^(\d+)[xX](\d+)$`)

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
	var err error
	if strings.Contains(strings.ToLower(opts.TileSize), "x") {
		matches := re_tileformat.FindStringSubmatch(opts.TileSize)
		if len(matches) != 3 {
			return fmt.Errorf("Invalid tile size format: %q", opts.TileSize)
		}

		opts.sizeWidth, err = strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("Invalid width: %q: %w", matches[1], err)
		}
		opts.sizeHeight, err = strconv.Atoi(matches[2])
		if err != nil {
			return fmt.Errorf("Invalid height: %q: %w", matches[2], err)
		}

	} else {
		n, err := strconv.Atoi(opts.TileSize)
		if err != nil {
			return fmt.Errorf("Invalid tile size: %q: %w", opts.TileSize, err)
		}

		if n < 1 {
			return fmt.Errorf("Tile size cannot be less than one")
		}

		opts.sizeWidth = n
		opts.sizeHeight = n
	}

	pt, err := nesimg.LoadBitmap(opts.Input)
	if err != nil {
		return fmt.Errorf("Error loading input: %w", err)
	}
	pt.RemoveDuplicates(false)

	tilesWidth := pt.SourceWidth / 8
	if tilesWidth%opts.sizeWidth != 0 {
		return fmt.Errorf("Source image incorrect width for metatile size")
	}

	tilesHeight := pt.SourceHeight / 8
	if tilesHeight%opts.sizeHeight != 0 {
		return fmt.Errorf("Source image incorrect width for metatile size")
	}

	if opts.PadTiles > 0 {
		pt.PadTileCount(opts.PadTiles)
	}

	chr := pt.Chr(false)
	err = os.WriteFile(opts.OutputChr, chr, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write CHR output: %w", err)
	}

	// Figure out how many meta tiles there are
	mtWidth := tilesWidth / opts.sizeWidth
	mtHeight := tilesHeight / opts.sizeHeight

	if opts.Count == 0 {
		opts.Count = mtWidth * mtHeight
	}

	metaTiles := []MetaTile{}
	count := 0

OUTER:
	for y := 0; y < mtHeight; y++ {
		for x := 0; x < mtWidth; x++ {
			count++
			if count > opts.Count {
				break OUTER
			}

			mt := MetaTile{Width: opts.sizeWidth, Height: opts.sizeHeight}
			pal := -1
			for i := 0; i < opts.sizeHeight; i++ {
				for j := 0; j < opts.sizeWidth; j++ {

					id := (y * opts.sizeHeight * tilesWidth) + // mt row start tile
						(x * opts.sizeWidth) + // mt tile start
						(i * opts.sizeWidth * mtWidth) + // mt inner row
						j // mt inner col

					realid := pt.ReducedIds[id]
					if pal == -1 {
						pal = pt.Patterns[realid].PaletteId
					}

					if pal != pt.Patterns[realid].PaletteId {
						return fmt.Errorf("MetaTile ID %d has more than one palette", count-1)
					}

					mt.Tiles = append(mt.Tiles, realid)
				}
			}
			mt.Palette = pal
			metaTiles = append(metaTiles, mt)
		}
	}

	asmOut, err := os.Create(opts.OutputData)
	if err != nil {
		return fmt.Errorf("Error opening output file: %w", err)
	}
	defer asmOut.Close()

	// Don't write a label.  Have the including source do that instead.
	//_, err = fmt.Fprintln(asmOut, "MetaTiles:")
	//if err != nil {
	//	return fmt.Errorf("Error writing to output file: %w", err)
	//}
	for i := 0; i < len(metaTiles); i++ {
		s := []string{}
		for x := 0; x < i; x++ {
			s = append(s, "+")
		}
		_, err = fmt.Fprintf(asmOut, "    .word :+%s\n", strings.Join(s, ""))
		if err != nil {
			return fmt.Errorf("Error writing to output file: %w", err)
		}
	}

	_, err = fmt.Fprintln(asmOut, "\n; MetaTile Data:\n; Width, Height\n; Palette, Total tiles (W*H)\n; List of tiles\n")
	if err != nil {
		return fmt.Errorf("Error writing to output file: %w", err)
	}

	for _, mt := range metaTiles {
		_, err = fmt.Fprintln(asmOut, mt.Asm(opts.Offset))
		if err != nil {
			return fmt.Errorf("Error writing to output file: %w", err)
		}
	}

	return nil
}
