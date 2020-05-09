package studybox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	//"reflect"
)

// .studybox file format

type StudyBox struct {
	Data  *TapeData
	Audio *TapeAudio
}

type dataType int

const (
	DT_None dataType = iota
	DT_Nametable
	DT_Pattern
	DT_Script
)

/* --------------------- */

func Read(filename string) (*StudyBox, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sb, err := readTape(raw)
	if err != nil {
		return nil, err
	}

	for _, page := range sb.Data.Pages {
		err = page.decode()
		if err != nil {
			fmt.Printf("==> %v\n", err)
		}
	}

	return sb, nil
}

func (sb *StudyBox) Write(filename string) error {
	return fmt.Errorf("Not implemented")
}

func Import(filename string) (*StudyBox, error) {
	return nil, fmt.Errorf("Not implemented")
}

type StudyBoxJson struct {
	Version uint
	Pages   []jsonPage
	Audio   string // filename of the audio
}

type jsonPage struct {
	AudioOffsetLeadIn int
	AudioOffsetData   int
	Data              []jsonData
}

type jsonData struct {
	Type   string
	Values []int
	File   string `json:",omitempty"`
	Reset  bool   `json:",omitempty"`
}

func (sb *StudyBox) Export(directory string) error {
	sbj := StudyBoxJson{
		Version: 1,
		Pages:   []jsonPage{},
		Audio:   directory + "/audio" + sb.Audio.ext(),
	}

	for pidx, page := range sb.Data.Pages {
		jp := jsonPage{
			AudioOffsetLeadIn: page.AudioOffsetLeadIn,
			AudioOffsetData:   page.AudioOffsetData,
			Data:              []jsonData{},
		}

		file, err := os.Create(fmt.Sprintf("%s/Page_%02d.txt", directory, pidx))
		if err != nil {
			return err
		}
		fmt.Fprintln(file, page.InfoString())
		file.Close()

		var dataStartId int
		chrData := []byte{}
		ntData := []byte{}
		scriptData := []byte{}

		data := jsonData{}
		for i, packet := range page.Packets {
			switch p := packet.(type) {
			case *packetHeader:
				data.Type = "header"
				data.Values = []int{int(p.PageNumber)}

				jp.Data = append(jp.Data, data)
				data = jsonData{}

			case *packetDelay:
				data.Type = "delay"
				data.Values = []int{p.Length}

			case *packetWorkRamLoad:
				data.Type = "script"
				data.Values = []int{int(p.bankId), int(p.loadAddressHigh)}

			case *packetPadding:
				data.Type = "padding"
				data.Values = []int{p.Length}
				data.Reset = false

				jp.Data = append(jp.Data, data)
				data = jsonData{}

			case *packetMarkDataStart:
				data.Values = []int{int(p.ArgA), int(p.ArgB)}
				data.Type = p.dataType()
				dataStartId = i

			case *packetMarkDataEnd:
				data.Reset = p.Reset
				var rawData []byte

				switch data.Type {
				case "pattern":
					if len(chrData) == 0 {
						continue
					}
					data.File = fmt.Sprintf("%s/chrData_page%02d_%04d.chr", directory, pidx, dataStartId)
					rawData = chrData
					chrData = []byte{}

				case "nametable":
					if len(ntData) == 0 {
						continue
					}
					data.File = fmt.Sprintf("%s/ntData_page%02d_%04d.dat", directory, pidx, dataStartId)
					rawData = chrData
					ntData = []byte{}

				case "script":
					if len(scriptData) == 0 {
						continue
					}

					data.File = fmt.Sprintf("%s/scriptData_page%02d_%04d.dat", directory, pidx, dataStartId)

					script, err := DissassembleScript(scriptData)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Printf("Script OK Page %02d @ %04d\n", pidx, dataStartId)
						err = script.WriteToFile(fmt.Sprintf("%s/script_page%02d_%04d.txt", directory, pidx, dataStartId))
						if err != nil {
							return fmt.Errorf("Unable to write data to file: %v", err)
						}
					}

					rawData = scriptData
					scriptData = []byte{}
				default:
					jp.Data = append(jp.Data, data)
					data = jsonData{}
					continue
				}

				err = ioutil.WriteFile(data.File, rawData, 0777)
				if err != nil {
					return fmt.Errorf("Unable to write data to file [%q]: %v", data.File, err)
				}

				jp.Data = append(jp.Data, data)
				data = jsonData{}

			case *packetBulkData:
				switch data.Type {
				case "pattern":
					chrData = append(chrData, p.Data...)
				case "nametable":
					ntData = append(ntData, p.Data...)
				case "script":
					scriptData = append(scriptData, p.Data...)
				}

			default:
				return fmt.Errorf("Encountered an unknown packet: %s page: %d", p.Asm(), pidx)
			}
		}

		sbj.Pages = append(sbj.Pages, jp)
	}

	if sb.Audio == nil {
		return fmt.Errorf("Missing audio!")
	}

	err := sb.Audio.WriteToFile(directory + "/audio")
	if err != nil {
		return fmt.Errorf("Error writing audio file: %v", err)
	}

	rawJson, err := json.MarshalIndent(sbj, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(directory+".json", rawJson, 0777)
}

func (sb StudyBox) String() string {
	return fmt.Sprintf("%s\n%s", sb.Data.String(), sb.Audio.String())
}

/* --------------------- */

