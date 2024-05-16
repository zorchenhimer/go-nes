package studybox

import (
	"encoding/json"
	"fmt"
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

		file, err := os.Create(fmt.Sprintf("%s/page%02d_0000.txt", directory, pidx))
		if err != nil {
			return err
		}
		fmt.Fprintln(file, page.InfoString())
		file.Close()

		var dataStartId int
		jData := jsonData{}
		rawData := []byte{}

		for i, packet := range page.Packets {
			switch p := packet.(type) {
			case *packetHeader:
				jData.Type = "header"
				jData.Values = []int{int(p.PageNumber)}

				jp.Data = append(jp.Data, jData)
				jData = jsonData{}

			case *packetDelay:
				jData.Type = "delay"
				jData.Values = []int{p.Length}

			case *packetWorkRamLoad:
				jData.Type = "script"
				jData.Values = []int{int(p.bankId), int(p.loadAddressHigh)}
				dataStartId = i

			case *packetPadding:
				jData.Type = "padding"
				jData.Values = []int{p.Length}
				jData.Reset = false

				jp.Data = append(jp.Data, jData)
				jData = jsonData{}

			case *packetMarkDataStart:
				jData.Values = []int{int(p.ArgA), int(p.ArgB)}
				jData.Type = p.dataType()
				dataStartId = i

			case *packetMarkDataEnd:
				jData.Reset = p.Reset

				if jData.Values == nil || len(jData.Values) == 0 {
					fmt.Printf("[WARN] No data at page %d, dataStartId: %d\n", pidx, dataStartId)
					jp.Data = append(jp.Data, jData)
					jData = jsonData{}
					continue
				}

				switch jData.Type {
				case "pattern":
					jData.File = fmt.Sprintf("%s/page%02d_%04d_chrData.chr", directory, pidx, dataStartId)

				case "nametable":
					jData.File = fmt.Sprintf("%s/page%02d_%04d_ntData.dat", directory, pidx, dataStartId)

				case "script":
					jData.File = fmt.Sprintf("%s/page%02d_%04d_scriptData.dat", directory, pidx, dataStartId)

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

				case "delay":
					jp.Data = append(jp.Data, jData)
					jData = jsonData{}
					continue

				default:
					return fmt.Errorf("[WARN] unknown end data type: %s\n", jData.Type)
				}

				err = os.WriteFile(jData.File, rawData, 0777)
				if err != nil {
					return fmt.Errorf("Unable to write data to file [%q]: %v", jData.File, err)
				}

				jp.Data = append(jp.Data, jData)
				jData = jsonData{}
				rawData = []byte{}

			case *packetBulkData:
				if rawData == nil {
					rawData = []byte{}
				}
				rawData = append(rawData, p.Data...)

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

	return os.WriteFile(directory+".json", rawJson, 0777)
}
