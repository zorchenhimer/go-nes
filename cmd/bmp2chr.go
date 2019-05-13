package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zorchenhimer/go-nes/image"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing input file")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		fmt.Println("Too many input files")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := StripExtension(inputFile) + ".chr"

	table, err := image.LoadBitmap(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	before, after := table.RemoveDuplicates()

	fmt.Println(table.Debug())
	fmt.Printf("tiles before: %d, after: %d\n", before, after)

	err = table.WriteFile(outputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func StripExtension(filename string) string {
	return filename[:len(filename) - len(filepath.Ext(filename))]
}
