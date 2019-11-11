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

	// Only write the first bit plane of CHR.  Only usable with --asm.
	cp.AddOption("first-plane", "", false, "false",
		"// TODO\nOnly write the first bit plane of CHR data.  Only usable with --asm.")

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

		rmEmpty := cp.GetBoolOption("remove-empty")

		if cp.GetBoolOption("remove-duplicates") {
			pt.RemoveDuplicates(rmEmpty)
		} else if rmEmpty {
			pt.RemoveEmpty()
		}

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

		switch strings.ToLower(ext) {
		case ".chr":
			data = pt.Chr(cp.GetBoolOption("first-plane"))
		case ".png":
			pt.PadTiles()
			buff := bytes.NewBuffer([]byte{})
			err = png.Encode(buff, pt)
			data = buff.Bytes()
		case ".asm":
			data = []byte(pt.Asm(cp.GetBoolOption("first-plane")))
		default:
			fmt.Printf("Unsupported output format: %q\n", ext)
			os.Exit(1)
		}

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
