package ines

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"
)

// NesRom holds the complete decoded ROM image.
type NesRom struct {
	Header *Header

	PrgRom  []byte
	ChrRom  []byte
	MiscRom []byte // data after the CHR rom

	PrgCrc  uint32
	ChrCrc  uint32
	MiscCrc uint32
	RomCrc	uint32
}

func (r *NesRom) Debug() string {
	return r.Header.Debug() +
	fmt.Sprintf("\nRomCrc: %08X\nPrgCrc: %08X\nChrCrc: %08X\nMiscCrc: %08X", r.RomCrc, r.PrgCrc, r.ChrCrc, r.MiscCrc)
}

func (r *NesRom) WriteFile(filename string) error {
	if r.Header.ChrSize != uint(len(r.ChrRom)) {
		return fmt.Errorf("CHR Size missmatch expected $%04X, found $%04X", r.Header.ChrSize, len(r.ChrRom))
	}

	if r.Header.PrgSize != uint(len(r.PrgRom)) {
		return fmt.Errorf("PRG Size missmatch expected $%04X, found $%04X", r.Header.PrgSize, len(r.PrgRom))
	}

	data := r.Header.Bytes()
	data = append(data, r.PrgRom...)
	data = append(data, r.ChrRom...)

	return ioutil.WriteFile(filename, data, 0777)
}

// ReadRom() opens the given file and attempts to load it as an iNES ROM.
func ReadRom(filename string) (*NesRom, error) {
	rawrom, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error reading ROM file: %v", err)
	}

	rom := &NesRom{}

	h, err := ParseHeader(rawrom[:16])
	if err != nil {
		return nil, fmt.Errorf("Error parsing header: %v", err)
	}

	rom.Header = h

	rom.RomCrc = crc32.ChecksumIEEE(rawrom[16:len(rawrom)])

	prgEnd := rom.Header.PrgStart() + rom.Header.PrgSize
	chrEnd := rom.Header.ChrStart() + rom.Header.ChrSize

	if chrEnd > uint(len(rawrom)) {
		return nil, fmt.Errorf("Sizes too large:\n  chrEnd: %d\n  len(nesraw): %d\n", chrEnd, len(rawrom))
	}

	rom.PrgRom = rawrom[rom.Header.PrgStart():prgEnd]
	if rom.Header.HasChr() {
		rom.ChrRom = rawrom[rom.Header.ChrStart():chrEnd]
	}

	rom.PrgCrc = crc32.ChecksumIEEE(rom.PrgRom)
	if rom.Header.HasChr() {
		fmt.Println("HasChr(), grabbing CRC")
		rom.ChrCrc = crc32.ChecksumIEEE(rom.ChrRom)
	}

	return rom, nil
}

func (r *NesRom) PrgCrcString() string {
	return fmt.Sprintf("%08X", r.PrgCrc)
}

func (r *NesRom) ChrCrcString() string {
	return fmt.Sprintf("%08X", r.ChrCrc)
}
