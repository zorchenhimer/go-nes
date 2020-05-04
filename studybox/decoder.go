package studybox

import (
	"bytes"
	"fmt"
	"strings"
	//"encoding/binary"
)

type DecodedData struct {
	Packets []Packet
}

type PacketMeta struct {
	Start  int
	Data   int
	Length int // length of whole packet
	State  int // packet state type. -1 is unknown
	Type   int // packet type ID. usually the second byte
}

func (dd *DecodedData) String() string {
	str := []string{}
	for _, p := range dd.Packets {
		str = append(str, fmt.Sprintf("%08X: %s", p.Meta().Start, p.Asm()))
	}
	return strings.Join(str, "\n")
}

type Packet interface {
	RawBytes() []byte
	Asm() string
	Meta() PacketMeta // offset in the file to the start of the data packet
}

type decodeFunction func(page *Page, startIdx int) (Packet, int, error)

// map of states.  each state is a map of types
var definedPackets = map[int]map[byte]decodeFunction{
	0: map[byte]decodeFunction{
		0x01: decodeHeader,
	},

	1: map[byte]decodeFunction{
		0x00: decodeMarkDataEnd,
	},

	2: map[byte]decodeFunction{
		0x02: decodeSetWorkRamLoad,
		0x03: decodeMarkDataStart,
		0x04: decodeMarkDataStart,
		0x05: decodeDelay,
	},
}

var state int // weeeeee
func DecodePage(page *Page) (*DecodedData, error) {
	var err error
	decoded := &DecodedData{
		Packets: []Packet{},
	}
	state = 0

	fmt.Printf("Decoding page %s\n", page)

	for idx := 0; idx < len(page.Data); {
		//fmt.Printf("decoding packet at file offset: %08X\n", idx+page.DataOffset)

		if page.Data[idx] != 0xC5 {
			return decoded, fmt.Errorf("Packet at offset %08X does not start with $C5: %02X", idx+page.DataOffset, page.Data[idx])
		}

		var packet Packet
		if state == 1 && page.Data[idx+1] != 0x00 {
			// bulk data
			packet, state, err = decodeBulkData(page, idx)
			if err != nil {
				return decoded, err
			}
		} else {
			df, ok := definedPackets[state][page.Data[idx+1]]
			if !ok {
				return decoded, fmt.Errorf("State %d packet with type %02X isn't implemented",
					state, page.Data[idx+1])
			}
			packet, state, err = df(page, idx)
			if err != nil {
				return decoded, err
			}
		}
		decoded.Packets = append(decoded.Packets, packet)
		idx += packet.Meta().Length
	}

	return decoded, nil
}

// Returns packet and next state
func decodeHeader(page *Page, idx int) (Packet, int, error) {
	if !bytes.Equal(page.Data[idx+1:idx+5], []byte{0x01, 0x01, 0x01, 0x01}) {
		return nil, 0, fmt.Errorf("Packet header at offset %08X has invalid payload: $%08X", idx+page.DataOffset, page.Data[idx+1:idx+5])
	}

	if page.Data[idx+5] != page.Data[idx+6] {
		return nil, 0, fmt.Errorf("Packet header at offset %08X has missmatched page numbers at offset %08X: %02X vs %02X",
			idx+page.DataOffset,
			idx+page.DataOffset+5,
			page.Data[idx+5],
			page.Data[idx+6],
		)
	}

	ph := &PacketHeader{
		PageNumber: uint8(page.Data[idx+6]),
		Checksum:   page.Data[idx+8],
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			State:  state,
			Type:   int(page.Data[idx+1]),
			Length: 8,
		},
	}
	ph.meta.Data = ph.meta.Start + idx + 5

	checksum := calcChecksum(page.Data[idx : idx+ph.meta.Length-1])
	if checksum != ph.Checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for header packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, ph.Checksum)
	}

	return ph, 2, nil
}

func decodeDelay(page *Page, idx int) (Packet, int, error) {
	if page.Data[idx+1] != page.Data[idx+2] {
		return nil, 0, fmt.Errorf("State 2 packet at offset %08X has missmatched type [%08X]: %d vs %d",
			idx+page.DataOffset, idx+1+page.DataOffset, page.Data[idx+1], page.Data[idx+2])
	}

	count := 0
	var i int
	for i = idx + 3; i < len(page.Data) && page.Data[i] != 0x00 && page.Data[i] != 0xC5; i++ {
		count++
	}
	if count%2 != 0 {
		fmt.Printf("0xAA delay packet at offset %08X has odd number of 0xAA's", idx+page.FileOffset)
	}
	pd := &PacketDelay{
		Length: count,
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			State:  state,
			Type:   int(page.Data[idx+1]),
			Length: count + 3,
		},
	}
	pd.meta.Data = pd.meta.Start + 3

	checksum := calcChecksum(page.Data[idx : pd.meta.Length+idx])
	if checksum != 0xC5 {
		return nil, 0, fmt.Errorf("Invalid checksum for delay packet starting at offset %08X. Got %02X, expected %02X", checksum, 0xC5)
	}

	idx += count + 3
	return pd, 1, nil
}

