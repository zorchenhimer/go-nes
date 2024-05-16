package main

// TODO: Add CHR to the output

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	nesimg "github.com/zorchenhimer/go-nes/image"
	ines "github.com/zorchenhimer/go-nes/rom"
)

type Options struct {
	Input   string `arg:"positional,required" help:"input rom file"`
	Output  string `arg:"required" help:"output image file"`
	ChrSize int    `arg:"-c,--chr-size" default:"8" help:"CHR split size"`
}

func main() {
	args := &Options{}
	arg.MustParse(args)

	switch args.ChrSize {
	case 8, 4, 2, 1, 0:
		// OK
	default:
		fmt.Fprintln(os.Stderr, "--chr-size must be one of: 8, 4, 2, 1, 0")
		os.Exit(1)
	}

	err := Run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run(args *Options) error {
	raw, err := os.ReadFile(args.Input)
	if err != nil {
		return fmt.Errorf("unable te read input: %w", err)
	}

	if args.Output == "" {
		ext := filepath.Ext(args.Input)
		args.Output = filepath.Base(args.Input[:len(args.Input)-len(ext)]) + ".png"
	}

	header, err := ines.ParseHeader(raw)
	if err != nil {
		return fmt.Errorf("error parsing header: %w", err)
	}

	prg := raw[16:header.PrgSize]
	chr := raw[16+header.PrgSize : 16+header.PrgSize+header.ChrSize]

	fmt.Printf("prg:%d(%d) chr:%d(%d)\n", len(prg), header.PrgSize, len(chr), header.ChrSize)
	fmt.Printf("prg:$%06X chr:$%06X\n", len(prg), len(chr))
	slices := len(prg) / (1024 * 16)
	fmt.Printf("slices:%d\n", slices)

	pal := color.Palette{
		color.White,
		color.Black,
	}

	images := []*image.Paletted{}
	for i := 0; i <= slices; i++ {
		start := i * 1024 * 16
		img, err := GetChunkImage(prg[start:start+(1024*16)], pal)
		if err != nil {
			return err
		}

		images = append(images, img)
	}

	finalimg := image.NewPaletted(image.Rect(0, 0, 16*8*len(images), 1024), pal)
	fmt.Printf("finalimg: %#v\n", finalimg.Bounds())
	for i := 0; i < len(images); i++ {
		draw.Draw(
			finalimg,
			image.Rect(i*128, 0, (i*128)+(128), 1024),
			images[i],
			image.Pt(0, 0),
			draw.Over)
	}

	err = WriteImage(finalimg, args.Output)
	if err != nil {
		return err
	}

	if header.ChrSize == 0 {
		return nil
	}

	pt, err := nesimg.ReadCHR(bytes.NewReader(chr))
	if err != nil {
		return err
	}

	// 0%, 33%, 66%, 100% grey
	chrpal, err := nesimg.ParseHexPalette("#000000,#555555,#AAAAAA,#FFFFFF")
	if err != nil {
		return err
	}
	pt.SetPalette(chrpal)

	switch args.ChrSize {
	case 0:
		buff := bytes.NewBuffer([]byte{})
		err = png.Encode(buff, pt)
		if err != nil {
			return err
		}
		err = os.WriteFile("chr.png", buff.Bytes(), 0666)
		if err != nil {
			return err
		}
	default:
		chunk := args.ChrSize * 1024
		count := int(header.ChrSize) / chunk
		tilesPer := chunk / 16

		//fmt.Printf("header.ChrSize:$%06X %d\nargs.ChrSize:%d\nchunk:%04X %d\ncount:%d\ntilesPer:%d\n",
		//	header.ChrSize, header.ChrSize,
		//	args.ChrSize,
		//	chunk, chunk,
		//	count,
		//	tilesPer,
		//)
		for i := 0; i < count; i++ {
			p := nesimg.NewPatternTable()
			//fmt.Printf("start id:%d\n", (tilesPer*i))
			for id := 0; id < tilesPer; id++ {
				//fmt.Printf("adding tile $%02X\n", id)
				p.AddTile(pt.Patterns[id+(tilesPer*i)])
				//fmt.Printf("  %d\n", id+(tilesPer*i))
			}

			buff := bytes.NewBuffer([]byte{})
			err = png.Encode(buff, p)
			if err != nil {
				return err
			}
			err = os.WriteFile(fmt.Sprintf("chr_%04d.png", i), buff.Bytes(), 0666)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func WriteImage(raw image.Image, filename string) error {
	fmt.Println(filename)
	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return png.Encode(outfile, raw)
}

func GetChunkImage(raw []byte, pal color.Palette) (*image.Paletted, error) {
	imgprg := image.NewPaletted(image.Rect(0, 0, 16*8, len(raw)/16), pal)
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		for x := 0; x < 8; x++ {
			if (i*8)+x >= len(imgprg.Pix) {
				return nil, fmt.Errorf("too many pixels!")
			}
			imgprg.Pix[(i*8)+x] = uint8(b & 0x1)
			b = b >> 1
		}
	}
	return imgprg, nil
}
