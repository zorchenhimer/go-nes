package rom

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Remap struct {
	Pattern   string
	Mapper    uint
	Submapper uint8
	Mirroring MirrorType
	PrgRam    uint
	ChrRam    uint
}

// Board name -> mapper, submapper, etc
var UnifRemap = []Remap{
	Remap{Pattern: "*-8237", Mapper: 215},
	Remap{Pattern: "*-8237a", Mapper: 215, Submapper: 1},
	Remap{Pattern: "*-cnrom", Mapper: 3, Submapper: 2},
	Remap{Pattern: "*-h2288", Mapper: 123},
	Remap{Pattern: "*-kof97", Mapper: 263},
	Remap{Pattern: "*-sa-0036", Mapper: 149},
	Remap{Pattern: "*-sa-0037", Mapper: 148},
	Remap{Pattern: "*-sa-016-1m", Mapper: 79},
	Remap{Pattern: "*-sa-72007", Mapper: 145},
	Remap{Pattern: "*-sa-72008", Mapper: 133},
	Remap{Pattern: "*-sa-nrom", Mapper: 143},
	Remap{Pattern: "*-sachen-74ls374n", Mapper: 150},
	Remap{Pattern: "*-sachen-8259a", Mapper: 141},
	Remap{Pattern: "*-sachen-8259b", Mapper: 138},
	Remap{Pattern: "*-sachen-8259c", Mapper: 139},
	Remap{Pattern: "*-sachen-8259d", Mapper: 137},
	Remap{Pattern: "*-shero", Mapper: 262, Mirroring: M_IGNORE},
	Remap{Pattern: "*-tc-u01-1.5m", Mapper: 147},
	Remap{Pattern: "*-tlrom", Mapper: 4},
	Remap{Pattern: "dreamtech01", Mapper: 521, ChrRam: 8},
	Remap{Pattern: "nes-nrom-128", Mapper: 0},
	Remap{Pattern: "nes-nrom-256", Mapper: 0},
}

type UnifRom struct {
	Version int
	Name    string
	Mapper  string

	PrgData []*ChunkData
	ChrData []*ChunkData
	Chunks  []string

	Writer      string // dumping software
	Read        string // comments
	DumpInfo    []byte
	TvStandard  byte
	Controllers byte
	Battery     bool
	ChrRam      bool // ignored by emulators
	Mirroring   byte
}

func (r *UnifRom) Ines() []byte {
	prg := []byte{}
	prgSlice := ChunkSlice(r.PrgData)
	sort.Sort(prgSlice)

	for _, chunk := range prgSlice {
		prg = append(prg, chunk.Data...)
	}

	chr := []byte{}
	chrSlice := ChunkSlice(r.ChrData)
	sort.Sort(chrSlice)

	for _, chunk := range chrSlice {
		chr = append(chr, chunk.Data...)
	}

	header := Header{
		PrgSize:          uint(len(prg)),
		ChrSize:          uint(len(chr)),
		PersistentMemory: r.Battery,
	}

	found := false
	fmt.Printf("Looking for %q\n", strings.ToLower(r.Mapper))
	for _, remap := range UnifRemap {
		match, err := filepath.Match(remap.Pattern, strings.ToLower(r.Mapper))
		if err != nil {
			panic(err)
		}

		if !match {
			continue
		}

		found = true
		header.Mapper = remap.Mapper
		header.Nes2Mapper = uint16(remap.Mapper)
		header.Mirroring = remap.Mirroring

		if remap.PrgRam != 0 {
			header.PrgRamSize = unshift(remap.PrgRam)
		}

		if remap.ChrRam != 0 {
			header.ChrRamSize = unshift(remap.ChrRam)
		}

		break
	}

	if !found {
		panic(r.Mapper + " not implemented")
	}

	raw := header.Bytes()
	raw = append(raw, prg...)
	raw = append(raw, chr...)

	return raw
}

func unshift(val uint) uint {
	count := uint(0)
	for val > 64 {
		val >>= 1
		count++
	}
	return count
}

