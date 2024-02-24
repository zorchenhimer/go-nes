package main

// Mesen2 workspace to da65 config

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/zorchenhimer/go-nes/mesen"
)

type Args struct {
	Input     string `arg:"positional,required" help:"Mesen2 JSON workspace file"`
	Output    string `arg:"positional" help:"Output config file [default: STDOUT]"`
	PrgStart  string `arg:"--prg-start" default:"$8000" help:"Offset to add to PRG ROM address values"`
	prgoffset uint
}

func (a Args) Description() string {
	return `Extract labels from Mesen2 debugger workspace json files and write them as da65 config values.`
}

func main() {
	args := &Args{}
	arg.MustParse(args)

	if args.PrgStart != "" {
		off, err := parseOffset(args.PrgStart)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		args.prgoffset = off
	}

	err := run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOffset(input string) (uint, error) {
	base := 10
	re_hex := regexp.MustCompile(`^(?:\$|0x)([a-fA-F0-9]+)$`)
	match := re_hex.FindStringSubmatch(input)
	if len(match) == 2 {
		base = 16
		input = match[1]
	}

	u64, err := strconv.ParseUint(input, base, 32)
	if err != nil {
		return 0, err
	}

	return uint(u64), nil
}

func run(args *Args) error {

	ws, err := readWorkspace(args.Input)
	if err != nil {
		return err
	}

	if args.Output == "" {
		return fmt.Errorf("Missing input file")
	}

	file, err := os.Create(args.Output)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, label := range ws.Labels {
		addr := label.Address
		if label.MemoryType == "NesPrgRom" {
			addr += args.prgoffset
		}

		items := []string{
			fmt.Sprintf("ADDR $%04X;", addr),
			fmt.Sprintf("NAME %q;", label.Label),
		}

		if label.Length > 1 {
			items = append(items, fmt.Sprintf("SIZE %d;", label.Length))
		}

		if label.Comment != "" {
			items = append(items, fmt.Sprintf("COMMENT %q;",
				strings.ReplaceAll(label.Comment, "\n", "\n;")))
		}

		fmt.Fprintf(file, "LABEL { %s };\n", strings.Join(items, " "))
	}

	return nil
}

func readWorkspace(source string) (*mesen.Workspace, error) {
	var input io.Reader

	if source == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(source)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		input = f
	}

	ws, err := mesen.LoadWorkspace(input)
	if err != nil {
		return nil, fmt.Errorf("unable to load workspace: %w", err)
	}

	return ws, nil
}
