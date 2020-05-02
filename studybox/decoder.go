package studybox

//sntaohu

import (
	"bytes"
	"fmt"
	"strings"
	//"encoding/binary"
)

type DecodedData struct {
	Packets []Packet
}

func (dd *DecodedData) String() string {
	str := []string{}
	for _, p := range dd.Packets {
		//s := []string{}
		//for _, b := range p.RawBytes() {
		//	s = append(s, fmt.Sprintf("%02X", b))
		//}
		str = append(str, fmt.Sprintf("%08X: %s", p.Offset(), p.Asm()))
	}
	return strings.Join(str, "\n")
}

type Packet interface {
	RawBytes() []byte
	Asm() string
	Offset() int // offset in the file to the start of the data packet
}

type PacketHeader struct {
	PageNumber  uint8
	Checksum    uint8
	startOffset int
}

func (ph *PacketHeader) RawBytes() []byte {
	return []byte{0xC5, 0x01, 0x01, 0x01, 0x01, byte(ph.PageNumber), byte(ph.PageNumber), 0xC5}
}

func (ph *PacketHeader) Asm() string {
	return fmt.Sprintf("header %d", ph.PageNumber)
}

func (ph *PacketHeader) Offset() int {
	return ph.startOffset
}

type PacketDelay struct {
	Length      int
	startOffset int
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

func (pd *PacketDelay) Offset() int {
	return pd.startOffset
}

func DecodePage(page *Page) (*DecodedData, error) {
	decoded := &DecodedData{
		Packets: []Packet{},
	}

	fmt.Printf("Decoding page %s\n", page)

	state := 0
	for idx := 0; idx < len(page.Data); {
		fmt.Printf("%08X\n", idx+page.StartOffset)

		switch state {
		case 0: // start
			if page.Data[idx] != 0xC5 {
				return decoded, fmt.Errorf("Packet at offset %08X does not start with $C5: %02X", idx+page.StartOffset+16, page.Data[idx])
			}

			if !bytes.Equal(page.Data[idx+1:idx+5], []byte{0x01, 0x01, 0x01, 0x01}) {
				return decoded, fmt.Errorf("Packet header at offset %08X has invalid payload: $%08X", idx+page.StartOffset+16, page.Data[idx+1:idx+5])
			}

			if page.Data[idx+5] != page.Data[idx+6] {
				return decoded, fmt.Errorf("Packet header at offset %08X has missmatched page numbers at offset %08X: %02X vs %02X",
					idx+page.StartOffset+16,
					idx+page.StartOffset+5+16,
					page.Data[idx+5],
					page.Data[idx+6],
				)
			}

			ph := &PacketHeader{
				PageNumber:  uint8(page.Data[idx+6]),
				Checksum:    page.Data[idx+8],
				startOffset: page.StartOffset + idx + 16,
			}

			decoded.Packets = append(decoded.Packets, ph)
			state = 2
			idx += 8
			// TODO: check the checksum

		//case 1: // reading data
		case 2: // delay and wait for audio
			if page.Data[idx] != 0xC5 {
				return decoded, fmt.Errorf("State 2 packet at offset %08X does not start with $C5", idx+page.StartOffset)
			}

			if page.Data[idx+1] != page.Data[idx+2] {
				return decoded, fmt.Errorf("State 2 packet at offset %08X has missmatched type: %d vs %d", idx+page.StartOffset, page.Data[idx+1], page.Data[idx+2])
			}

			switch page.Data[idx+1] {
			case 0x05:
				// Delay
				count := 0
				var i int
				for i = idx + 3; i < len(page.Data) && page.Data[i] == 0xAA; i++ {
					count++
				}
				if count%2 != 0 {
					fmt.Printf("0xAA delay packet at offset %08X has odd number of 0xAA's", idx+page.StartOffset)
				}
				pd := &PacketDelay{
					Length:      count,
					startOffset: page.StartOffset + idx,
				}
				decoded.Packets = append(decoded.Packets, pd)
				idx += count + 3
				// TODO: start back up here.  fix the offfset for the next packet... i think
			default:
				return decoded, fmt.Errorf("State 2 packet at offset %08X and type %d isn't implemented", idx+page.StartOffset, page.Data[idx+1])
			}
		default:
			return decoded, fmt.Errorf("Unknown state at offset %08X", idx+page.StartOffset)
		}
	}

	return decoded, nil
}