func (r *UnifRom) Debug(w io.Writer) error {
	_, err := fmt.Fprintf(w, `%s
	Mapper: %s
	Version: %d
	Writer: %s
	Read: %s
	DumpInfo: %v
	TvStandard: %X
	Controllers: %08b
	Battery: %t
	ChrRam: %t
	Mirroring: %X

`,
		strings.Trim(r.Name, "\x00"),
		strings.Trim(r.Mapper, "\x00"),
		r.Version,
		r.Writer,
		r.Read,
		r.DumpInfo,
		r.TvStandard,
		r.Controllers,
		r.Battery,
		r.ChrRam,
		r.Mirroring,
	)

	return err
}

func (r *UnifRom) RomType() RomType {
	return UNIF
}

func (r *UnifRom) PrgRom() []byte {
	return []byte{}
}

func (r *UnifRom) ChrRom() []byte {
	return []byte{}
}

// PRG or CHR
type ChunkData struct {
	Id   int
	Data []byte
	Crc  []byte
}

type ChunkSlice []*ChunkData

func (s ChunkSlice) Len() int           { return len(s) }
func (s ChunkSlice) Less(i, j int) bool { return s[i].Id < s[j].Id }
func (s ChunkSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

var chunkLengths = map[string]int{
	"DINF": 204,
	"TVCI": 1,
	"CTRL": 1,
	"BATR": 1,
	"VROR": 1,
	"MIRR": 1,

	"PCK0": 4,
	"PCK1": 4,
	"PCK2": 4,
	"PCK3": 4,
	"PCK4": 4,
	"PCK5": 4,
	"PCK6": 4,
	"PCK7": 4,
	"PCK8": 4,
	"PCK9": 4,
	"PCKA": 4,
	"PCKB": 4,
	"PCKC": 4,
	"PCKD": 4,
	"PCKE": 4,
	"PCKF": 4,

	"CCK0": 4,
	"CCK1": 4,
	"CCK2": 4,
	"CCK3": 4,
	"CCK4": 4,
	"CCK5": 4,
	"CCK6": 4,
	"CCK7": 4,
	"CCK8": 4,
	"CCK9": 4,
	"CCKA": 4,
	"CCKB": 4,
	"CCKC": 4,
	"CCKD": 4,
	"CCKE": 4,
	"CCKF": 4,
}

func ReadUnif(r io.Reader) (*UnifRom, error) {
	reader := bufio.NewReader(r)
	magic := make([]byte, 4)

	_, err := reader.Read(magic)
	if err != nil {
		return nil, fmt.Errorf("Unable to read magic: %w", err)
	}

	if !bytes.Equal(magic, []byte("UNIF")) {
		return nil, fmt.Errorf("Not a UNIF rom")
	}

	rom := &UnifRom{}

	version := make([]byte, 4)
	_, err = reader.Read(version)
	if err != nil {
		return nil, fmt.Errorf("Error reading version: %w", err)
	}

	vbuf := bytes.NewReader(version)
	var version32 int32
	err = binary.Read(vbuf, binary.LittleEndian, &version32)
	if err != nil {
		return nil, fmt.Errorf("Invalid version: %w", err)
	}
	rom.Version = int(version32)

	_, err = reader.Discard(24)
	if err != nil {
		return nil, err
	}

	// chunks
	for {
		rawChunkType := make([]byte, 4)
		_, err = reader.Read(rawChunkType)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		chunkType := string(rawChunkType)
		rom.Chunks = append(rom.Chunks, chunkType)

		rawChunkLength := make([]byte, 4)
		_, err = reader.Read(rawChunkLength)
		if err != nil {
			return nil, err
		}

		var chunkLen32 int32
		lbuf := bytes.NewReader(rawChunkLength)
		err = binary.Read(lbuf, binary.LittleEndian, &chunkLen32)
		if err != nil {
			return nil, err
		}
		chunkLen := int(chunkLen32)

		//fmt.Printf("%q %d\n", rawChunkType, chunkLen)

		if l, ok := chunkLengths[chunkType]; ok && l != chunkLen {
			return nil, fmt.Errorf("Chunk %s has invalid length %d", chunkType, chunkLen)
		}

		if chunkLen <= 0 {
			return nil, fmt.Errorf("Chunk length for %s is invalid: %v", chunkType, rawChunkLength)
		}
		rawVal := make([]byte, chunkLen)
		rawLen, err := io.ReadFull(reader, rawVal)
		if err != nil {
			return nil, fmt.Errorf("Error reading %s chunk: %w", string(rawChunkType), err)
		}

		if rawLen != chunkLen {
			return nil, fmt.Errorf("Invalid read for rawVal: %d vs %d", rawLen, chunkLen)
		}

		switch chunkType {
		case "MAPR":
			rom.Mapper = trimnul(string(rawVal))

		case "PRG0", "PRG1", "PRG2", "PRG3", "PRG4", "PRG5", "PRG6", "PRG7", "PRG8",
			"PRG9", "PRGA", "PRGB", "PRGC", "PRGD", "PRGE", "PRGF", "CHR0", "CHR1",
			"CHR2", "CHR3", "CHR4", "CHR5", "CHR6", "CHR7", "CHR8", "CHR9", "CHRA",
			"CHRB", "CHRC", "CHRD", "CHRE", "CHRF":

			id64, err := strconv.ParseInt(string(rawChunkType[3]), 16, 8)
			if err != nil {
				return nil, fmt.Errorf("Invalid PRG chunk type: %s", chunkType)
			}
			id := int(id64)

			var chunkData *[]*ChunkData
			if strings.HasPrefix(chunkType, "PRG") {
				chunkData = &rom.PrgData
			} else {
				chunkData = &rom.ChrData
			}

			found := false
			for _, c := range *chunkData {
				if c.Id != id {
					continue
				}

				found = true
				if c.Data != nil || len(c.Data) != 0 {
					return nil, fmt.Errorf("Duplicate chunk data for chunk %s", string(rawChunkType))
				}

				c.Data = rawVal
				break
			}

			if !found {
				*chunkData = append(*chunkData, &ChunkData{
					Id:   id,
					Data: rawVal,
					Crc:  nil,
				})
			}

		case "PCK0", "PCK1", "PCK2", "PCK3", "PCK4", "PCK5", "PCK6", "PCK7", "PCK8",
			"PCK9", "PCKA", "PCKB", "PCKC", "PCKD", "PCKE", "PCKF", "CCK0", "CCK1",
			"CCK2", "CCK3", "CCK4", "CCK5", "CCK6", "CCK7", "CCK8", "CCK9", "CCKA",
			"CCKB", "CCKC", "CCKD", "CCKE", "CCKF":

			var chunkData *[]*ChunkData
			if strings.HasPrefix(chunkType, "PCK") {
				chunkData = &rom.PrgData
			} else {
				chunkData = &rom.ChrData
			}

			id64, err := strconv.ParseInt(string(rawChunkType[3]), 16, 8)
			if err != nil {
				return nil, fmt.Errorf("Invalid PRG chunk type: %s", string(rawChunkType))
			}
			id := int(id64)

			found := false
			for _, c := range *chunkData {
				if c.Id != id {
					continue
				}
				found = true

				if c.Crc != nil || len(c.Crc) != 0 {
					return nil, fmt.Errorf("Duplicate CRC for chunk %s", string(rawChunkType))
				}
				c.Crc = rawVal
				break
			}

			if !found {
				*chunkData = append(*chunkData, &ChunkData{
					Id:   id,
					Data: nil,
					Crc:  rawVal,
				})
			}

		case "NAME":
			rom.Name = trimnul(string(rawVal))

		case "WRTR":
			rom.Writer = trimnul(string(rawVal))

		case "READ":
			rom.Read = trimnul(string(rawVal))

		case "DINF":
			rom.DumpInfo = rawVal

		case "TVCI":
			rom.TvStandard = rawVal[0]

		case "CTRL":
			rom.Controllers = rawVal[0]

		case "BATR":
			rom.Battery = (rawVal[0] != 0x00)

		case "VROR":
			rom.ChrRam = (rawVal[0] != 0x00)

		case "MIRR":
			rom.Mirroring = rawVal[0]
		}
	}

	return rom, nil
}

func trimnul(str string) string {
	return strings.Trim(str, "\x00")
}
