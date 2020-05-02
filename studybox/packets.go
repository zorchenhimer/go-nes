package studybox

import (
	"fmt"
	"strings"
)

type PacketHeader struct {
	PageNumber uint8
	Checksum   uint8

	meta PacketMeta
}

func (ph *PacketHeader) RawBytes() []byte {
	return []byte{0xC5, 0x01, 0x01, 0x01, 0x01, byte(ph.PageNumber), byte(ph.PageNumber), 0xC5}
}

func (ph *PacketHeader) Asm() string {
	return fmt.Sprintf("header %d", ph.PageNumber)
}

func (ph *PacketHeader) Meta() PacketMeta {
	return ph.meta
}

type PacketDelay struct {
	Length int

	meta PacketMeta
}

func (pd *PacketDelay) RawBytes() []byte {
	payload := make([]byte, pd.Length)
	for i := 0; i < pd.Length; i++ {
		payload[i] = 0xAA
	}

	return append([]byte{0xC5, 0x05, 0x05}, payload...)
}

func (pd *PacketDelay) Asm() string {
	return fmt.Sprintf("delay %d", pd.Length)
}

func (pd *PacketDelay) Meta() PacketMeta {
	return pd.meta
}

type PacketUnknown struct {
	rawData []byte
	notes   string

	meta PacketMeta
}

func (pu *PacketUnknown) Asm() string {
	data := []string{}
	for i := 2; i < pu.meta.Length-1; i++ {
		data = append(data, fmt.Sprintf("%02X", pu.rawData[i]))
	}
	notes := ""
	if pu.notes != "" {
		notes = "; " + pu.notes
	}
	return fmt.Sprintf("unknown_state%d_type%d %s %s", pu.meta.State, pu.meta.Type, strings.Join(data, " "), notes)
}

func (pu *PacketUnknown) Meta() PacketMeta {
	return pu.meta
}

func (pu *PacketUnknown) RawBytes() []byte {
	return pu.rawData
}
