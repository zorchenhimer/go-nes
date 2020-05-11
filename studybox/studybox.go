package studybox

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type StudyBox struct {
	Data  *TapeData
	Audio *TapeAudio
}

func (sb StudyBox) String() string {
	return fmt.Sprintf("%s\n%s", sb.Data.String(), sb.Audio.String())
}

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

	//Data    []byte
	Packets []Packet
	state   int
}

func (p *Page) Debug() string {
	lines := []string{}
	for _, packet := range p.Packets {
		raw := packet.RawBytes()
		s := []string{}
		for _, b := range raw {
			s = append(s, fmt.Sprintf("%02X", b))
		}

		lines = append(lines, fmt.Sprintf("%s: %s", packet.Name(), strings.Join(s, " ")))
	}

	return strings.Join(lines, "\n")
}

func (page *Page) InfoString() string {
	str := []string{}
	for _, p := range page.Packets {
		str = append(str, fmt.Sprintf("%08X: %s", p.Address(), p.Asm()))
	}
	return strings.Join(str, "\n")
}

func (p Page) String() string {
	return fmt.Sprintf("%s @ %08X: %d %d %d",
		p.Identifier,
		p.FileOffset,
		p.Length,
		p.AudioOffsetLeadIn,
		p.AudioOffsetData,
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

func readAudio(filename string) (*TapeAudio, error) {
	ta := &TapeAudio{
		Identifier: "AUDI",
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".wav":
		ta.Format = AUDIO_WAV
	case ".flac":
		ta.Format = AUDIO_FLAC
	case ".ogg":
		ta.Format = AUDIO_OGG
	case ".mp3":
		ta.Format = AUDIO_MP3
	default:
		return nil, fmt.Errorf("Unsupported audio format: %s", filepath.Ext(filename))
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	ta.Data = raw

	return ta, nil
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

type Packet interface {
	RawBytes() []byte
	Asm() string
	Address() int // Address this packet starts in the .studybox file (if loaded from a .studybox file)
	Name() string
}

func calcChecksum(data []byte) uint8 {
	var sum uint8
	for i := 0; i < len(data); i++ {
		sum ^= data[i]
	}
	return sum
}

type StudyBoxJson struct {
	Version  uint
	Filename string // .studybox filename.  defaults to the name of the json file if empty.
	Audio    string // filename of the audio
	Pages    []jsonPage
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

func byteString(data []byte) string {
	s := []string{}
	for _, b := range data {
		s = append(s, fmt.Sprintf("$%02X", b))
	}
	return strings.Join(s, ", ")
}
