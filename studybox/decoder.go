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

func DecodePage(page *Page) (*DecodedData, error) {
	decoded := &DecodedData{
		Packets: []Packet{},
	}

	fmt.Printf("Decoding page %s\n", page)

	state := 0
	for idx := 0; idx < len(page.Data); {
		fmt.Printf("decoding packet at file offset: %08X\n", idx+page.DataOffset)

		if page.Data[idx] != 0xC5 {
			return decoded, fmt.Errorf("Packet at offset %08X does not start with $C5: %02X", idx+page.DataOffset, page.Data[idx])
		}

		switch state {
		case 0: // start
			switch page.Data[idx+1] {
			case 0x01:
				if !bytes.Equal(page.Data[idx+1:idx+5], []byte{0x01, 0x01, 0x01, 0x01}) {
					return decoded, fmt.Errorf("Packet header at offset %08X has invalid payload: $%08X", idx+page.DataOffset, page.Data[idx+1:idx+5])
				}

				if page.Data[idx+5] != page.Data[idx+6] {
					return decoded, fmt.Errorf("Packet header at offset %08X has missmatched page numbers at offset %08X: %02X vs %02X",
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
						State: 0,
						Type:  1,
					},
				}

				ph.meta.Start = page.DataOffset + idx
				ph.meta.Data = ph.meta.Start + idx + 5
				ph.meta.Length = 8

				//fmt.Printf("Header packet\n  page.FileOffset: %08X page.DataOffset: %08X\n  offsets.Start: %08X offsets.Data: %08X\n  idx: %d Length: %d\n",
				//	page.FileOffset,
				//	page.DataOffset,
				//	ph.meta.Start,
				//	ph.meta.Data,
				//	idx,
				//	ph.meta.Length,
				//)

				checksum := calcChecksum(page.Data[idx : idx+ph.meta.Length-1])
				if checksum != ph.Checksum {
					return decoded, fmt.Errorf("Invalid checksum for header packet starting at offset %08X. Got %02X, expected %02X",
						page.DataOffset+idx, checksum, ph.Checksum)
				}

				decoded.Packets = append(decoded.Packets, ph)
				state = 2
				idx += 8

			default:
				return decoded, fmt.Errorf("State 0 packet at offset %08X with type %d isn't implemented", idx+page.FileOffset, page.Data[idx+1])
			}

		case 1: // reading data
			switch page.Data[idx+1] {
			case 0x00:
				// unknown packet
				unk := &PacketUnknown{
					rawData: page.Data[idx : idx+4],
					meta: PacketMeta{
						Start:  page.DataOffset + idx,
						Data:   page.DataOffset + idx + 2,
						Length: 4,
						State:  2,
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

				if page.Data[idx+2]&0xF0 == 0xF0 {
					state = 0
				} else {
					state = 2
				}

				decoded.Packets = append(decoded.Packets, unk)
				idx += 4

			default:
				return decoded, fmt.Errorf("State 1 packet at offset %08X with type %d isn't implemented", idx+page.FileOffset, page.Data[idx+1])
			}

			//case 0x02:
		case 2: // delay and wait for audio
			if page.Data[idx] != 0xC5 {
				return decoded, fmt.Errorf("State 2 packet at offset %08X does not start with $C5", idx+page.FileOffset)
			}

			if page.Data[idx+1] != page.Data[idx+2] {
				return decoded, fmt.Errorf("State 2 packet at offset %08X has missmatched type [%08X]: %d vs %d",
					idx+page.DataOffset, idx+1+page.DataOffset, page.Data[idx+1], page.Data[idx+2])
			}

			switch page.Data[idx+1] {

			case 0x05:
				// Delay
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
					meta:   PacketMeta{},
					//startOffset: page.StartOffset + idx,
				}

				pd.meta.Start = page.DataOffset + idx
				pd.meta.Data = pd.meta.Start + idx + 3
				pd.meta.Length = count + 3

				checksum := calcChecksum(page.Data[idx : pd.meta.Length+idx])
				if checksum != 0xC5 {
					return decoded, fmt.Errorf("Invalid checksum for delay packet starting at offset %08X. Got %02X, expected %02X", checksum, 0xC5)
				}

				decoded.Packets = append(decoded.Packets, pd)
				idx += count + 3
				state = 1
			default:
				return decoded, fmt.Errorf("State 2 packet at offset %08X with type %d isn't implemented", idx+page.DataOffset, page.Data[idx+1])
			}
		default:
			return decoded, fmt.Errorf("Unknown state at offset %08X", idx+page.FileOffset)
		}
	}

	return decoded, nil
}

func calcChecksum(data []byte) uint8 {
	var sum uint8
	for i := 0; i < len(data); i++ {
		sum ^= data[i]
	}
	return sum
}
