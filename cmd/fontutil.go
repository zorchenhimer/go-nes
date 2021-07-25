package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"

	nesimg "github.com/zorchenhimer/go-nes/image"
)

type options struct {
	// Input BMP file
	Input string `arg:"-i,--input,required" help:"Input bitmap file containing the font."`

	// Output files
	Output string `arg:"-o,--output,required" help:"Output file for CHR ASM data."`
	Remap  string `arg:"-r,--remap" help:"Output file for ca65 character remappings."`
	Widths string `arg:"-w,--widths" help:"Output file for character width values."`

	//Help bool `arg:"-h,--help" help:"Print help and exit."`

	//Verbose bool `arg:"-v,--verbose" help:"Add verbosity"`
	InputOffset int `arg:"--input-offset" default:"0" help:"Tile offset to start reading at from the input file."`
	InputLength int `arg:"--input-length" default:"0" help:"Number of tiles to look at.  Zero turns this off."`

	SpaceWidth int `arg:"-s,--space-width" default:"5" help:"Width of the SPACE character."`
}

// Take input image file (single file) and output the tile-reduced
// font CHR (as assembly), width values, and character remappings
// to three files.
func main() {
	var opts *options = &options{}
	parser := arg.MustParse(opts)

	if opts.Output == "" {
		fmt.Println("Missing output file\n")
		parser.WriteUsage(os.Stdout)
		os.Exit(1)
	}

	if opts.Input == "" {
		fmt.Println("Missing input file\n")
		parser.WriteUsage(os.Stdout)
		os.Exit(1)
	}

	if filepath.Ext(opts.Input) != ".bmp" {
		fmt.Println("Only bitmap images are supported as input\n")
		parser.WriteUsage(os.Stdout)
		os.Exit(1)
	}

	if opts.InputOffset < 0 {
		fmt.Println("Input offset cannot be negative")
		os.Exit(1)
	}

	pt, err := nesimg.LoadBitmap(opts.Input)
	if err != nil {
		fmt.Printf("Unable to load bitmap: %v\n", err)
		os.Exit(1)
	}

	if opts.InputOffset >= len(pt.Patterns) {
		fmt.Println("Input offset is larger that the tile count")
		os.Exit(1)
	}

	if opts.InputOffset > 0 {
		if opts.InputLength == 0 {
			opts.InputLength = len(pt.Patterns)
		}
		pt.Patterns = pt.Patterns[opts.InputOffset:opts.InputLength]
	}

	// Remove dupes manually after the input offset so the first tile
	// in the output has the correct OrigId value.
	pt.RemoveDuplicates(false)

	if opts.Remap != "" {
		if err = writeRemap(opts.Remap, pt); err != nil {
			fmt.Printf("Unable to write remap file: %v\n", err)
			os.Exit(1)
		}
	}

	if opts.Widths != "" {
		if err = writeWidths(opts.Widths, pt, opts.SpaceWidth); err != nil {
			fmt.Printf("Unable to write widths file: %v\n", err)
			os.Exit(1)
		}
	}

	data := []byte(pt.Asm(true))
	if err = ioutil.WriteFile(opts.Output, data, 0644); err != nil {
		fmt.Printf("Unable to write ASM file: %v\n", err)
		os.Exit(1)
	}
}

func writeRemap(filename string, pt *nesimg.PatternTable) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create remap file: %v", err)
	}

	for idx, tile := range pt.Patterns {
		// Use hex notation for non-printing characters
		if tile.OrigId < 0x20 || tile.OrigId > 0x7E {
			fmt.Fprintf(file, ".charmap $%02X, $%02X ; $%X\n", tile.OrigId, idx, tile.OrigId)
		} else {
			fmt.Fprintf(file, ".charmap $%02X, $%02X ; '%c'\n", tile.OrigId, idx, tile.OrigId)
		}
	}
	return file.Close()
}

func writeWidths(filename string, pt *nesimg.PatternTable, spaceWidth int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create widths file: %v", err)
	}

	for _, tile := range pt.Patterns {
		w := tile.CharacterWidth()
		if w == 0 {
			w = spaceWidth
		} else {
			w += 1
		}

		// Use hex notation for non-printing characters
		if tile.OrigId < 0x20 || tile.OrigId > 0x7E {
			fmt.Fprintf(file, ".byte %d ; $%X\n", w, tile.OrigId)
		} else {
			fmt.Fprintf(file, ".byte %d ; '%c'\n", w, tile.OrigId)
		}
	}
	return file.Close()
}
