package ines

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	//"io"
	"io/ioutil"
	"math"
	"os"
)

const NesFilename string = "Barbie.nes"
const RomMaskFile string = "Barbie.zip.Barbie.cdl"

// Header holds all the metadata for the NES ROM.
type Header struct {
	// Sizes are in bytes
	PRGSize  uint
	CHRSize  uint
	MiscSize uint

	TrainerPresent   bool
	PersistentMemory bool
	Mirroring        MirrorType
	VSUnisystem      bool
	PlayChoice_10    bool
	Nes2             bool
	Mapper           uint
}

type MirrorType uint

const (
	M_HORIZONTAL MirrorType = 0
	M_VERTICAL   MirrorType = 1
	M_IGNORE     MirrorType = 2 // four-screen VRAM
)

//type PRGUsage byte
//type CHRUsage byte
const (
	// PRG ROM
	UP_CODE    byte = 0x01
	UP_DATA    byte = 0x02
	UP_JMPCODE byte = 0x10 // Indirect code; via JMP instruction
	UP_LDADATA byte = 0x20 // Indirect data; via LDA instruction
	UP_AUDIO   byte = 0x40
	UP_UNKNOWN byte = 0x00

	// CHR ROM
	UC_DRAWN  byte = 0x01 // rendered by PPU
	UC_READ   byte = 0x02 // Read via port $2007
	UC_UNKOWN byte = 0x00
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

	prgEnd := rom.Header.PrgStart() + rom.Header.PRGSize
	chrEnd := rom.Header.ChrStart() + rom.Header.CHRSize

	if chrEnd > uint(len(rawrom)) {
		return nil, fmt.Errorf("Sizes too large:\n  chrEnd: %d\n  len(nesraw): %d\n", chrEnd, len(rawrom))
	}

	rom.PrgRom = rawrom[rom.Header.PrgStart():prgEnd]
	if rom.Header.HasChr() {
		rom.ChrRom = rawrom[rom.Header.ChrStart():chrEnd]
	}

	rom.PrgCrc = crc32.ChecksumIEEE(rom.PrgRom)
	if rom.Header.HasChr() {
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

func (h *Header) PrgStart() uint {
	if h.TrainerPresent {
		return 16 + 512
	}

	return 16
}

// ChrStart returns the start offset in the rom of the CHR data.  If zero is
// returned, there is no CHR.
func (h *Header) ChrStart() uint {
	if h.CHRSize == 0 {
		return 0
	}

	return h.PRGSize + h.PrgStart()
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
		VSUnisystem:      false,
		PlayChoice_10:    false,
		Nes2:             false,
	}

	header.PRGSize = uint(raw[4]) * 16 * 1024
	header.CHRSize = uint(raw[5]) * 8 * 1024

	//fmt.Printf("PRG: %d 0x%X\nCHR: %d 0x%X\n", header.PRGSize, nesraw[4], header.CHRSize, nesraw[5])

	flagSix := raw[6]
	if (flagSix & 0x01) == 0x01 {
		header.Mirroring = M_VERTICAL
	}

	if (flagSix & 0x02) == 0x02 {
		header.PersistentMemory = true
	}

	if (flagSix & 0x03) == 0x03 {
		header.TrainerPresent = true
		//fmt.Println("Trainer is present")
	}

	if (flagSix & 0x04) == 0x04 {
		header.Mirroring = M_IGNORE
	}

	flagSeven := raw[7]
	if (flagSeven & 0x01) == 0x01 {
		header.VSUnisystem = true
	}

	if (flagSeven & 0x02) == 0x02 {
		header.PlayChoice_10 = true
	}

	if (flagSeven & 0x03) == 0x03 {
		header.Nes2 = true
	}

	uppermap := flagSeven & 0xF0

	lowermap := flagSix & 0xF0
	lowermap = lowermap >> 4

	header.Mapper = uint(lowermap | uppermap)

	return header, nil
}

func main() {
	//	nesraw, err := ioutil.ReadFile(NesFilename)
	//	if err != nil {
	//		fmt.Println("Unabte to read .nes file: ", err)
	//		return
	//	}
	//
	//	if !bytes.Equal(nesraw[:4], []byte{0x4E, 0x45, 0x53, 0x1A}) {
	//		fmt.Println("iNES header constant missing")
	//		fmt.Printf("%X\n", nesraw[:3])
	//		return
	//	}
	//
	//	header.PRGSize = uint(nesraw[4]) * 16 * 1024
	//	header.CHRSize = uint(nesraw[5]) * 8 * 1024
	//
	//	fmt.Printf("PRG: %d 0x%X\nCHR: %d 0x%X\n", header.PRGSize, nesraw[4], header.CHRSize, nesraw[5])
	//
	//	flagSix := nesraw[6]
	//	fmt.Printf("flacgSix: 0x%X\n", flagSix)
	//
	//	if (flagSix & 0x01) == 0x01 {
	//		header.Mirroring = M_VERTICAL
	//	}
	//
	//	if (flagSix & 0x02) == 0x02 {
	//		header.PersistentMemory = true
	//	}
	//
	//	if (flagSix & 0x03) == 0x03 {
	//		header.TrainerPresent = true
	//		fmt.Println("Trainer is present")
	//	}
	//
	//	if (flagSix & 0x04) == 0x04 {
	//		header.Mirroring = M_IGNORE
	//	}
	//
	//	lowermap := flagSix & 0xF0
	//	fmt.Printf("lowermap: %X\n", lowermap)
	//
	//	flagSeven := nesraw[7]
	//
	//	if (flagSeven & 0x01) == 0x01 {
	//		header.VSUnisystem = true
	//	}
	//
	//	if (flagSeven & 0x02) == 0x02 {
	//		header.PlayChoice_10 = true
	//	}
	//
	//	if (flagSeven & 0x03) == 0x03 {
	//		header.Nes2 = true
	//	}
	//
	//	uppermap := flagSeven & 0xF0
	//	fmt.Printf("uppermap: %X\n", uppermap)
	//
	//	lowermap = lowermap >> 4
	//	mapnum := lowermap | uppermap
	//
	//	fmt.Printf("mapnum: 0x%X (%d)\n", mapnum, int(mapnum))
	//
	//	datastart := uint(16)
	//	if header.TrainerPresent {
	//		datastart += 512
	//	}
	//
	//	prgEnd := datastart + header.PRGSize
	//	chrEnd := prgEnd + header.CHRSize
	//
	//	if chrEnd > uint(len(nesraw)) {
	//		fmt.Printf("Sizes too large:\n  chrEnd: %d\n  len(nesraw): %d\n", chrEnd, len(nesraw))
	//		return
	//	}
	//
	//	prgData := nesraw[datastart:prgEnd]
	//	chrData := nesraw[prgEnd:chrEnd]
	//
	//	cdlraw, err := ioutil.ReadFile("Barbie.zip.Barbie.cdl")
	//	if err != nil {
	//		fmt.Println("Unable to open Code/Data Log file: ", err)
	//		return
	//		//} else if len(cdlraw) != len(nesraw) {
	//		//    fmt.Printf("nesraw length doesn't match Code/Data Log length\n%d\n%d\n", len(nesraw), len(cdlraw))
	//		//    return
	//	}
	//
	//	prgFilteredCode := []byte{}
	//	prgFilteredData := []byte{}
	//	chrFilteredDrawn := []byte{}
	//	chrFilteredRead := []byte{}
	//	unkFiltered := []byte{}
	//
	//	//headerRaw := []byte{}
	//
	//	for i := 0; i < len(cdlraw); i++ {
	//		if i+144 > len(nesraw) {
	//			break
	//		}
	//
	//		nesbyte := nesraw[i+16]
	//		if (cdlraw[i]&UP_CODE == UP_CODE) || (cdlraw[i]&UP_JMPCODE == UP_JMPCODE) {
	//			prgFilteredCode = append(prgFilteredCode, nesbyte)
	//		} else {
	//			//prgFilteredCode = append(prgFilteredCode, 0x00)
	//		}
	//
	//		if (cdlraw[i]&UP_DATA == UP_DATA) || (cdlraw[i]&UP_LDADATA == UP_LDADATA) {
	//			prgFilteredData = append(prgFilteredData, nesbyte)
	//		} else {
	//			//prgFilteredData = append(prgFilteredData, 0x00)
	//		}
	//
	//		if cdlraw[i]&UC_DRAWN == UC_DRAWN {
	//			chrFilteredDrawn = append(chrFilteredDrawn, nesbyte)
	//		} else {
	//			chrFilteredDrawn = append(chrFilteredDrawn, 0x00)
	//		}
	//
	//		if cdlraw[i]&UC_READ == UC_READ {
	//			chrFilteredRead = append(chrFilteredRead, nesbyte)
	//		} else {
	//			chrFilteredRead = append(chrFilteredRead, 0x00)
	//		}
	//
	//		if cdlraw[i] == UP_UNKNOWN {
	//			unkFiltered = append(unkFiltered, nesbyte)
	//		} else {
	//			unkFiltered = append(unkFiltered, 0x00)
	//		}
	//	}
	//
	//	fmt.Printf("\ndatastart: %d\nprgData: %d\nchrData: %d\n\n", datastart, len(prgData), len(chrData))
	//	fmt.Printf("chrEnd: %d\nnes:    %d\n\n", chrEnd, len(nesraw))
	//	diff := len(nesraw) - int(chrEnd)
	//	if diff == 128 || diff == 127 {
	//		fmt.Printf("Found title tail: %s\n", nesraw[chrEnd:])
	//	} else if diff > 0 {
	//		fmt.Printf("Found something at end: %s\n", nesraw[chrEnd:])
	//	}
	//
	//	// Write raw data
	//	if err = ioutil.WriteFile("barbie.prg", prgData, 0777); err != nil {
	//		fmt.Println("Unable to write PRG file: ", err)
	//		return
	//	}
	//
	//	if err = ioutil.WriteFile("barbie.chr", chrData, 0777); err != nil {
	//		fmt.Println("Unable to write CHR file: ", err)
	//		return
	//	}
	//
	//	if err = ioutil.WriteFile("barbie.prg.code", prgFilteredCode, 0777); err != nil {
	//		fmt.Println("Unable to write filtered prg code file: ", err)
	//		return
	//	}
	//
	//	if err = ioutil.WriteFile("barbie.prg.data", prgFilteredData, 0777); err != nil {
	//		fmt.Println("Unable to write filtered prg data file: ", err)
	//		return
	//	}
	//
	//	if err = ioutil.WriteFile("barbie.chr.drawn", chrFilteredDrawn, 0777); err != nil {
	//		fmt.Println("Unable to write filtered chr drawn file: ", err)
	//		return
	//	}
	//
	//	if err = ioutil.WriteFile("barbie.chr.read", chrFilteredRead, 0777); err != nil {
	//		fmt.Println("Unable to write filtered chr read file: ", err)
	//		return
	//	}
	//
	//	// Write out images
	//	if err = ImgFromBin("prg.png", prgData); err != nil {
	//		fmt.Println("Unable to write PRG image: ", err)
	//		return
	//	}
	//
	//	if err = ImgFromBin("chr.png", chrData); err != nil {
	//		fmt.Println("Unable to write CHR image: ", err)
	//		return
	//	}
	//
	//	if err = ImgFromBin("nes.png", nesraw); err != nil {
	//		fmt.Println("Unable to write NES image: ", err)
	//		return
	//	}
	//
	//	if err = ImgFromBin("barbie.prg.code.png", prgFilteredCode); err != nil {
	//		fmt.Println("Unable to write prg code image: ", err)
	//		return
	//	}
	//
	//	if err = ImgFromBin("barbie.prg.data.png", prgFilteredData); err != nil {
	//		fmt.Println("Unable to write prg data image: ", err)
	//		return
	//	}
	//
	//	fmt.Println("ok")
}

func ImgFromBin(filename string, data []byte) error {
	width := 256
	height := int(math.Ceil(float64(len(data)) / float64(width)))

	fmt.Printf("width: %d\nheight: %d\n", width, height)

	img := image.NewGray(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{width, height}})

	for i := 0; i < len(data); i++ {
		x := int(math.Mod(float64(i), float64(width)))
		y := int(math.Floor(float64(i) / float64(width)))

		img.Set(x, y, color.Gray{Y: uint8(data[i]) + 255})
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create %q: %s", filename, err)
	}

	if err = png.Encode(file, img); err != nil {
		return fmt.Errorf("Unable to write %q: %s", filename, err)
	}

	return nil
}
