package rom

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

type Crc32 uint32

func (crc Crc32) HexString() string {
	return fmt.Sprintf("%08X", crc)
}

// NesRom holds the complete decoded ROM image.
type NesRom struct {
	Header *Header

	Prgrom  []byte
	Chrrom  []byte
	MiscRom []byte // data after the CHR rom

	PrgCrc  Crc32
	ChrCrc  Crc32
	MiscCrc Crc32
	RomCrc  Crc32
}

func (r *NesRom) ChrRom() []byte {
	return r.Chrrom
}

func (r *NesRom) PrgRom() []byte {
	return r.Prgrom
}

func (r *NesRom) RomType() RomType {
	if r.Header.Nes2 {
		return NES2
	}
	return INES
}

func (r *NesRom) Debug() string {
	return r.Header.Debug() +
		fmt.Sprintf("\nRomCrc: %s\nPrgCrc: %s\nChrCrc: %s\nMiscCrc: %s", r.RomCrc.HexString(), r.PrgCrc.HexString(), r.ChrCrc.HexString(), r.MiscCrc.HexString())
}

func (r *NesRom) WriteFile(filename string) error {
	if r.Header.ChrSize != uint(len(r.Chrrom)) {
		return fmt.Errorf("CHR Size missmatch expected $%04X, found $%04X", r.Header.ChrSize, len(r.Chrrom))
	}

	if r.Header.PrgSize != uint(len(r.Prgrom)) {
		return fmt.Errorf("PRG Size missmatch expected $%04X, found $%04X", r.Header.PrgSize, len(r.Prgrom))
	}

	data := r.Header.Bytes()
	data = append(data, r.Prgrom...)
	data = append(data, r.Chrrom...)

	return os.WriteFile(filename, data, 0777)
}

// ReadRom() opens the given file and attempts to load it as an iNES ROM.
func ReadRom(filename string) (*NesRom, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open %q: %w", err)
	}
	defer file.Close()

	return ReadInes(file)
}

func ReadInes(reader io.Reader) (*NesRom, error) {
	rawrom, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Error reading ROM: %w", err)
	}

	rom := &NesRom{}

	h, err := ParseHeader(rawrom[:16])
	if err != nil {
		return nil, fmt.Errorf("Error parsing header: %v", err)
	}

	rom.Header = h

	rom.RomCrc = Crc32(crc32.ChecksumIEEE(rawrom[16:len(rawrom)]))

	prgEnd := rom.Header.PrgStart() + rom.Header.PrgSize
	chrEnd := rom.Header.ChrStart() + rom.Header.ChrSize

	if chrEnd > uint(len(rawrom)) {
		return nil, fmt.Errorf("Sizes too large: chrEnd:%d len(nesraw):%d", chrEnd, len(rawrom))
	}

	rom.Prgrom = rawrom[rom.Header.PrgStart():prgEnd]
	if rom.Header.HasChr() {
		rom.Chrrom = rawrom[rom.Header.ChrStart():chrEnd]
	}

	rom.PrgCrc = Crc32(crc32.ChecksumIEEE(rom.Prgrom))
	if rom.Header.HasChr() {
		rom.ChrCrc = Crc32(crc32.ChecksumIEEE(rom.Chrrom))
	}

	return rom, nil
}

func (r *NesRom) PrgCrcString() string {
	return fmt.Sprintf("%08X", r.PrgCrc)
}

func (r *NesRom) ChrCrcString() string {
	return fmt.Sprintf("%08X", r.ChrCrc)
}
