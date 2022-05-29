package studybox

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Import(filename string) (*StudyBox, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		return nil, fmt.Errorf("Can only import .json files")
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sbj := &StudyBoxJson{}
	err = json.Unmarshal(raw, sbj)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal json: %v", err)
	}

	audio, err := readAudio(sbj.Audio)
	if err != nil {
		return nil, fmt.Errorf("Unable to read audio: %v", err)
	}

	sb := &StudyBox{
		Data:  &TapeData{Pages: []*Page{}},
		Audio: audio,
	}

	for _, jpage := range sbj.Pages {
		page := &Page{
			AudioOffsetLeadIn: jpage.AudioOffsetLeadIn,
			AudioOffsetData:   jpage.AudioOffsetData,
		}

		packets, err := importPackets(jpage.Data)
		if err != nil {
			return nil, err
		}
		page.Packets = packets

		sb.Data.Pages = append(sb.Data.Pages, page)
	}

	return sb, nil
}

func importPackets(jdata []jsonData) ([]Packet, error) {
	packets := []Packet{}
	for idx, data := range jdata {
		switch data.Type {
		case "header":
			if len(data.Values) < 1 {
				return nil, fmt.Errorf("Missing header value from script data in element %d", idx)
			}
			packets = append(packets, newPacketHeader(uint8(data.Values[0])))

		case "delay":
			if len(data.Values) < 1 {
				return nil, fmt.Errorf("Missing delay value from script data in element %d", idx)
			}
			packets = append(packets, newPacketDelay(data.Values[0]))
			packets = append(packets, newPacketMarkDataEnd(packet_Delay, data.Reset))

		case "script":
			if len(data.Values) < 2 {
				return nil, fmt.Errorf("Missing bank id and/or load address high values from script data in element %d", idx)
			}

			if data.File == "" {
				fmt.Println("[WARN] No script file given in data element %d\n", idx)
			}

			packets = append(packets, newPacketWorkRamLoad(uint8(data.Values[0]), uint8(data.Values[1])))
			if data.File != "" {
				raw, err := os.ReadFile(data.File)
				if err != nil {
					return nil, fmt.Errorf("Error reading script data file: %v", err)
				}
				packets = append(packets, newBulkDataPackets(raw)...)
			}
			packets = append(packets, newPacketMarkDataEnd(packet_Script, data.Reset))

		case "nametable":
			if len(data.Values) < 2 {
				return nil, fmt.Errorf("Missing bank id and/or load address high values from nametable data in element %d", idx)
			}

			if data.File == "" {
				fmt.Println("[WARN] No script file given in data element %d\n", idx)
			}

			packets = append(packets, newPacketMarkDataStart(packet_Nametable, uint8(data.Values[0]), uint8(data.Values[1])))
			if data.File != "" {
				raw, err := os.ReadFile(data.File)
				if err != nil {
					return nil, fmt.Errorf("Error reading nametable data file: %v", err)
				}
				packets = append(packets, newBulkDataPackets(raw)...)
			}
			packets = append(packets, newPacketMarkDataEnd(packet_Nametable, data.Reset))

		case "pattern":
			if len(data.Values) < 2 {
				return nil, fmt.Errorf("Missing bank id and/or load address high values from pattern data in element %d", idx)
			}

			if data.File == "" {
				fmt.Printf("[WARN] No pattern file given in data element %d\n", idx)
			}

			packets = append(packets, newPacketMarkDataStart(packet_Pattern, uint8(data.Values[0]), uint8(data.Values[1])))
			if data.File != "" {
				raw, err := os.ReadFile(data.File)
				if err != nil {
					return nil, fmt.Errorf("Error reading pattern data file: %v", err)
				}
				packets = append(packets, newBulkDataPackets(raw)...)
			}
			packets = append(packets, newPacketMarkDataEnd(packet_Pattern, data.Reset))

		case "padding":
			if len(data.Values) < 1 {
				return nil, fmt.Errorf("Missing padding value from script data in element %d", idx)
			}

			packets = append(packets, newPacketPadding(data.Values[0]))

		default:
			return nil, fmt.Errorf("Unknown packet type: %s", data.Type)
		}
	}

	return packets, nil
}
