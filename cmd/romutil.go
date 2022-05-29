package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/ines"
)

const usage string = `NES ROM Utility
Usage: %s <command> <input> [options]

Commands:
	unpack
		Unpack a ROM into a directory
	pack
		Pack an unpacked ROM given a directory
	info
		Print the header info about the ROM.
	usage // TODO
	nes2  // TODO
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println(usage)
		os.Exit(1)
	}

	var err error

	switch strings.ToLower(os.Args[1]) {
	case "unpack":
		err = cmdUnpack(os.Args[2:]...) //len(os.Args)
	case "pack":
		// TODO: test this and clean it up
		dir := strings.Trim(os.Args[2], `/\`) + "/"
		cmdPack(dir, "packed.nes")
	case "info":
		cmdInfo(os.Args[2])
	case "usage":
	case "nes2":
	default:
		fmt.Printf("Invalid command: %q\n", os.Args[2])
		// TODO: print usage
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func cmdPack(dirName, romName string) {
	headerRaw, err := os.ReadFile(dirName + "header.json")
	if err != nil {
		fmt.Printf("Unable to read header.json file: %v\n", err)
		os.Exit(1)
	}

	header, err := ines.LoadHeader(headerRaw)
	if err != nil {
		fmt.Printf("Unable to load header data: %v\n", err)
		os.Exit(1)
	}

	prgRaw, err := os.ReadFile(dirName + "prg.dat")
	if err != nil {
		fmt.Printf("Unable to read prg.dat file: %v\n", err)
		os.Exit(1)
	}

	chr := []byte{}
	for i := 0; i < 16; i++ {
		chrfile := fmt.Sprintf("bank_%02d.chr", i)
		fmt.Println(chrfile)
		chrRaw, err := os.ReadFile(dirName + chrfile)
		if err != nil {
			fmt.Printf("Unable to read %s file: %v\n", chrfile, err)
			os.Exit(1)
		}

		chr = append(chr, chrRaw...)
	}

	rom := ines.NesRom{
		Header: header,
		PrgRom: prgRaw,
		ChrRom: chr,
	}

	err = rom.WriteFile(romName)
	if err != nil {
		fmt.Printf("Unable to write rom: %v\n", err)
		os.Exit(1)
	}
}

func cmdUnpack(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing filename")
	}

	filename := args[0]
	outdir := filepath.Base(filename)
	outdir = outdir[:len(outdir)-len(filepath.Ext(outdir))] + "/"

	var (
		SplitPrg     bool
		SplitChr     bool
		OutDirectory string
	)

	fs := flag.NewFlagSet("unpacking", 0)
	fs.BoolVar(&SplitPrg, "split-prg", false, "Split the PRG data into 16kb chunks")
	fs.BoolVar(&SplitChr, "split-chr", false, "Split the CHR data into 4kb chunks")
	fs.StringVar(&OutDirectory, "directory", outdir, "Output directory")

	var err error
	if len(args) > 1 {
		err = fs.Parse(args[1:])
	} else {
		err = fs.Parse([]string{})
	}

	if err != nil {
		return err
	}

	if !strings.HasSuffix(OutDirectory, "/") {
		OutDirectory = OutDirectory + "/"
	}

	err = os.MkdirAll(OutDirectory, 0777)
	if err != nil {
		return fmt.Errorf("Unable to create output directory: %v", err)
	}

	rom, err := ines.ReadRom(filename)
	if err != nil {
		return fmt.Errorf("Error reading rom: %v", err)
	}

	err = rom.Header.WriteMeta(OutDirectory + "header.json")
	if err != nil {
		return fmt.Errorf("Error writing header: %v", err)
	}

	if SplitPrg {
		size := 16 * 1024
		for i := 0; i < len(rom.PrgRom)/size; i++ {
			var raw []byte
			start, end := i*size, (i*size)+size

			if i+size > len(rom.PrgRom) {
				raw = rom.PrgRom[start:len(rom.PrgRom)]
			} else {
				raw = rom.PrgRom[start:end]
			}

			err = os.WriteFile(fmt.Sprintf("%sprg_%d.dat", OutDirectory, i), raw, 0777)
			if err != nil {
				return fmt.Errorf("Error writing PRG data: %v", err)
			}
		}
	} else {
		err = os.WriteFile(OutDirectory+"prg.dat", rom.PrgRom, 0777)
		if err != nil {
			return fmt.Errorf("Error writing PRG data: %v", err)
		}
	}

	if rom.Header.ChrSize > 0 {
		if SplitChr {
			size := 4 * 1024
			for i := 0; i < len(rom.ChrRom)/size; i++ {
				var raw []byte
				start, end := i*size, (i*size)+size

				if i+size > len(rom.ChrRom) {
					raw = rom.ChrRom[start:len(rom.ChrRom)]
				} else {
					raw = rom.ChrRom[start:end]
				}

				err = os.WriteFile(fmt.Sprintf("%schr_%d.dat", OutDirectory, i), raw, 0777)
				if err != nil {
					return fmt.Errorf("Error writing CHR data: %v", err)
				}
			}
		} else {
			err = os.WriteFile(OutDirectory+"chr.dat", rom.ChrRom, 0777)
			if err != nil {
				return fmt.Errorf("Error writing CHR data: %v", err)
			}
		}
	}

	if rom.Header.MiscSize > 0 {
		err = os.WriteFile(OutDirectory+"misc.dat", rom.MiscRom, 0777)
		if err != nil {
			return fmt.Errorf("Error writing MISC data: %v", err)
		}
	}

	return nil
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
