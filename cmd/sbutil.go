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
	if len(os.Args) < 3 {
		fmt.Println("Missing command")
		os.Exit(1)
	}

	matches := []string{}
	for _, glob := range os.Args[2:len(os.Args)] {
		m, err := filepath.Glob(glob)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		matches = append(matches, m...)
	}

	if len(matches) == 0 {
		fmt.Println("No files found!")
	}

	switch strings.ToLower(os.Args[1]) {
	case "unpack":
		for _, file := range matches {
			fmt.Println("-- Processing " + file)
			outDir := filepath.Base(file)
			outDir = strings.ReplaceAll(outDir, ".studybox", "")

			err := os.MkdirAll(outDir, 0777)
			if err != nil {
				fmt.Println(err)
				continue
			}

			sb, err := studybox.ReadFile(file)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = sb.Export(outDir)
			if err != nil {
				fmt.Println(err)
			}
		}
	case "pack":
		for _, file := range matches {
			fmt.Println("-- Processing " + file)
			sb, err := studybox.Import(file)
			if err != nil {
				fmt.Println(err)
				continue
			}

			outDir := filepath.Base(file)
			outDir = strings.ReplaceAll(outDir, ".json", "_output")

			err = os.MkdirAll(outDir, 0777)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = sb.Export(outDir)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// TODO: put this in the json file?

			err = sb.Write(outDir + ".studybox")
			if err != nil {
				fmt.Println(err)
				continue
			}

		}
	}

}
