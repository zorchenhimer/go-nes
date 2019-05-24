package main

import (
	"fmt"
	"os"
	"io/ioutil"
	//"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/ines"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing input directory")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		fmt.Println("Too many input files")
		os.Exit(1)
	}

	dir := strings.Trim(os.Args[1], `/\`) + "/"

	headerRaw, err := ioutil.ReadFile(dir + "header.json")
	if err != nil {
		fmt.Printf("Unable to read header.json file: %v\n", err)
		os.Exit(1)
	}

	header, err := ines.LoadHeader(headerRaw)
	if err != nil {
		fmt.Printf("Unable to load header data: %v\n", err)
		os.Exit(1)
	}

	prgRaw, err := ioutil.ReadFile(dir + "prg.dat")
	if err != nil {
		fmt.Printf("Unable to read prg.dat file: %v\n", err)
		os.Exit(1)
	}

	chr := []byte{}
	for i := 0; i < 16; i++ {
		chrfile := fmt.Sprintf("bank_%02d.chr", i)
		fmt.Println(chrfile)
		chrRaw, err := ioutil.ReadFile(dir + chrfile)
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

	err = rom.WriteFile("packed.nes")
	if err != nil {
		fmt.Printf("Unable to write rom: %v\n", err)
		os.Exit(1)
	}

}
