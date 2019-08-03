package main

import (
	"fmt"
	"os"
	//"path/filepath"

	"github.com/zorchenhimer/go-nes/common"
	"github.com/zorchenhimer/go-nes/image"
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

	// Unimplemented
	cp.AddOption("8x16-sprites", "", false, "false",
		"// TODO")
	cp.AddOption("asm", "a", false, "false",
		"// TODO\nWrite output as assembly instead of binary CHR data.")
	cp.AddOption("text", "t", true, "",
		"// TODO")
	cp.AddOption("start-id", "i", true, "0",
		"// TODO\nStart at this ID when reading the input file.")

	// Assumes --asm --first-plane --remove-duplaciates
	cp.AddOption("font", "f", false, "false",
		"// TODO\nConvert bitmap font to assembly.  Assumes --asm --first-plane --remove-duplaciates")

	// Only write the first bit plane of CHR.  Only usable with --asm.
	cp.AddOption("first-plane", "", false, "false",
		"// TODO\nOnly write the first bit plane of CHR data.  Only usable with --asm.")

	err := cp.Parse()
	if err != nil {
		fmt.Printf("Command parse error: %v\n", err)
		os.Exit(1)
	}

	if cp.GetBoolOption("debug") {
		cp.Debug()
	}

	// Keep track of open files and close them when we're done.  This is
	// here so we can concatenate multiple input files easier.
	openFiles := map[string]*os.File{}

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

		// === Do the things ===
		pt, err := image.LoadBitmap(inputFile)
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

		var file *os.File
		var ok bool

		if file, ok = openFiles[outputFile]; !ok {
			file, err = os.Create(outputFile)
			if err != nil {
				fmt.Printf("Error opening output file %q: %v\n", outputFile, err)
				os.Exit(1)
			}
			openFiles[outputFile] = file
		}

		file.Write(pt.Chr())
	}

	errored := false
	for name, file := range openFiles {
		err = file.Close()
		if err != nil {
			fmt.Printf("Error closing file %q: %v\n", name, err)
			errored = true
		}
	}

	if errored {
		os.Exit(1)
	}
}
