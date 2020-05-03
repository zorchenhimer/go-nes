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
	rawData  []byte
	notes    string
	checksum uint8

	meta PacketMeta
}

func (pu *PacketUnknown) Asm() string {
	data := []string{}
	i := 2
	if pu.meta.State == 2 {
		i = 3
	}
	for ; i < pu.meta.Length-1; i++ {
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

type PacketWorkRamLoad struct {
	bankId          uint8
	loadAddressHigh uint8
	checksum        uint8

	meta PacketMeta
}

func (p *PacketWorkRamLoad) Meta() PacketMeta {
	return p.meta
}

func (p *PacketWorkRamLoad) Asm() string {
	return fmt.Sprintf("work_ram_load $%02X $%02X", p.bankId, p.loadAddressHigh)
}

func (p *PacketWorkRamLoad) RawBytes() []byte {
	return []byte{0xC5, 0x02, 0x02, p.bankId, p.loadAddressHigh, p.checksum}
}

type PacketBulkData struct {
	checksum uint8
	Data     []byte

	meta PacketMeta
}

func (p *PacketBulkData) Asm() string {
	//data := []string{}
	//for _, b := range p.Data {
	//	data = append(data, fmt.Sprintf("$%02X", b))
	//}
	//return fmt.Sprintf("data %s", strings.Join(data, ", "))
	return fmt.Sprintf("data $%02X, [...], $%02X ; Length %d", p.Data[0], p.Data[len(p.Data)-1], len(p.Data))
}

func (p *PacketBulkData) RawBytes() []byte {
	data := []byte{0xC5, uint8(len(p.Data))}
	data = append(data, p.Data...)
	data = append(data, p.checksum)
	return data
}

func (p *PacketBulkData) Meta() PacketMeta { return p.meta }

type PacketMarkDataStart struct {
	ArgA uint8
	ArgB uint8

	meta     PacketMeta
	checksum uint8
}

func (p *PacketMarkDataStart) Asm() string {
	tstr := "unknown"
	switch p.meta.Type {
	case 2:
		tstr = "script"
	case 3:
		tstr = "nametable"
	case 4:
		tstr = "pattern"
	}
	return fmt.Sprintf("mark_datatype_start %s $%02X $%02X", tstr, p.ArgA, p.ArgB)
}

func (p *PacketMarkDataStart) Meta() PacketMeta { return p.meta }

func (p *PacketMarkDataStart) RawBytes() []byte {
	return []byte{0xC5, uint8(p.meta.Type), uint8(p.meta.Type), p.ArgA, p.ArgB, p.checksum}
}

type PacketMarkDataEnd struct {
	Arg uint8

	meta     PacketMeta
	checksum uint8
}

func (p *PacketMarkDataEnd) Meta() PacketMeta { return p.meta }

func (p *PacketMarkDataEnd) RawBytes() []byte {
	return []byte{0xC5, uint8(p.meta.Type), p.Arg, p.checksum}
}

func (p *PacketMarkDataEnd) Asm() string {
	var tstr string
	switch p.Arg & 0x0F {
	case 2:
		tstr = "script"
	case 3:
		tstr = "nametable"
	case 4:
		tstr = "pattern"
	case 5:
		tstr = "delay"
	default:
		tstr = fmt.Sprintf("unknown $%02X", p.Arg)
	}

	if p.Arg&0xF0 == 0xF0 {
		tstr += " reset_state"
	}
	return fmt.Sprintf("mark_datatype_end %s", tstr)
}