func decodeUnknownS1T0(page *Page, idx int) (Packet, int, error) {
	// unknown packet
	unk := &PacketUnknown{
		rawData: page.Data[idx : idx+4],
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 2,
			Length: 4,
			State:  state,
			Type:   0,
		},
	}

	var checksum uint8
	for i := idx; i < unk.meta.Length+idx-1; i++ {
		checksum ^= page.Data[i]
	}

	switch page.Data[idx+2] {
	case 0x02:
		unk.notes = "finished loading (script?) data?"
	case 0x03:
		unk.notes = "finished loading (nametable?) data?"
	case 0x04:
		unk.notes = "finished loading (pattern?) data?"
	case 0x05:
		unk.notes = "prepare for next address?"
	case 0xF5:
		unk.notes = "stop at end of page?"
	default:
		unk.notes = "???"
	}

	state := 2
	if page.Data[idx+2]&0xF0 == 0xF0 {
		// this changes to state 3, not zero!
		state = 0
	}

	return unk, state, nil
}

func decodeMarkDataStart(page *Page, idx int) (Packet, int, error) {
	packet := &PacketMarkDataStart{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 3,
			Length: 6,
			State:  state,
			Type:   int(page.Data[idx+1]),
		},
		checksum: page.Data[idx+5],
		ArgA:     page.Data[idx+3],
		ArgB:     page.Data[idx+4],
	}

	checksum := calcChecksum(page.Data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for UnknownS2T3 packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}
	return packet, 1, nil
}

func decodeMarkDataEnd(page *Page, idx int) (Packet, int, error) {
	packet := &PacketMarkDataEnd{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 2,
			Length: 4,
			State:  state,
			Type:   int(page.Data[idx+1]),
		},
		checksum: page.Data[idx+3],
		Arg:      page.Data[idx+2],
	}

	checksum := calcChecksum(page.Data[idx : idx+3])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for UnknownS2T3 packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	newstate := 2
	//if page.Data[idx+2]&0xF0 == 0xF0 {
	//	// this changes to state 3, not zero!
	//	newstate = 0
	//}

	return packet, newstate, nil

}

func decodeUnknownS2(page *Page, idx int) (Packet, int, error) {
	packet := &PacketUnknown{
		rawData: page.Data[idx : idx+6],
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 3,
			Length: 6,
			State:  state,
			Type:   int(page.Data[idx+1]),
		},
		checksum: page.Data[idx+5],
	}

	switch packet.meta.Type {
	case 3:
		packet.notes = "possibly nametable data?"
	case 4:
		packet.notes = "Pattern table data"
	default:
		packet.notes = "???"
	}

	checksum := calcChecksum(page.Data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for UnknownS2T3 packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	return packet, 1, nil
}

// C5 02 02 nn mm zz
// Map 8k ram bank nn to $6000-$7FFF; set load address to $mm00; zz = checksum
func decodeSetWorkRamLoad(page *Page, idx int) (Packet, int, error) {
	if page.Data[idx+1] != page.Data[idx+2] {
		return nil, 0, fmt.Errorf("State 1 packet at offset %08X has missmatched type [%08X]: %d vs %d",
			idx+page.DataOffset, idx+1+page.DataOffset, page.Data[idx+1], page.Data[idx+2])
	}

	packet := &PacketWorkRamLoad{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Length: 6,
			State:  state,
			Type:   2,
		},
		bankId:          page.Data[idx+3],
		loadAddressHigh: page.Data[idx+4],
		checksum:        page.Data[idx+5],
	}
	packet.meta.Data = packet.meta.Start + 3

	checksum := calcChecksum(page.Data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for SetWorkRamLoad packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	return packet, 1, nil
}

func decodeBulkData(page *Page, idx int) (Packet, int, error) {
	if page.Data[idx+1] == 0 {
		return nil, 0, fmt.Errorf("Bulk data packet has a length of zero at offset %08X",
			page.DataOffset+idx)
	}

	packet := &PacketBulkData{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Length: int(page.Data[idx+1]) + 3,
			State:  state,
			Type:   1,
		},
		//data:     page.Data[idx+2 : idx+2+int(page.Data[idx+1])],
		//checksum: page.Data[page.Data[idx+1]+3],
	}

	datalen := int(page.Data[idx+1])
	packet.Data = page.Data[idx+2 : idx+2+datalen]
	packet.checksum = page.Data[idx+len(packet.Data)+2]

	checksum := calcChecksum(page.Data[idx : idx+packet.meta.Length-1])
	if checksum != packet.checksum {
		data := []string{}
		for _, b := range packet.Data {
			data = append(data, fmt.Sprintf("$%02X", b))
		}
		fmt.Printf("checksum data: %s\n", strings.Join(data, " "))
		fmt.Printf("checksum address: %08X\n", page.DataOffset+idx+len(packet.Data)+2)
		return nil, 0, fmt.Errorf("Invalid checksum for BulkData packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	return packet, 1, nil
}

func calcChecksum(data []byte) uint8 {
	var sum uint8
	for i := 0; i < len(data); i++ {
		sum ^= data[i]
	}
	return sum
}
