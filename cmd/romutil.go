package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	//"strings"
	"hash/crc32"

	"github.com/alexflint/go-arg"
	ines "github.com/zorchenhimer/go-nes/rom"
	//"github.com/zorchenhimer/go-nes/rom/ines"
	//"github.com/zorchenhimer/go-nes/rom/unif"
)

type MainArgs struct {
	Pack   *CmdPack   `arg:"subcommand:pack" help:"Assemble a complete ROM from pieces"`
	Unpack *CmdUnpack `arg:"subcommand:unpack" help:"Split a rom into pieces"`
	Info   *CmdInfo   `arg:"subcommand:info" help:"Print ROM info"`
}

type CmdPack struct {
	Input  string `arg:"positional,required" help:"Directory containing meta.json and data to assemble"`
	Output string `arg:"-o,--output" default:"" placeholder:"FILENAME" help:"Output ROM filename"`
}

type CmdUnpack struct {
	Input        string `arg:"positional,required" help:"ROM file to split into pieces"`
	Output       string `arg:"-o,--output" default:"" help:"Directory to put the pieces.  Defaults to the name of the input file without the extension."`
	PrgSplitSize int    `arg:"-p,--prg-split" default:"0" help:"PRG split file sizes.  A size of zero does not split the data across multiple files."`
	ChrSplitSize int    `arg:"-c,--chr-split" default:"0" help:"CHR split file sizes.  A size of zero does not split the data across multiple files."`
}

type CmdInfo struct {
	Input string `arg:"positional,required" help:"Input ROM file"`
}

type Metadata struct {
	RomName string
	Header  *ines.Header
	Prg     []string
	Chr     []string `json:",omitempty"`
	Misc    string   `json:",omitempty"`
}

func info(args *CmdInfo) error {
	rom, err := ines.ReadRom(args.Input)
	if err != nil {
		return fmt.Errorf("Error reading rom: %v", err)
	}

	prghsh := crc32.New(crc32.IEEETable)
	chrhsh := crc32.New(crc32.IEEETable)

	_, err = prghsh.Write(rom.PrgRom())
	if err != nil {
		return err
	}

	_, err = chrhsh.Write(rom.ChrRom())
	if err != nil {
		return err
	}

	fmt.Println("PrgSize:     ", rom.Header.PrgSize)
	fmt.Println("ChrSize:     ", rom.Header.ChrSize)
	fmt.Println("MiscSize:    ", rom.Header.ChrSize)
	fmt.Println("Trainer:     ", rom.Header.TrainerPresent)
	fmt.Println("PersistMem:  ", rom.Header.PersistentMemory)
	fmt.Println("Mirroring:   ", rom.Header.Mirroring)
	fmt.Println("NES 2.0:     ", rom.Header.Nes2)
	fmt.Println("Mapper:      ", rom.Header.Mapper)
	fmt.Println("SubMapper:   ", rom.Header.SubMapper)

	if rom.Header.PrgRamSize > 0 {
		fmt.Println("PrgRamSize:  ", 64<<rom.Header.PrgRamSize)
	} else {
		fmt.Println("PrgRamSize:  0")
	}

	if rom.Header.PrgNvramSize > 0 {
		fmt.Println("PrgNvramSize:", 64<<rom.Header.PrgNvramSize)
	} else {
		fmt.Println("PrgNvramSize: 0")
	}

	if rom.Header.ChrRamSize > 0 {
		fmt.Println("ChrRamSize:  ", 64<<rom.Header.ChrRamSize)
	} else {
		fmt.Println("ChrRamSize:   0")
	}

	if rom.Header.ChrNvramSize > 0 {
		fmt.Println("ChrNvramSize:", 64<<rom.Header.ChrNvramSize)
	} else {
		fmt.Println("ChrNvramSize: 0")
	}

	fmt.Printf("PRG CRC:      %08X\n", prghsh.Sum32())
	fmt.Printf("CHR CRC:      %08X\n", chrhsh.Sum32())

	return nil
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
		Header:  rom.Header,
		Prg:     []string{},
		Chr:     []string{},
	}

	if args.PrgSplitSize == 0 {
		args.PrgSplitSize = int(rom.Header.PrgSize) / 1024
	}

	if args.ChrSplitSize == 0 && rom.Header.ChrSize > 0 {
		args.ChrSplitSize = int(rom.Header.ChrSize) / 1024
	}

	if args.PrgSplitSize%8 != 0 {
		return fmt.Errorf("PRG can only be split in multiples of 8kb")
	}

	err, meta.Prg = writeBin(rom.PrgRom(), args.PrgSplitSize, args.Output, "prg")
	if err != nil {
		return fmt.Errorf("Error writing PRG data: %w", err)
	}

	if rom.Header.ChrSize > 0 {
		err, meta.Chr = writeBin(rom.ChrRom(), args.ChrSplitSize, args.Output, "chr")
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
	case args.Info != nil:
		return info(args.Info)
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
