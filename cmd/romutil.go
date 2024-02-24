package main

import (
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	//"strings"

	"github.com/alexflint/go-arg"
	"github.com/zorchenhimer/go-nes/ines"
)

type MainArgs struct {
	Pack *CmdPack `arg:"subcommand:pack"`
	Unpack *CmdUnpack `arg:"subcommand:unpack"`
}

type CmdPack struct {
	Input string `arg:"positional,required"`
	Output string `arg:"-o,--output" default:""`
}

type CmdUnpack struct {
	Input string `arg:"positional,required"`
	Output string `arg:"-o,--output" default:""`
	PrgSplitSize int `arg:"-p,--prg-split" default:"0"`
	ChrSplitSize int `arg:"-c,--chr-split" default:"0"`
}

type Metadata struct {
	RomName string
	Header *ines.Header
	Prg []string
	Chr []string `json:",omitempty"`
	Misc string `json:",omitempty"`
}

func pack(args *CmdPack) error {
	metaraw, err := os.ReadFile(filepath.Join(args.Input, "meta.json"))
	if err != nil {
		return fmt.Errorf("Unable to open meta.json: %w", err)
	}
	meta := &Metadata{}

	err = json.Unmarshal(metaraw, meta)
	if err != nil {
		return fmt.Errorf("Error reading meta.json: %w", err)
	}

	if args.Output == "" {
		args.Output = meta.RomName
	}

	rom := meta.Header.Bytes()
	for _, prg := range meta.Prg {
		infile := filepath.Join(args.Input, prg)
		raw, err := os.ReadFile(infile)
		if err != nil {
			return fmt.Errorf("Error reading %s: %w", infile, err)
		}
		rom = append(rom, raw...)
	}

	for _, chr := range meta.Chr {
		infile := filepath.Join(args.Input, chr)
		raw, err := os.ReadFile(infile)
		if err != nil {
			return fmt.Errorf("Error reading %s: %w", infile, err)
		}
		rom = append(rom, raw...)
	}

	if meta.Misc != "" {
		infile := filepath.Join(args.Input, meta.Misc)
		raw, err := os.ReadFile(infile)
		if err != nil {
			return fmt.Errorf("Error reading %s: %w", infile, err)
		}
		rom = append(rom, raw...)
	}

	return os.WriteFile(args.Output, rom, 0666)
}

func unpack(args *CmdUnpack) error {
	if args.Output == "" {
		ext := filepath.Ext(args.Input)
		args.Output = filepath.Base(args.Input[:len(args.Input)-len(ext)])
	}
	fmt.Println("Unpacking", args.Input, "to", args.Output)

	err := os.MkdirAll(args.Output, 0777)
	if err != nil {
		return err
	}

	rom, err := ines.ReadRom(args.Input)
	if err != nil {
		return fmt.Errorf("Error reading rom: %v", err)
	}

	meta := Metadata{
		RomName: filepath.Base(args.Input),
		Header: rom.Header,
		Prg: []string{},
		Chr: []string{},
	}

	if args.PrgSplitSize == 0 {
		args.PrgSplitSize = int(rom.Header.PrgSize)/1024
	}

	if args.ChrSplitSize == 0 && rom.Header.ChrSize > 0 {
		args.ChrSplitSize = int(rom.Header.ChrSize)/1024
	}

	if args.PrgSplitSize % 8 != 0 {
		return fmt.Errorf("PRG can only be split in multiples of 8kb")
	}

	err, meta.Prg = writeBin(rom.PrgRom, args.PrgSplitSize, args.Output, "prg")
	if err != nil {
		return fmt.Errorf("Error writing PRG data: %w", err)
	}

	if rom.Header.ChrSize > 0 {
		err, meta.Chr = writeBin(rom.ChrRom, args.ChrSplitSize, args.Output, "chr")
		if err != nil {
			return fmt.Errorf("Error writing CHR data: %w", err)
		}
	}

	if rom.Header.MiscSize > 0 {
		err = os.WriteFile(filepath.Join(args.Output, "misc.dat"), rom.MiscRom, 0666)
		if err != nil {
			return fmt.Errorf("Error writing Misc data: %w", err)
		}
		meta.Misc = filepath.Join(args.Output, "misc.dat")
	}

	rawjson, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(args.Output, "meta.json"), rawjson, 0666)
	if err != nil {
		return err
	}

	return nil
}

func writeBin(raw []byte, size int, outdir, prefix string) (error, []string) {
	names := []string{}
	size *= 1024
	for i := 0; i < len(raw)/size; i++ {
		start, end := i*size, (i*size)+size
		//fmt.Printf("[%d] start:%d end:%d len:%d exit:%d\n", i, start, end, len(raw), len(raw)/size)

		var dat []byte
		if i+size > len(raw) {
			dat = raw[start:len(raw)]
		} else {
			dat = raw[start:end]
		}

		outname := filepath.Join(outdir, fmt.Sprintf("%s_%02X.bin", prefix, i))
		names = append(names, filepath.Base(outname))
		err := os.WriteFile(outname, dat, 0666)
		if err != nil {
			return err, nil
		}
	}
	return nil, names
}

func run(args *MainArgs) error {
	switch {
	case args.Pack != nil:
		return pack(args.Pack)
	case args.Unpack != nil:
		return unpack(args.Unpack)
	default:
		return fmt.Errorf("huh?")
	}

	return nil
}

func main() {
	args := &MainArgs{}
	p := arg.MustParse(args)
	if p.Subcommand() == nil {
		fmt.Fprintln(os.Stderr, "Missing command")
		p.WriteUsage(os.Stderr)
		os.Exit(1)
	}

	err := run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
