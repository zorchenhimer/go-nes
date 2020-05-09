package main

import (
	"fmt"
	//"io/ioutil"
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
		fmt.Println("-- Processing " + file)
		outDir := filepath.Base(file)
		outDir = strings.ReplaceAll(outDir, ".studybox", "")

		err := os.MkdirAll(outDir, 0777)
		if err != nil {
			fmt.Println(err)
			continue
		}

		sb, err := studybox.Read(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = sb.Export(outDir)
		if err != nil {
			fmt.Println(err)
		}
	}
}