type TapeData struct {
	Identifier string // MUST be "STBX"
	Length     int    // length of everything following this field (excluding Pages)
	Version    int

	Pages []*Page
}

func (td TapeData) String() string {
	return fmt.Sprintf("%s %d %v", td.Identifier, td.Length, td.Pages)
}

type Page struct {
	Identifier        string // MUST be "PAGE"
	Length            int
	AudioOffsetLeadIn int
	AudioOffsetData   int
	FileOffset        int // offset in the file
	DataOffset        int // offset in the file for the data

	Data    []byte
	Packets []Packet
}

func (page Page) InfoString() string {
	str := []string{}
	for _, p := range page.Packets {
		str = append(str, fmt.Sprintf("%08X: %s", p.Meta().Start, p.Asm()))
	}
	return strings.Join(str, "\n")
}

func (p Page) String() string {
	return fmt.Sprintf("%s @ %08X: %d %d %d %d",
		p.Identifier,
		p.FileOffset,
		p.Length,
		p.AudioOffsetLeadIn,
		p.AudioOffsetData,
		len(p.Data),
	)
}

type AudioType string

const (
	AUDIO_WAV  AudioType = "WAV"
	AUDIO_FLAC AudioType = "FLAC"
	AUDIO_OGG  AudioType = "OGG"
	AUDIO_MP3  AudioType = "MP3"
)

type TapeAudio struct {
	Identifier string // MUST be "AUDI"
	Length     int
	Format     AudioType
	Data       []byte
}

func (ta TapeAudio) String() string {
	return fmt.Sprintf("%s %d %s %d", ta.Identifier, ta.Length, ta.Format, len(ta.Data))
}

func (ta *TapeAudio) WriteToFile(basename string) error {
	ext := "." + strings.ToLower(string(ta.Format))
	return ioutil.WriteFile(basename+ext, ta.Data, 0777)
}

func (ta *TapeAudio) ext() string {
	return "." + strings.ToLower(string(ta.Format))
}

func readTape(data []byte) (*StudyBox, error) {
	// check for length and identifier
	if len(data) < 16 {
		return nil, fmt.Errorf("Not enough data")
	}

	if !bytes.Equal(data[:4], []byte("STBX")) {
		return nil, fmt.Errorf("Missing STBX identifier")
	}

	sb := &StudyBox{
		Data: &TapeData{
			Identifier: "STBX",
		},
	}

	// header data length and version
	sb.Data.Length = int(binary.LittleEndian.Uint32(data[4:8]))
	sb.Data.Version = int(binary.LittleEndian.Uint32(data[8:12]))

	// decode page chunks
	var idx = 12
	if string(data[idx:idx+4]) != "PAGE" {
		return nil, fmt.Errorf("Missing PAGE chunks")
	}

	for string(data[idx:idx+4]) == "PAGE" {
		page, err := decodePage(idx+4, data)
		if err != nil {
			return nil, err
		}
		page.FileOffset = idx
		page.DataOffset = idx + 16
		idx += page.Length + 8
		sb.Data.Pages = append(sb.Data.Pages, page)
	}

	// audio is a single chunk
	if string(data[idx:idx+4]) != "AUDI" {
		return nil, fmt.Errorf("Missing AUDI chunk")
	}

	audio, err := decodeAudio(idx+4, data)
	if err != nil {
		return nil, err
	}
	sb.Audio = audio

	return sb, nil
}

func decodePage(start int, data []byte) (*Page, error) {
	tp := &Page{Identifier: "PAGE"}

	if len(data) < start+12 {
		return nil, fmt.Errorf("Not enough data in PAGE")
	}

	tp.Length = int(binary.LittleEndian.Uint32(data[start+0 : start+4]))
	tp.AudioOffsetLeadIn = int(binary.LittleEndian.Uint32(data[start+4 : start+8]))
	tp.AudioOffsetData = int(binary.LittleEndian.Uint32(data[start+8 : start+12]))

	if tp.Length > len(data)-start+12 {
		return nil, fmt.Errorf("PAGE Length too large: %d with %d bytes remaining.",
			tp.Length, len(data)-start)
	}

	if tp.AudioOffsetLeadIn > len(data)-start+12 {
		return nil, fmt.Errorf("PAGE Audio offest lead-in too large: %d with %d bytes remaining.",
			tp.Length, len(data)-start)
	}

	if tp.AudioOffsetData > len(data)-start+12 {
		return nil, fmt.Errorf("PAGE Audio offest data too large: %d with %d bytes remaining.",
			tp.Length, len(data)-start)
	}

	tp.Data = data[start+12 : start+12+tp.Length-1]
	return tp, nil
}

func decodeAudio(start int, data []byte) (*TapeAudio, error) {
	if len(data) < start+12 {
		return nil, fmt.Errorf("Not enough data in AUDI")
	}

	ta := &TapeAudio{
		Identifier: "AUDI",
		Length:     int(binary.LittleEndian.Uint32(data[start : start+4])),
	}
	format := binary.LittleEndian.Uint32(data[start+4 : start+8])
	switch format {
	case 0:
		ta.Format = AUDIO_WAV
	case 1:
		ta.Format = AUDIO_FLAC
	case 2:
		ta.Format = AUDIO_OGG
	case 3:
		ta.Format = AUDIO_MP3
	default:
		return nil, fmt.Errorf("Unknown audio format: %d", format)
	}

	ta.Data = data[start+8 : start+8+ta.Length]

	return ta, nil
}
