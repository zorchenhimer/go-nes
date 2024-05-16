package rom

import (
	"bytes"
	"fmt"
	"io"
	"os"
	//"github.com/zorchenhimer/go-nes/rom/ines"
	//"github.com/zorchenhimer/go-nes/rom/unif"
)

type Rom interface {
	RomType() RomType
	PrgRom() []byte
	ChrRom() []byte
}

type RomType string

const (
	UNIF RomType = "UNIF"
	INES RomType = "iNES"
	NES2 RomType = "NES 2.0"
)

func LoadRom(filename string) (Rom, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Load(file)
}

func Load(r io.ReadSeeker) (Rom, error) {
	magic := make([]byte, 4)
	_, err := r.Read(magic)
	if err != nil {
		return nil, fmt.Errorf("Unable to read magic value: %w", err)
	}

	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("Seek error: %w", err)
	}

	if bytes.Equal(magic, []byte("UNIF")) {
		return ReadUnif(r)

	} else if bytes.Equal(magic, []byte{0x4E, 0x45, 0x53, 0x1A}) { // NES<EOF>
		return ReadInes(r)
	}

	return nil, fmt.Errorf("Unknown magic bytes: %q 0x%08X", magic, magic)
}
