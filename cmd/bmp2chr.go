package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/image"
)

var supportedInputFormats []string = []string{".bmp", ".json"}

func main() {
	var doubleHigh bool
	var outputFilename string
	var debug bool
	var asmOutput bool
	var asmOutputHalf bool

	flag.StringVar(&outputFilename, "o", "", "Output filename")
	flag.BoolVar(&doubleHigh, "16", false, "8x16 tiles")
	flag.BoolVar(&debug, "debug", false, "Debug printing")
	flag.BoolVar(&asmOutput, "asm", false, "Output data in ASM")
	flag.BoolVar(&asmOutputHalf, "asmhalf", false, "Output data in ASM (first bit plane only)")
	flag.Parse()

	fileList := []string{}

	if asmOutput && asmOutputHalf {
		fmt.Printf("--asm and --asmhalf cannot be used together")
		os.Exit(1)
	}

	if len(flag.Args()) > 0 {
		for _, target := range flag.Args() {
			found, err := filepath.Glob(target)
			if err == nil && len(found) > 0 {
				fileList = append(fileList, found...)
			} else {
				fmt.Printf("%q not found\n", target)
				os.Exit(1)
			}
		}
	}

	if len(fileList) == 0 {
		fmt.Println("Missing input file(s)")
		os.Exit(1)
	}

	// Require an output filename if there's more than one input.
	if len(outputFilename) == 0 {
		if len(fileList) == 1 {
			outputFilename = fileList[0]
			ext := filepath.Ext(fileList[0])
			outputFilename = outputFilename[0:len(outputFilename)-len(ext)] + ".chr"
		} else {
			fmt.Println("Missing output filename")
			os.Exit(1)
		}
	}

	for _, file := range fileList {
		ext := filepath.Ext(file)
		found := false
		for _, supp := range supportedInputFormats {
			if ext == supp {
				found = true
			}
		}
		if !found {
			fmt.Printf("Unsupported input format for file %q\n", file)
			os.Exit(1)
		}
	}

	chrFile, err := os.Create(outputFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer chrFile.Close()

	inputBitmaps := map[string]*image.Bitmap{}

	for _, inputfile := range fileList {
		var err error
		var bitmap *image.Bitmap

		switch strings.ToLower(filepath.Ext(inputfile)) {
		case ".bmp":
			bitmap, err = image.OpenBitmap(inputfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			inputBitmaps[inputfile] = bitmap

			if debug {
				err := ioutil.WriteFile("upright.dat", bitmap.RawData, 0777)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

		default:
			panic("Unsupported 'supported' format: " + filepath.Ext(inputfile))
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		inputBitmaps[inputfile] = bitmap

	}

	patternTable := []image.Tile{}
	blankTile := image.NewTile(0)

	total_width := 0
	total_height := 0

	for _, infile := range inputBitmaps {
		rect := infile.Rect()
		total_width += rect.Dx()
		total_height += rect.Dy()
	}

	for i := 0; i < (total_width*total_height)/64; i++ {
		patternTable = append(patternTable, *blankTile)
	}

	index := 0 // current tile index
	for _, bitmap := range inputBitmaps {
		// If it's 8x16 mode, transform tiles.  Tiles on odd rows will be put
		// after the tile directly above them. The first four tiles would be
		// $00, $10, $01, $11.
		// If the number of rows is not even, ignore 8x16 mode.
		if doubleHigh && bitmap.Rect().Max.Y%16 == 0 {
			newtiles := []image.Tile{}
			for i := 0; i < len(bitmap.Tiles)/2; i++ {
				if i%bitmap.TilesPerRow == 0 && i > 0 {
					i += bitmap.TilesPerRow
				}

				newtiles = append(newtiles, bitmap.Tiles[i])
				newtiles = append(newtiles, bitmap.Tiles[i+bitmap.TilesPerRow])
			}
			bitmap.Tiles = newtiles
		}

		for _, tile := range bitmap.Tiles {
			patternTable[index] = tile

			index++
		}
	}

	// Remove duplicates
	// But only in 8x8 mode
	if !doubleHigh {
		tileIds, err := os.Create(outputFilename + ".ids.asm")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer tileIds.Close()

		noDupes := []image.Tile{}
		var outid int
	OUTER:
		for _, tile := range patternTable {
			for id, nodup := range noDupes {
				if tile.IsIdentical(nodup) {
					outid = id
					for outid > 255 {
						outid -= 255
					}

					fmt.Fprintf(tileIds, ".byte $%02X\n", outid)
					continue OUTER
				}
			}

			outid = len(noDupes)
			for outid > 255 {
				outid -= 255
			}

			fmt.Fprintf(tileIds, ".byte $%02X\n", outid)
			noDupes = append(noDupes, tile)
		}

		fmt.Printf("Duplicate reduction %d -> %d\n", len(patternTable), len(noDupes))
		patternTable = noDupes
	}

	if len(patternTable) >= 16*16 {
		fmt.Println("More than 256 tiles!")
	}

	if asmOutputHalf || asmOutput {
		asmOut, err := os.Create(outputFilename + ".bin.asm")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer asmOut.Close()

		for _, tile := range patternTable {
			ascii := strings.Split(tile.ASCII(), "\n")
			ascii_comment := strings.Join(ascii, "\n; ")
			fmt.Fprintf(asmOut, "%s\n", ascii_comment)
			fmt.Fprintf(asmOut, "%s\n", tile.Asm())
		}
	}

	for _, tile := range patternTable {
		if debug {
			fmt.Println(tile.ASCII())
		}

		tchr := tile.ToChr()
		_, err = chrFile.Write(tchr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
