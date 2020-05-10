package studybox

import (
	"bytes"
	"fmt"
	"strings"
	//"encoding/binary"
)

type decodeFunction func(page *Page, data []byte, startIdx int) (Packet, int, error)

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

func (page *Page) decode(data []byte) error {
	var err error
	page.Packets = []Packet{}
	page.state = 0

	for idx := 0; idx < len(data); {
		if data[idx] != 0xC5 {
			// Padding after the last valid packet.
			dataLeft := len(data) - idx + 1
			page.Packets = append(page.Packets, &packetPadding{Length: dataLeft})
			return nil
		}

		var packet Packet
		if page.state == 1 && data[idx+1] != 0x00 {
			// bulk data
			packet, page.state, err = decodeBulkData(page, data, idx)
			if err != nil {
				return err
			}
		} else {
			df, ok := definedPackets[page.state][data[idx+1]]
			if !ok {
				return fmt.Errorf("State %d packet with type %02X isn't implemented",
					page.state, data[idx+1])
			}
			packet, page.state, err = df(page, data, idx)
			if err != nil {
				return err
			}
		}
		page.Packets = append(page.Packets, packet)
		idx += packet.Meta().Length
	}

	return nil
}

// Returns packet and next state
func decodeHeader(page *Page, data []byte, idx int) (Packet, int, error) {
	if !bytes.Equal(data[idx+1:idx+5], []byte{0x01, 0x01, 0x01, 0x01}) {
		return nil, 0, fmt.Errorf("Packet header at offset %08X has invalid payload: $%08X",
			idx+page.DataOffset, data[idx+1:idx+5])
	}

	if data[idx+5] != data[idx+6] {
		return nil, 0, fmt.Errorf("Packet header at offset %08X has missmatched page numbers at offset %08X: %02X vs %02X",
			idx+page.DataOffset,
			idx+page.DataOffset+5,
			data[idx+5],
			data[idx+6],
		)
	}

	ph := &packetHeader{
		PageNumber: uint8(data[idx+6]),
		Checksum:   data[idx+8],
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			State:  page.state,
			Type:   int(data[idx+1]),
			Length: 8,
		},
	}
	ph.meta.Data = ph.meta.Start + idx + 5

	checksum := calcChecksum(data[idx : idx+ph.meta.Length-1])
	if checksum != ph.Checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for header packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, ph.Checksum)
	}

	return ph, 2, nil
}

func decodeDelay(page *Page, data []byte, idx int) (Packet, int, error) {
	if data[idx+1] != data[idx+2] {
		return nil, 0, fmt.Errorf("State 2 packet at offset %08X has missmatched type [%08X]: %d vs %d",
			idx+page.DataOffset, idx+1+page.DataOffset, data[idx+1], data[idx+2])
	}

	count := 0
	var i int
	for i = idx + 3; i < len(data) && data[i] != 0x00 && data[i] != 0xC5; i++ {
		count++
	}
	if count%2 != 0 {
		fmt.Printf("0xAA delay packet at offset %08X has odd number of 0xAA's", idx+page.FileOffset)
	}
	pd := &packetDelay{
		Length: count,
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			State:  page.state,
			Type:   int(data[idx+1]),
			Length: count + 3,
		},
	}
	pd.meta.Data = pd.meta.Start + 3

	checksum := calcChecksum(data[idx : pd.meta.Length+idx])
	if checksum != 0xC5 {
		return nil, 0, fmt.Errorf("Invalid checksum for delay packet starting at offset %08X. Got %02X, expected %02X", checksum, 0xC5)
	}

	idx += count + 3
	return pd, 1, nil
}

func decodeMarkDataStart(page *Page, data []byte, idx int) (Packet, int, error) {
	packet := &packetMarkDataStart{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 3,
			Length: 6,
			State:  page.state,
			Type:   int(data[idx+1]),
		},
		checksum: data[idx+5],
		ArgA:     data[idx+3],
		ArgB:     data[idx+4],
	}

	checksum := calcChecksum(data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for UnknownS2T3 packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}
	return packet, 1, nil
}

func decodeMarkDataEnd(page *Page, data []byte, idx int) (Packet, int, error) {
	packet := &packetMarkDataEnd{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 2,
			Length: 4,
			State:  page.state,
			Type:   int(data[idx+1]),
		},
		checksum: data[idx+3],
		Arg:      data[idx+2],
		Reset:    (data[idx+2]&0xF0 == 0xF0),
	}

	checksum := calcChecksum(data[idx : idx+3])
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

func decodeUnknownS2(page *Page, data []byte, idx int) (Packet, int, error) {
	packet := &packetUnknown{
		rawData: data[idx : idx+6],
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Data:   page.DataOffset + idx + 3,
			Length: 6,
			State:  page.state,
			Type:   int(data[idx+1]),
		},
		checksum: data[idx+5],
	}

	switch packet.meta.Type {
	case 3:
		packet.notes = "possibly nametable data?"
	case 4:
		packet.notes = "Pattern table data"
	default:
		packet.notes = "???"
	}

	checksum := calcChecksum(data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for UnknownS2T3 packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	return packet, 1, nil
}

// C5 02 02 nn mm zz
// Map 8k ram bank nn to $6000-$7FFF; set load address to $mm00; zz = checksum
func decodeSetWorkRamLoad(page *Page, data []byte, idx int) (Packet, int, error) {
	if data[idx+1] != data[idx+2] {
		return nil, 0, fmt.Errorf("State 1 packet at offset %08X has missmatched type [%08X]: %d vs %d",
			idx+page.DataOffset, idx+1+page.DataOffset, data[idx+1], data[idx+2])
	}

	packet := &packetWorkRamLoad{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Length: 6,
			State:  page.state,
			Type:   2,
		},
		bankId:          data[idx+3],
		loadAddressHigh: data[idx+4],
		checksum:        data[idx+5],
	}
	packet.meta.Data = packet.meta.Start + 3

	checksum := calcChecksum(data[idx : idx+5])
	if checksum != packet.checksum {
		return nil, 0, fmt.Errorf("Invalid checksum for SetWorkRamLoad packet starting at offset %08X. Got %02X, expected %02X",
			page.DataOffset+idx, checksum, packet.checksum)
	}

	return packet, 1, nil
}

func decodeBulkData(page *Page, data []byte, idx int) (Packet, int, error) {
	if data[idx+1] == 0 {
		return nil, 0, fmt.Errorf("Bulk data packet has a length of zero at offset %08X",
			page.DataOffset+idx)
	}

	packet := &packetBulkData{
		meta: PacketMeta{
			Start:  page.DataOffset + idx,
			Length: int(data[idx+1]) + 3,
			State:  page.state,
			Type:   1,
		},
		//data:     data[idx+2 : idx+2+int(data[idx+1])],
		//checksum: data[data[idx+1]+3],
	}

	datalen := int(data[idx+1])
	packet.Data = data[idx+2 : idx+2+datalen]
	packet.checksum = data[idx+len(packet.Data)+2]

	checksum := calcChecksum(data[idx : idx+packet.meta.Length-1])
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
