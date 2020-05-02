package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zorchenhimer/go-nes/studybox"
)

const inputDir string = `F:/nes/StudyBox/studybox-emu/StudyBoxOthers/`
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

	decoded, err := studybox.DecodePage(sb.Data.Pages[0])
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(decoded)
}
