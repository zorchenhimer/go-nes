package studybox

import (
	"fmt"
	"strings"
)

type packetHeader struct {
	PageNumber uint8
	Checksum   uint8

	meta PacketMeta
}

func (ph *packetHeader) RawBytes() []byte {
	return []byte{0xC5, 0x01, 0x01, 0x01, 0x01, byte(ph.PageNumber), byte(ph.PageNumber), 0xC5}
}

func (ph *packetHeader) Asm() string {
	return fmt.Sprintf("header %d", ph.PageNumber)
}

func (ph *packetHeader) Meta() PacketMeta {
	return ph.meta
}

type packetDelay struct {
	Length int

	meta PacketMeta
}

func (pd *packetDelay) RawBytes() []byte {
	payload := make([]byte, pd.Length)
	for i := 0; i < pd.Length; i++ {
		payload[i] = 0xAA
	}

	return append([]byte{0xC5, 0x05, 0x05}, payload...)
}

func (pd *packetDelay) Asm() string {
	return fmt.Sprintf("delay %d", pd.Length)
}

func (pd *packetDelay) Meta() PacketMeta {
	return pd.meta
}

type packetUnknown struct {
	rawData  []byte
	notes    string
	checksum uint8

	meta PacketMeta
}

func (pu *packetUnknown) Asm() string {
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

func (pu *packetUnknown) Meta() PacketMeta {
	return pu.meta
}

func (pu *packetUnknown) RawBytes() []byte {
	return pu.rawData
}

type packetWorkRamLoad struct {
	bankId          uint8
	loadAddressHigh uint8
	checksum        uint8

	meta PacketMeta
}

func (p *packetWorkRamLoad) Meta() PacketMeta {
	return p.meta
}

func (p *packetWorkRamLoad) Asm() string {
	return fmt.Sprintf("work_ram_load $%02X $%02X", p.bankId, p.loadAddressHigh)
}

func (p *packetWorkRamLoad) RawBytes() []byte {
	return []byte{0xC5, 0x02, 0x02, p.bankId, p.loadAddressHigh, p.checksum}
}

type packetBulkData struct {
	checksum uint8
	Data     []byte

	meta PacketMeta
}

func (p *packetBulkData) Asm() string {
	//data := []string{}
	//for _, b := range p.Data {
	//	data = append(data, fmt.Sprintf("$%02X", b))
	//}
	//return fmt.Sprintf("data %s", strings.Join(data, ", "))
	return fmt.Sprintf("data $%02X, [...], $%02X ; Length %d", p.Data[0], p.Data[len(p.Data)-1], len(p.Data))
}

func (p *packetBulkData) RawBytes() []byte {
	data := []byte{0xC5, uint8(len(p.Data))}
	data = append(data, p.Data...)
	data = append(data, p.checksum)
	return data
}

func (p *packetBulkData) Meta() PacketMeta { return p.meta }

type packetMarkDataStart struct {
	ArgA uint8
	ArgB uint8

	meta     PacketMeta
	checksum uint8
}

func (p *packetMarkDataStart) dataType() string {
	tstr := "unknown"
	switch p.meta.Type {
	case 2:
		tstr = "script"
	case 3:
		tstr = "nametable"
	case 4:
		tstr = "pattern"
	}
	return tstr
}

func (p *packetMarkDataStart) Asm() string {
	return fmt.Sprintf("mark_datatype_start %s $%02X $%02X", p.dataType(), p.ArgA, p.ArgB)
}

func (p *packetMarkDataStart) Meta() PacketMeta { return p.meta }

func (p *packetMarkDataStart) RawBytes() []byte {
	return []byte{0xC5, uint8(p.meta.Type), uint8(p.meta.Type), p.ArgA, p.ArgB, p.checksum}
}

type packetMarkDataEnd struct {
	Arg   uint8
	Reset bool

	meta     PacketMeta
	checksum uint8
}

func (p *packetMarkDataEnd) Meta() PacketMeta { return p.meta }

func (p *packetMarkDataEnd) RawBytes() []byte {
	return []byte{0xC5, uint8(p.meta.Type), p.Arg, p.checksum}
}

func (p *packetMarkDataEnd) Asm() string {
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

	if p.Reset {
		tstr += " reset_state"
	}
	return fmt.Sprintf("mark_datatype_end %s", tstr)
}

type packetPadding struct {
	Length int
	meta   PacketMeta
}

func (p *packetPadding) Meta() PacketMeta { return p.meta }

func (p *packetPadding) Asm() string {
	return fmt.Sprintf("page_padding %d", p.Length)
}

func (p *packetPadding) RawBytes() []byte {
	b := []byte{}
	for i := 0; i < p.Length; i++ {
		b = append(b, 0xAA)
	}
	return b
}
