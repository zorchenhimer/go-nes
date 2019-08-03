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

	outdir := filepath.Base(os.Args[1])
	outdir = outdir[:len(outdir)-len(filepath.Ext(outdir))] + "/"
	err := os.MkdirAll(outdir, 0777)
	if err != nil {
		fmt.Printf("Unable to create output directory: %v", err)
		os.Exit(1)
	}

	rom, err := ines.ReadRom(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading rom: %v", err)
		os.Exit(1)
	}

	fmt.Println(rom.Debug())
	fmt.Println(rom.Header.RomOffsets())

	err = rom.Header.WriteMeta(outdir + "header.json")
	if err != nil {
		fmt.Printf("Error writing header: %v", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(outdir+"prg.dat", rom.PrgRom, 0777)
	if err != nil {
		fmt.Printf("Error writing PRG data: %v", err)
		os.Exit(1)
	}

	if rom.Header.ChrSize > 0 {
		err = ioutil.WriteFile(outdir+"chr.dat", rom.ChrRom, 0777)
		if err != nil {
			fmt.Printf("Error writing CHR data: %v", err)
			os.Exit(1)
		}
	}

	if rom.Header.MiscSize > 0 {
		err = ioutil.WriteFile(outdir+"misc.dat", rom.MiscRom, 0777)
		if err != nil {
			fmt.Printf("Error writing MISC data: %v", err)
			os.Exit(1)
		}
	}

	// Split CHR into 8k chunks
	if rom.Header.ChrSize == 0 || rom.Header.ChrSize%8 != 0 {
		fmt.Printf("Unexpected CHR size %d ($%04X). Unable to split.", rom.Header.ChrSize, rom.Header.ChrSize)
		os.Exit(1)
	}

	chunkSize := 0x2000
	num := 0
	for offset := 0; offset < int(rom.Header.ChrSize); offset += chunkSize {
		err = ioutil.WriteFile(fmt.Sprintf("%sbank_%02d.chr", outdir, num), rom.ChrRom[offset:offset+chunkSize], 0777)
		if err != nil {
			fmt.Printf("Unable to write bank_%02X.chr: %v", err)
			os.Exit(1)
		}
		num += 1
	}
}
