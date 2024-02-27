package unif

import (
	"io"
	"bytes"
	"strings"
	"strconv"
	"bufio"
	"fmt"
	"encoding/binary"
)

type Rom struct {
	Version int
	Name string
	Mapper string

	PrgData []*ChunkData
	ChrData []*ChunkData
	Chunks []string

	Writer string // dumping software
	Read string // comments
	DumpInfo []byte
	TvStandard byte
	Controllers byte
	Battery bool
	ChrRam bool // ignored by emulators
	Mirroring byte
}

func (r *Rom) RomType() string {
	return "UNIF"
}

// PRG or CHR
type ChunkData struct {
	Id int
	Data []byte
	Crc []byte
}

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

func ReadUnif(r io.Reader) (*Rom, error) {
	reader := bufio.NewReader(r)
	magic := make([]byte, 4)

	_, err := reader.Read(magic)
	if err != nil {
		return nil, fmt.Errorf("Unable to read magic: %w", err)
	}

	if !bytes.Equal(magic, []byte("UNIF")) {
		return nil, fmt.Errorf("Not a UNIF rom")
	}

	rom := &Rom{}

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
			rom.Mapper = string(rawVal)

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
					Id: id,
					Data: rawVal,
					Crc: nil,
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
					Id: id,
					Data: nil,
					Crc: rawVal,
				})
			}

		case "NAME":
			rom.Name = string(rawVal)

		case "WRTR":
			rom.Writer = string(rawVal)

		case "READ":
			rom.Read = string(rawVal)

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
