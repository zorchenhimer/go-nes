package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/ines"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing command")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Missing input file")
		os.Exit(1)
	}

	switch strings.ToLower(os.Args[1]) {
	case "unpack":
		cmdUnpack(os.Args[2])
	case "pack":
		cmdPack(os.Args[2])
	case "info":
		cmdInfo(os.Args[2])
	case "usage":
	case "nes2":
	default:
		fmt.Printf("Invalid command: %q\n", os.Args[2])
		// TODO: print usage
		os.Exit(1)
	}
}

func cmdPack(dirName string) {
	// TODO
	fmt.Printf("TODO: Pack directory %q\n", dirName)
}

func cmdUnpack(filename string) {
	outdir := filepath.Base(filename)
	outdir = outdir[:len(outdir)-len(filepath.Ext(outdir))] + "/"
	err := os.MkdirAll(outdir, 0777)
	if err != nil {
		fmt.Printf("Unable to create output directory: %v", err)
		os.Exit(1)
	}

	rom, err := ines.ReadRom(filename)
	if err != nil {
		fmt.Printf("Error reading rom: %v", err)
		os.Exit(1)
	}

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
}

func cmdInfo(filename string) {
	rom, err := ines.ReadRom(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(rom.Debug())
	fmt.Println(rom.Header.RomOffsets())
}
