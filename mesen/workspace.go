package mesen

import (
	"encoding/json"
	"fmt"
	"io"
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
	// TODO: detect this, instead of assuming it's there.
	garbo := make([]byte, 3)
	_, err := reader.Read(garbo)
	if err != nil {
		return nil, fmt.Errorf("Error reading past BOM: %w", err)
	}

	d := json.NewDecoder(reader)
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
