package studybox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// Read opens and decodes a `.studybox` file.
func Read(filename string) (*StudyBox, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sb, err := readTape(raw)
	if err != nil {
		return nil, err
	}

	return sb, nil
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
		page, err := unpackPage(idx+4, data)
		if err != nil {
			return nil, err
		}
		idx += page.Length + 8
		sb.Data.Pages = append(sb.Data.Pages, page)
	}

	// audio is a single chunk
	if string(data[idx:idx+4]) != "AUDI" {
		return nil, fmt.Errorf("Missing AUDI chunk")
	}

	audio, err := unpackAudio(idx+4, data)
	if err != nil {
		return nil, err
	}
	sb.Audio = audio

	return sb, nil
}

func unpackPage(start int, data []byte) (*Page, error) {
	tp := &Page{Identifier: "PAGE"}

	tp.FileOffset = start - 4
	tp.DataOffset = start + 12

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

	//tp.Data = data[start+12 : start+12+tp.Length-1]
	err := tp.decode(data[start+12 : start+12+tp.Length-8])
	if err != nil {
		return nil, fmt.Errorf("Error decoding: %v", err)
	}
	return tp, nil
}

func unpackAudio(start int, data []byte) (*TapeAudio, error) {
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
