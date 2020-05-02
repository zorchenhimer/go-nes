package studybox

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// .studybox file format

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
	StartOffset       int // offset in the file

	Data []byte
}

func (p Page) String() string {
	return fmt.Sprintf("%s @ %08X: %d %d %d %d",
		p.Identifier,
		p.StartOffset,
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

func ReadTape(data []byte) (*StudyBox, error) {
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
		page.StartOffset = idx
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
