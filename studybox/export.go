package studybox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

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
						fmt.Printf("[WARN] No pattern data at page %d, dataStartId: %d\n", pidx, dataStartId)
						continue
					}

					data.File = fmt.Sprintf("%s/chrData_page%02d_%04d.chr", directory, pidx, dataStartId)
					rawData = chrData
					chrData = []byte{}

				case "nametable":
					if len(ntData) == 0 {
						fmt.Printf("[WARN] No nametable data at page %d, dataStartId: %d\n", pidx, dataStartId)
						continue
					}

					data.File = fmt.Sprintf("%s/ntData_page%02d_%04d.dat", directory, pidx, dataStartId)
					rawData = ntData
					ntData = []byte{}

				case "script":
					if len(scriptData) == 0 {
						fmt.Printf("[WARN] No script data at page %d, dataStartId: %d\n", pidx, dataStartId)
						continue
					}

					data.File = fmt.Sprintf("%s/scriptData_page%02d_%04d.dat", directory, pidx, dataStartId)

					//script, err := DissassembleScript(scriptData)
					//if err != nil {
					//	fmt.Println(err)
					//} else {
					//	fmt.Printf("Script OK Page %02d @ %04d\n", pidx, dataStartId)
					//	err = script.WriteToFile(fmt.Sprintf("%s/script_page%02d_%04d.txt", directory, pidx, dataStartId))
					//	if err != nil {
					//		return fmt.Errorf("Unable to write data to file: %v", err)
					//	}
					//}

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
