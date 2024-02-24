package ines

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type ConsoleType uint8

const (
	CT_STANDARD   ConsoleType = 0x00
	CT_VSSYSTEM   ConsoleType = 0x01
	CT_PLAYCHOICE ConsoleType = 0x02
	CT_EXTENDED   ConsoleType = 0x03
)

func (ct ConsoleType) String() string {
	switch ct {
	case CT_STANDARD:
		return "CT_STANDARD"
	case CT_VSSYSTEM:
		return "CT_VSSYSTEM"
	case CT_PLAYCHOICE:
		return "CT_PLAYCHOICE"
	case CT_EXTENDED:
		return "CT_EXTENDED"
	}
	return "CT_UNKNOWN"
}

// Header holds all the metadata for the NES ROM.
type Header struct {
	// Sizes are in bytes
	PrgSize  uint
	ChrSize  uint
	MiscSize uint

	TrainerPresent   bool
	PersistentMemory bool // Battery present?
	Mirroring        MirrorType
	//VSUnisystem      bool
	//PlayChoice_10    bool
	Nes2   bool
	Mapper uint

	//ExtendedConsole bool
	Nes2Mapper uint16
	SubMapper  uint8
	Console    ConsoleType

	// Unshifted values
	PrgRamSize   uint
	PrgNvramSize uint
	ChrRamSize   uint
	ChrNvramSize uint
}

func (h Header) Debug() string {
	return fmt.Sprintf(`Header:
	PrgSize: %d
	ChrSize: %d
	MiscSize: %d
	TrainerPresent: %t
	PersistentMemory: %t
	Mirroring: %s
	Nes2: %t
	Mapper: %d
	Console: %s`,
		h.PrgSize/1024,
		h.ChrSize/1024,
		h.MiscSize,
		h.TrainerPresent,
		h.PersistentMemory,
		h.Mirroring,
		h.Nes2,
		h.Mapper,
		h.Console,
	)
}

// Return the offsets of various parts of the ROM file
func (h Header) RomOffsets() string {
	trainer := "n/a"
	if h.TrainerPresent {
		trainer = "$0010"
	}
	return fmt.Sprintf(`Offsets:
	Trainer: %s
	PrgRom: $%06X
	ChrRom: $%06X`,
		trainer,
		h.PrgStart(),
		h.ChrStart(),
	)
}

func (h Header) WriteMeta(name string) error {
	raw, err := json.MarshalIndent(h, "", "    ")
	if err != nil {
		return fmt.Errorf("Unable to marshal header data: %v", err)
	}

	err = os.WriteFile(name, raw, 0777)
	if err != nil {
		return fmt.Errorf("Unable to write file: %v", err)
	}
	return nil
}

func LoadHeader(data []byte) (*Header, error) {
	h := &Header{}
	err := json.Unmarshal(data, h)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Header) PrgStart() uint {
	if h.TrainerPresent {
		return 16 + 512
	}

	return 16
}

// ChrStart returns the start offset in the rom of the CHR data.  If zero is
// returned, there is no CHR.
func (h *Header) ChrStart() uint {
	if h.ChrSize == 0 {
		return 0
	}

	return h.PrgSize + h.PrgStart()
}

func (h *Header) HasChr() bool {
	if h.ChrStart() == 0 {
		return false
	}
	return true
}

