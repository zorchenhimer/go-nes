package studybox

import (
	"fmt"
	//"strings"
)

type packetHeader struct {
	PageNumber uint8
	Checksum   uint8

	address int
}

func newPacketHeader(pageNumber uint8) *packetHeader {
	ph := &packetHeader{PageNumber: pageNumber}
	ph.Checksum = calcChecksum(ph.RawBytes()[0:7])
	return ph
}

func (ph *packetHeader) RawBytes() []byte {
	return []byte{0xC5, 0x01, 0x01, 0x01, 0x01,
		byte(ph.PageNumber), byte(ph.PageNumber), ph.Checksum}
}

func (ph *packetHeader) Asm() string {
	return fmt.Sprintf("header %d", ph.PageNumber)
}

func (ph *packetHeader) Address() int {
	return ph.address
}

type packetDelay struct {
	Length int

	address int
}

func newPacketDelay(length int) *packetDelay {
	return &packetDelay{Length: length}
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

func (p *packetDelay) Address() int {
	return p.address
}

type packetWorkRamLoad struct {
	bankId          uint8
	loadAddressHigh uint8
	checksum        uint8

	address int
}

func newPacketWorkRamLoad(bank, addressHigh uint8) *packetWorkRamLoad {
	p := &packetWorkRamLoad{bankId: bank, loadAddressHigh: addressHigh}
	p.checksum = calcChecksum(p.RawBytes()[0:5])
	return p
}

func (p *packetWorkRamLoad) Asm() string {
	return fmt.Sprintf("work_ram_load $%02X $%02X", p.bankId, p.loadAddressHigh)
}

func (p *packetWorkRamLoad) RawBytes() []byte {
	return []byte{0xC5, 0x02, 0x02, p.bankId, p.loadAddressHigh, p.checksum}
}

func (p *packetWorkRamLoad) Address() int {
	return p.address
}

type packetBulkData struct {
	checksum uint8
	Data     []byte

	address int
}

// Returns a list of packets
func newBulkDataPackets(raw []byte) []Packet {
	packets := []Packet{}
	for i := 0; i < len(raw); i += 128 {
		l := 128
		// TODO: veryfy this is actually correct
		if len(raw) < i+128 {
			l = len(raw) - i
		}
		p := &packetBulkData{Data: raw[i : i+l]}
		p.checksum = calcChecksum(p.RawBytes()[0 : len(p.Data)-1])
		packets = append(packets, p)
	}

	return packets
}

func (p *packetBulkData) Asm() string {
	// commented out code prints the full data
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

func (p *packetBulkData) Address() int {
	return p.address
}

type packetMarkDataStart struct {
	ArgA uint8
	ArgB uint8
	Type uint8

	address  int
	checksum uint8
}

func newPacketMarkDataStart(dataType packetType, a, b uint8) *packetMarkDataStart {
	p := &packetMarkDataStart{
		Type: uint8(dataType),
		ArgA: a,
		ArgB: b,
	}
	p.checksum = calcChecksum(p.RawBytes()[0:3])
	return p
}

func (p *packetMarkDataStart) dataType() string {
	tstr := "unknown"
	switch p.Type {
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

func (p *packetMarkDataStart) RawBytes() []byte {
	return []byte{0xC5, uint8(p.Type), uint8(p.Type), p.ArgA, p.ArgB, p.checksum}
}

func (p *packetMarkDataStart) Address() int {
	return p.address
}

type packetMarkDataEnd struct {
	//Arg   uint8
	Reset bool
	Type  uint8

	address  int
	checksum uint8
}

type packetType uint8

const (
	packet_Script packetType = 2
	packet_Nametable
	packet_Pattern
	packet_Delay
)

func newPacketMarkDataEnd(datatype packetType, reset bool) *packetMarkDataEnd {
	p := &packetMarkDataEnd{
		Reset: reset,
		Type:  uint8(datatype),
	}
	p.checksum = calcChecksum(p.RawBytes()[0:3])
	return p
}

func (p *packetMarkDataEnd) RawBytes() []byte {
	arg := uint8(p.Type)
	if p.Reset {
		arg |= 0xF0
	}
	return []byte{0xC5, 0x00, uint8(p.Type), p.checksum}
}

func (p *packetMarkDataEnd) Asm() string {
	var tstr string
	switch p.Type & 0x0F {
	case 2:
		tstr = "script"
	case 3:
		tstr = "nametable"
	case 4:
		tstr = "pattern"
	case 5:
		tstr = "delay"
	default:
		tstr = fmt.Sprintf("unknown $%02X", p.Type)
	}

	if p.Reset {
		tstr += " reset_state"
	}
	return fmt.Sprintf("mark_datatype_end %s", tstr)
}

func (p *packetMarkDataEnd) Address() int {
	return p.address
}

type packetPadding struct {
	Length  int
	address int
	raw     []byte
}

func newPacketPadding(length int) *packetPadding {
	return &packetPadding{Length: length}
}

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

func (p *packetPadding) Address() int {
	return p.address
}
