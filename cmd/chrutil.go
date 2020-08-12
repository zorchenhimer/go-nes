package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/common"
	nesimg "github.com/zorchenhimer/go-nes/image"
)

func main() {
	cp := common.NewCommandParser("Convert bitmap images into CHR images")
	cp.AddOption("output", "o", true, "",
		"File to write output.")
	cp.AddOption("remove-duplicates", "d", false, "false",
		"Remove duplicate tiles.")
	cp.AddOption("debug", "", false, "false",
		"Print debug info to console.")
	cp.AddOption("remove-empty", "", false, "false",
		"Remove empty tiles.")
	cp.AddOption("asm", "a", false, "false",
		"Write output as assembly instead of binary CHR data.")

	// Currently only used with the PNG output format
	cp.AddOption("palette", "p", true, "#003973,#ADB531,#845E21,#C6E79C",
		"Override the default palette with the supplied values.  Expects HTML Hex color codes separated by commas.  The default value being \"#003973,#ADB531,#845E21,#C6E79C\".  Currently only used with PNG output.")

	// Only write the first bit plane of CHR.  Only usable with --asm.
	cp.AddOption("first-plane", "", false, "false",
		"// TODO\nOnly write the first bit plane of CHR data.  Only usable with --asm.")

	cp.AddOption("tile-count", "", true, "0",
		"Number of tiles to read from the source image.")
	cp.AddOption("tile-offset", "", true, "0",
		"Number of tiles to skip from the source image.")

	cp.AddOption("pad-tiles", "", true, "0",
		"Pad the output with blank tiles until it the tile count is equal to or greater than the given value.")

	// Unimplemented
	cp.AddOption("8x16-sprites", "", false, "false",
		"// TODO")
	cp.AddOption("text", "t", true, "",
		"// TODO")
	cp.AddOption("start-id", "i", true, "0",
		"// TODO\nStart at this ID when reading the input file.")

	// Assumes --asm --first-plane --remove-duplaciates
	cp.AddOption("font", "f", false, "false",
		"// TODO\nConvert bitmap font to assembly.  Assumes --asm --first-plane --remove-duplaciates")

	cp.AddOption("write-ids", "", true, "",
		"Write tile IDs to a file to reconstruct an image.  Only available with --remove-duplicates or --remove-empty.")
	cp.AddOption("nt-ids", "", true, "",
		"Write tile IDs to a file to reconstruct an image.  Similar to --wirte-ids but different.  Only available with --remove-duplicates or --remove-empty.")

	err := cp.Parse()
	if err != nil {
		fmt.Printf("Command parse error: %v\n", err)
		os.Exit(1)
	}

	if cp.GetBoolOption("debug") {
		cp.Debug()
	}

	// List of destination images, but not converted into their
	// destination format.
	openPatterns := map[string]*nesimg.PatternTable{}

	for cp.NextInput() {
		// === Gather options ===
		inputFile, err := cp.GetOption("input-filename")
		if err != nil {
			fmt.Printf("Error getting filename: %v\n", err)
			os.Exit(1)
		}

		outputFile, err := cp.GetOption("output")
		if err != nil {
			fmt.Printf("Error getting output name: %v\n", err)
			os.Exit(1)
		}

		if outputFile == "" {
			fmt.Println("Missing output file!")
			os.Exit(1)
		}

		// === Do the things ===
		var pt *nesimg.PatternTable
		inExt := filepath.Ext(inputFile)

		switch strings.ToLower(inExt) {
		case ".bmp":
			pt, err = nesimg.LoadBitmap(inputFile)
		case ".chr":
			pt, err = nesimg.LoadCHR(inputFile)
		default:
			err = fmt.Errorf("Unsupported input format: %q", inExt)
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		count := cp.GetIntOption("tile-count")
		offset := cp.GetIntOption("tile-offset")
		if count != 0 || offset != 0 {
			//fmt.Printf("tile count: %d\ntile offset: %d\n", count, offset)
			if offset > len(pt.Patterns) {
				fmt.Println("Offset larger than pattern table length")
				os.Exit(1)
			}

			if count == 0 {
				count = len(pt.Patterns) - offset
			}

			npt := nesimg.NewPatternTable()
			for i := offset; i < offset+count; i++ {
				npt.AddTile(pt.Patterns[i])
			}

			pt = npt
		}

		rmEmpty := cp.GetBoolOption("remove-empty")

		if cp.GetBoolOption("remove-duplicates") {
			pt.RemoveDuplicates(rmEmpty)
		} else if rmEmpty {
			pt.RemoveEmpty()
		}

		if idfile, err := cp.GetOption("nt-ids"); idfile != "" && err == nil {
			if !(cp.GetBoolOption("remove-empty") || cp.GetBoolOption("remove-duplicates")) {
				fmt.Println("--write-ids cannot be used without the --remove-empty or --remove-duplicates option.  Ignoring.")
				goto SKIP_NT_IDS // i still feel dirty
			}
			if len(pt.ReducedIds) > 512 {
				fmt.Println("More than 512 tiles! Aborting.")
				os.Exit(1)
			}

			f, err := os.Create(idfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer f.Close()

			fmt.Fprintf(f, ": .word %d\n", len(pt.ReducedIds))
			table := []string{}
			for _, val := range pt.ReducedIds {
				table = append(table, fmt.Sprintf("%d", val))
			}
			fmt.Fprintf(f, ": .byte %s\n", strings.Join(table, ","))
		} else if err != nil {
			fmt.Printf("nt-ids error: %v\n", err)
		}
	SKIP_NT_IDS:

		// Use a PatternTable as the intermediate format, not the
		// files's destination format.
		if destPt, ok := openPatterns[outputFile]; !ok {
			openPatterns[outputFile] = pt
		} else {
			destPt.AddPatternTable(pt)
		}
	}

	// Write each Pattern table to its file, converting to the correct format
	// on the fly.
	for name, pt := range openPatterns {
		var data []byte
		ext := filepath.Ext(name)

		pt.PadTileCount(cp.GetIntOption("pad-tiles"))

		switch strings.ToLower(ext) {
		case ".chr":
			data = pt.Chr(cp.GetBoolOption("first-plane"))
		case ".png":
			pt.PadTiles()

			val, err := cp.GetOption("palette")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			pal, err := nesimg.ParseHexPalette(val)
			if err != nil {
				fmt.Printf("Invalid palette values: %v\n", err)
				os.Exit(1)
			}
			pt.SetPalette(pal)

			buff := bytes.NewBuffer([]byte{})
			err = png.Encode(buff, pt)
			data = buff.Bytes()
		case ".asm":
			data = []byte(pt.Asm(cp.GetBoolOption("first-plane")))
		default:
			fmt.Printf("Unsupported output format: %q\n", ext)
			os.Exit(1)
		}

		if idfile, err := cp.GetOption("write-ids"); idfile != "" && err == nil {
			if !(cp.GetBoolOption("remove-empty") || cp.GetBoolOption("remove-duplicates")) {
				fmt.Println("--write-ids cannot be used without the --remove-empty or --remove-duplicates option.  Ignoring.")
				goto SKIP_WRITE_IDS // i feel dirty
			}
			if len(pt.ReducedIds) > 512 {
				fmt.Println("More than 512 tiles! Aborting.")
				os.Exit(1)
			}

			tableA := []string{}
			tableB := []string{}
			for idx, id := range pt.ReducedIds {
				// Tiles on the second pattern table will start as N+255
				// so they need to be corrected.
				if id > 255 {
					id -= 255
				}

				// If a tile is more than 32 columns into a row, it's on
				// the second nametable.
				if idx%64 >= 32 {
					tableB = append(tableB, fmt.Sprintf("%d", id))
				} else {
					tableA = append(tableA, fmt.Sprintf("%d", id))
				}
			}

			f, err := os.Create(idfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer f.Close()

			fmt.Fprintf(f, ": .word %d\n", len(tableA))
			fmt.Fprintf(f, ": .byte %s\n", strings.Join(tableA, ", "))

			if len(tableB) > 0 {
				fmt.Fprintf(f, ": .word %d\n", len(tableB))
				fmt.Fprintf(f, ": .byte %s\n", strings.Join(tableB, ", "))
			}
		}
	SKIP_WRITE_IDS:

		if strings.ToLower(ext) != ".asm" &&
			strings.ToLower(ext) != ".chr" &&
			cp.GetBoolOption("first-plane") {
			fmt.Printf("--first-plane is only usable with the --asm flag.")
			os.Exit(1)
		}

		err = ioutil.WriteFile(name, data, 0777)
		if err != nil {
			fmt.Printf("Error writing file %q: %v\n", name, err)
			os.Exit(1)
		}
	}
}
