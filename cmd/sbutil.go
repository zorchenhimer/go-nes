package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zorchenhimer/go-nes/studybox"
)

func main() {
	matches, err := filepath.Glob("*.studybox")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Println("No .studybox files found")
		os.Exit(1)
	}

	for _, file := range matches {
		err = processFile(file)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type dataType int

const (
	DT_None dataType = iota
	DT_Nametable
	DT_Pattern
	DT_Script
)

func processFile(filename string) error {
	outDir := filepath.Base(filename)
	outDir = strings.ReplaceAll(outDir, ".studybox", "")

	err := os.MkdirAll(outDir, 0777)
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	fmt.Printf("length: %d\n", len(raw))

	sb, err := studybox.ReadTape(raw)
	if err != nil {
		fmt.Println(err)
	}

	//for _, page := range sb.Data.Pages {
	//	fmt.Println(page)
	//}

	for pidx, page := range sb.Data.Pages {
		decoded, err := studybox.DecodePage(page)
		if err != nil {
			fmt.Printf("==> %v\n", err)
		}

		file, err := os.Create(fmt.Sprintf("%s/Page_%02d.txt", outDir, pidx))
		if err != nil {
			return err
		}
		fmt.Fprintln(file, decoded)
		file.Close()

		var data dataType = DT_None
		var dataStartId int
		chrData := []byte{}
		ntData := []byte{}
		scriptData := []byte{}

		for i, packet := range decoded.Packets {
			meta := packet.Meta()
			if meta.State == 2 && meta.Type == 4 {
				data = DT_Pattern
				dataStartId = i
			} else if meta.State == 2 && meta.Type == 3 {
				data = DT_Nametable
				dataStartId = i
			} else if meta.State == 2 && meta.Type == 2 {
				data = DT_Script
				dataStartId = i
			} else if meta.State == 1 && meta.Type == 0 {
				switch data {
				case DT_Pattern:
					if len(chrData) > 0 {
						err = ioutil.WriteFile(fmt.Sprintf("%s/chrData_page%02d_%04d.chr", outDir, pidx, dataStartId), chrData, 0777)
						if err != nil {
							return fmt.Errorf("Unable to write data to file: ", err)
						}
					}
					chrData = []byte{}

				case DT_Nametable:
					if len(ntData) > 0 {
						err = ioutil.WriteFile(fmt.Sprintf("%s/ntData_page%02d_%04d.dat", outDir, pidx, dataStartId), ntData, 0777)
						if err != nil {
							return fmt.Errorf("Unable to write data to file: ", err)
						}
					}
					ntData = []byte{}

				case DT_Script:
					if len(scriptData) > 0 {
						err = ioutil.WriteFile(fmt.Sprintf("%s/scriptData_page%02d_%04d.dat", outDir, pidx, dataStartId), scriptData, 0777)
						if err != nil {
							return fmt.Errorf("Unable to write data to file: ", err)
						}

						script, err := studybox.DissassembleScript(scriptData)
						if err != nil {
							return err
						}

						err = script.WriteToFile(fmt.Sprintf("%s/script_page%02d_%04d.txt", outDir, pidx, dataStartId))
						if err != nil {
							return fmt.Errorf("Unable to write data to file: %v", err)
						}
					}
					scriptData = []byte{}
				}

				data = DT_None
			} else {
				if bulk, ok := packet.(*studybox.PacketBulkData); ok {
					switch data {
					case DT_Pattern:
						chrData = append(chrData, bulk.Data...)
					case DT_Nametable:
						ntData = append(ntData, bulk.Data...)
					case DT_Script:
						scriptData = append(scriptData, bulk.Data...)
					}
				}
			}
		}

	}

	return nil
}
