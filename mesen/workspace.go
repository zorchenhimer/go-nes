package mesen

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type MemoryType string

const (
	NesChrRam             MemoryType = "NesChrRam"
	NesChrRom             MemoryType = "NesChrRom"
	NesInternalRam        MemoryType = "NesInternalRam"
	NesMemory             MemoryType = "NesMemory"
	NesNametableRam       MemoryType = "NesNametableRam"
	NesPaletteRam         MemoryType = "NesPaletteRam"
	NesPrgRom             MemoryType = "NesPrgRom"
	NesSaveRam            MemoryType = "NesSaveRam"
	NesSecondarySpriteRam MemoryType = "NesSecondarySpriteRam"
	NesSpriteRam          MemoryType = "NesSpriteRam"
	NesWorkRam            MemoryType = "NesWorkRam"
)

type Workspace struct {
	WatchEntries []string

	Labels []struct {
		Address    uint
		MemoryType string
		Label      string
		Comment    string
		Flags      string
		Length     int
	}

	Breakpoints []struct {
		BreakOnRead  bool
		BreakOnWrite bool
		Enabled      bool
		MarkEvent    bool
		MemoryType   string
		StartAddress uint
		EndAddress   uint
		CpuType      string
		AnyAddress   bool
		IsAssert     bool
		Condition    string
	}

	//TableMappings []struct{
	//} `json:"TblMappings"`
}

type jsonWorkspace struct {
	WorkspaceByCpu map[string]Workspace
}

func LoadWorkspace(reader io.Reader) (*Workspace, error) {
	// There's a UTF-8 BOM here because *reasons*.  Get rid of it,
	// otherwise the JSON decoder dies.
	buf := bufio.NewReader(reader)
	garbo, err := buf.Peek(3)
	if err != nil {
		return nil, fmt.Errorf("Error peeking for BOM: %w", err)
	}

	// UTF-8 BOM.  Do we need to check for other BOMs?
	if bytes.Equal(garbo, []byte{0xEF, 0xBB, 0xBF}) {
		_, err = buf.Discard(3)
		if err != nil {
			return nil, fmt.Errorf("Error discarding BOM: %w", err)
		}
	}

	d := json.NewDecoder(buf)
	ws := &jsonWorkspace{}
	err = d.Decode(ws)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", d.InputOffset(), err)
	}

	if nes, exists := ws.WorkspaceByCpu["Nes"]; exists {
		return &nes, nil
	}

	return nil, fmt.Errorf("Nes workspace not found")
}
