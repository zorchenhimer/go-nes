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

	for _, page := range sb.Data.Pages {
		fmt.Println(page)
	}

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
			err = ioutil.WriteFile(fmt.Sprintf("%s/chrData_page%02d.chr", outDir, pidx), chrData, 0777)
			if err != nil {
				return fmt.Errorf("Unable to write data to file: ", err)
			}
		}
	}

	return nil
}
