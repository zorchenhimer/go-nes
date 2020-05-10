package studybox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func Import(filename string) (*StudyBox, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sbj := &StudyBoxJson{}
	err = json.Unmarshal(raw, sbj)
	if err != nil {
		return nil, err
	}

	audio, err := readAudio(sbj.Audio)
	if err != nil {
		return nil, err
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
	return nil, fmt.Errorf("importPackets() not implemented")
}
