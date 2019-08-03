package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zorchenhimer/go-nes/ines"
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

	rom, err := ines.ReadRom(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading rom: %v", err)
		os.Exit(1)
	}

	outfile := filepath.Base(os.Args[1])
	outfile = outfile[:len(outfile)-len(filepath.Ext(outfile))] + ".chr"
	fmt.Printf("Outfile: %s\n", outfile)

	fmt.Println(rom.Debug())
	err = ioutil.WriteFile(outfile, rom.ChrRom, 0777)
	if err != nil {
		fmt.Printf("Error writing CHR to disk: %v", err)
		os.Exit(1)
	}
}

func StripExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}
