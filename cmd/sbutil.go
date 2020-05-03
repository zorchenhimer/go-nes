package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zorchenhimer/go-nes/studybox"
)

const inputDir string = `` //`F:/nes/StudyBox/studybox-emu/StudyBoxOthers/`
const inputFile string = `ニュートンランド小3 算数 04月号 (13CE04).studybox`

func main() {
	var err error

	raw, err := ioutil.ReadFile(inputDir + inputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("length: %d\n", len(raw))

	sb, err := studybox.ReadTape(raw)
	if err != nil {
		fmt.Println(err)
	}

	for _, page := range sb.Data.Pages {
		fmt.Println(page)
	}

	for pidx, page := range sb.Data.Pages {
		decoded, err := studybox.DecodePage(page)
		if err != nil {
			fmt.Printf("==> %v\n", err)
		}

		file, err := os.Create(fmt.Sprintf("Page_%02d.txt", pidx))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Fprintln(file, decoded)
		file.Close()

		var data bool
		chrData := []byte{}
		for _, packet := range decoded.Packets {
			meta := packet.Meta()
			if meta.State == 2 && meta.Type == 4 {
				data = true
			} else if bulk, ok := packet.(*studybox.PacketBulkData); data && ok {
				chrData = append(chrData, bulk.Data...)
			} else if meta.State == 1 && meta.Type == 0 {
				data = false
			}
		}

		if len(chrData) > 0 {
			err = ioutil.WriteFile(fmt.Sprintf("chrData_page%02d.chr", pidx), chrData, 0777)
			if err != nil {
				fmt.Println("Unable to write data to file: ", err)
				os.Exit(1)
			}
		}
	}
}