func ParseHeader(raw []byte) (*Header, error) {
	if !bytes.Equal(raw[:4], []byte{0x4E, 0x45, 0x53, 0x1A}) {
		return nil, fmt.Errorf("iNES header constant missing, found 0x%X instead", raw[:3])
	}

	header := &Header{
		Mirroring:        M_HORIZONTAL,
		PersistentMemory: false,
		TrainerPresent:   false,
		Nes2:             false,
		Console:          CT_STANDARD,
	}

	header.PrgSize = uint(raw[4]) * 16 * 1024
	header.ChrSize = uint(raw[5]) * 8 * 1024

	//fmt.Printf("PRG: %d 0x%X\nCHR: %d 0x%X\n", header.PRGSize, nesraw[4], header.CHRSize, nesraw[5])

	flagSix := raw[6]
	if (flagSix & 0x01) == 0x01 {
		header.Mirroring = M_VERTICAL
	}

	header.PersistentMemory = (flagSix & 0x02) == 0x02
	header.TrainerPresent = (flagSix & 0x04) == 0x04

	// Hard-wired four-screen mode
	if (flagSix & 0x08) == 0x08 {
		header.Mirroring = M_IGNORE
	}

	flagSeven := raw[7]
	header.Console = ConsoleType(flagSeven & 0x03)

	header.Nes2 = (flagSeven & 0x0C) == 0x08

	// Read the MSB size bytes
	if header.Nes2 {
		header.PrgSize = uint(uint16(raw[4])|((uint16(raw[9])&0x0F)<<8)) * 16 * 1024
		header.ChrSize = uint(uint16(raw[5])|((uint16(raw[9])&0xF0)<<8)) * 8 * 1024
	}

	uppermap := flagSeven & 0xF0

	lowermap := flagSix & 0xF0
	lowermap = lowermap >> 4

	header.Mapper = uint(lowermap | uppermap)

	flags8 := raw[8]
	header.Nes2Mapper = uint16(header.Mapper)
	n2map := uint16(flags8&0xF) << 12
	header.Nes2Mapper = header.Nes2Mapper | n2map

	header.SubMapper = uint8(flags8 >> 4)

	// PrgRam
	if raw[10]&0xF0 != 0 {
		shift := uint(raw[10] >> 4)
		//header.PrgRamSize = uint(64 << shift)
		header.PrgRamSize = shift
	}

	// PRG-NVRAM/EEPROM
	if raw[10]&0x0F != 0 {
		shift := uint(raw[10] & 0x0F)
		//header.PrgNvramSize = uint(64 << shift)
		header.PrgNvramSize = shift
	}

	if raw[11]&0xF0 != 0 {
		shift := uint(raw[11] >> 4)
		//header.ChrRamSize = uint(64 << shift)
		header.ChrRamSize = shift
	}

	if raw[11]&0x0F != 0 {
		shift := uint(raw[11] & 0x0F)
		//header.ChrNvramSize = uint(64 << shift)
		header.ChrNvramSize = shift
	}

	return header, nil
}

func (h Header) Bytes() []byte {
	data := []byte{0x4E, 0x45, 0x53, 0x1A}
	prg := h.PrgSize / 1024 / 16
	chr := h.ChrSize / 1024 / 8

	data = append(data, uint8(prg))
	data = append(data, uint8(chr))

	flagSix := uint8(0)
	if h.Mirroring == M_VERTICAL {
		flagSix |= 0x01
	}

	if h.PersistentMemory {
		flagSix |= 0x02
	}

	if h.TrainerPresent {
		flagSix |= 0x04
	}

	if h.Mirroring == M_IGNORE {
		flagSix |= 0x08
	}

	lowermap := (h.Mapper & 0x0F) << 4
	uppermap := (h.Mapper & 0xF0)
	highmap := (h.Mapper & 0xF00) >> 8
	flagSix |= uint8(lowermap)

	data = append(data, flagSix)

	flagSeven := uint8(h.Console)

	if h.Nes2 {
		flagSeven |= 0x08
	}

	flagSeven |= uint8(uppermap)
	data = append(data, flagSeven)

	flagEight := uint8(h.SubMapper<<4 | uint8(highmap))
	data = append(data, flagEight)

	// TODO: flagNine (PRG/CHR ROM MSB)
	data = append(data, 0x00)

	flagTen := uint8(h.PrgRamSize<<4 | h.PrgNvramSize&0x0F)
	data = append(data, flagTen)

	flagEleven := uint8(h.ChrRamSize<<4 | h.ChrNvramSize&0x0F)
	data = append(data, flagEleven)

	for len(data) < 16 {
		data = append(data, 0x00)
	}

	return data
}

type MirrorType uint

const (
	M_HORIZONTAL MirrorType = 0
	M_VERTICAL   MirrorType = 1
	M_IGNORE     MirrorType = 2 // four-screen VRAM
)

func (mt MirrorType) String() string {
	switch mt {
	case M_HORIZONTAL:
		return "Horizontal"
	case M_VERTICAL:
		return "Vertical"
	case M_IGNORE:
		return "Ignore"
	default:
		return fmt.Sprintf("Unknown (%d)", mt)
	}
}
