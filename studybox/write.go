package studybox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

func (sb *StudyBox) Write(filename string) error {
	raw, err := sb.rawBytes()
	if err != nil {
		return err
	}
	if filename == "" {
		filename = "output.studybox"
	}

	fmt.Println("Writing to " + filename)

	return ioutil.WriteFile(filename, raw, 0777)
}

func (sb *StudyBox) rawBytes() ([]byte, error) {
	buffer := &bytes.Buffer{}
	_, err := buffer.WriteString("STBX")
	if err != nil {
		return nil, err
	}

	// Remaining field length
	err = binary.Write(buffer, binary.LittleEndian, uint32(4))
	if err != nil {
		return nil, fmt.Errorf("Error writing field length: %v", err)
	}

	// Version number (* 0x100)
	err = binary.Write(buffer, binary.LittleEndian, uint32(1*0x100))
	if err != nil {
		return nil, fmt.Errorf("Error writing version: %v", err)
	}

	for _, page := range sb.Data.Pages {
		raw, err := page.rawBytes()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(raw)
		if err != nil {
			return nil, err
		}
	}

	_, err = buffer.WriteString("AUDI")
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, uint32(len(sb.Audio.Data)))
	if err != nil {
		return nil, err
	}

	var format uint32
	switch sb.Audio.Format {
	case AUDIO_WAV:
		format = 0
	case AUDIO_FLAC:
		format = 1
	case AUDIO_OGG:
		format = 2
	case AUDIO_MP3:
		format = 3

	default:
		return nil, fmt.Errorf("Unsupported audio format: %s", sb.Audio.Format)
	}

	err = binary.Write(buffer, binary.LittleEndian, format)
	if err != nil {
		return nil, err
	}

	// For some reason there's 4 extra bytes. no idea why.  chomp them off.
	_, err = buffer.Write(sb.Audio.Data[0 : uint32(len(sb.Audio.Data))-4])
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (page *Page) rawBytes() ([]byte, error) {
	fieldBuffer := &bytes.Buffer{}

	err := binary.Write(fieldBuffer, binary.LittleEndian, uint32(page.AudioOffsetLeadIn))
	if err != nil {
		return nil, err
	}

	err = binary.Write(fieldBuffer, binary.LittleEndian, uint32(page.AudioOffsetData))
	if err != nil {
		return nil, err
	}

	for _, packet := range page.Packets {
		_, err = fieldBuffer.Write(packet.RawBytes())
		if err != nil {
			return nil, err
		}
	}

	pageBuffer := &bytes.Buffer{}
	_, err = pageBuffer.WriteString("PAGE")
	if err != nil {
		return nil, err
	}

	err = binary.Write(pageBuffer, binary.LittleEndian, uint32(fieldBuffer.Len()))
	if err != nil {
		return nil, err
	}

	_, err = pageBuffer.Write(fieldBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	return pageBuffer.Bytes(), nil
}
