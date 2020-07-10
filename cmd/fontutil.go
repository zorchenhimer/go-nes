package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	nesimg "github.com/zorchenhimer/go-nes/image"
)

var (
	opt_Output string // output CHR asm file
	opt_Remap  string // output .charmap remapping file
	opt_Width  string // output width listing file
	opt_Input  string // input bmp file
)

// Take input image file (single file) and output the tile-reduced
// font CHR (as assembly), width values, and character remappings
// to three files.
func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "A simple command line utility to take a single bitmap of a font and convert it into a single-plane, tile-reduced, CHR in assembly with a chracter remap file and a list of character widths.\n\nRequired options:\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&opt_Output, "o", "", "Output CHR asm file")
	flag.StringVar(&opt_Remap, "r", "", "Output .charmap remapping file file")
	flag.StringVar(&opt_Width, "w", "", "Output widths file")
	flag.StringVar(&opt_Input, "i", "", "Input BMP file")
	flag.Parse()

	opt_Output = strings.TrimSpace(opt_Output)
	opt_Remap = strings.TrimSpace(opt_Remap)
	opt_Input = strings.TrimSpace(opt_Input)

	if opt_Output == "" {
		fmt.Println("Missing output file\n")
		flag.Usage()
		os.Exit(1)
	}

	if opt_Remap == "" {
		fmt.Println("Missing remap file\n")
		flag.Usage()
		os.Exit(1)
	}

	if opt_Width == "" {
		fmt.Println("Missing width file\n")
		flag.Usage()
		os.Exit(1)
	}

	if opt_Input == "" {
		fmt.Println("Missing input file\n")
		flag.Usage()
		os.Exit(1)
	}

	if filepath.Ext(opt_Input) != ".bmp" {
		fmt.Println("Only bitmap images are supported as input\n")
		flag.Usage()
		os.Exit(1)
	}

	pt, err := nesimg.LoadBitmap(opt_Input)
	if err != nil {
		fmt.Printf("Unable to load bitmap: %v\n", err)
		os.Exit(1)
	}

	pt.RemoveDuplicates(true)
	if err = writeRemap(opt_Remap, pt); err != nil {
		fmt.Printf("Unable to write remap file: %v\n", err)
		os.Exit(1)
	}

	if err = writeWidths(opt_Width, pt); err != nil {
		fmt.Printf("Unable to write widths file: %v\n", err)
		os.Exit(1)
	}

	data := []byte(pt.Asm(true))
	if err = ioutil.WriteFile(opt_Output, data, 0777); err != nil {
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
		fmt.Fprintf(file, ".charmap $%02X, $%02X\n", tile.OrigId, idx)
	}
	return file.Close()
}

func writeWidths(filename string, pt *nesimg.PatternTable) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create widths file: %v", err)
	}

	for _, tile := range pt.Patterns {
		fmt.Fprintf(file, ".byte %d\n", tile.CharacterWidth()+1)
	}
	return file.Close()
}
