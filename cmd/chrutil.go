package main

import (
	"fmt"
	"os"
	//"path/filepath"

	"github.com/zorchenhimer/go-nes/common"
	"github.com/zorchenhimer/go-nes/image"
)

func main() {
	cp := common.NewCommandParser()
	cp.AddOption("output", "o", true, "")
	cp.AddOption("remove-duplicates", "d", false, "false")
	cp.AddOption("8x16-sprites", "", false, "false")
	cp.AddOption("assembly-output", "a", false, "false")
	cp.AddOption("font", "f", false, "false")
	cp.AddOption("text", "t", true, "")
	cp.AddOption("start-id", "i", true, "0")
	cp.AddOption("debug", "", false, "false")

	err := cp.Parse()
	if err != nil {
		fmt.Printf("Command parse error: %v\n", err)
		os.Exit(1)
	}

	// Probably shouldn't ignore this error, lol
	if dbg, _ := cp.GetBoolOption("debug"); dbg {
		cp.Debug()
	}

	// Keep track of open files and close them when we're done.  This is
	// here so we can concatenate multiple input files easier.
	openFiles := map[string]*os.File{}

	for cp.NextInput() {
		// === Gather options ===
		inputFile, err := cp.GetOption("input-filename")
		if err != nil {
			fmt.Printf("Error parsing filename: %v\n", err)
			os.Exit(1)
		}

		outputFile, err := cp.GetOption("output")
		if err != nil {
			fmt.Printf("Error parsing output name: %v\n", err)
			os.Exit(1)
		}

		// === Do the things ===
		pt, err := image.LoadBitmap(inputFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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
