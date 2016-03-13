// framecut removes the transparent frame from a png image so only the rect of
// the image remains that actually has content.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	flag "github.com/ogier/pflag"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Framecut is a tool to remove a transparent frame from a picture.\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [-f px] [-o] file [file...]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var frame int
	var override bool
	flag.IntVarP(&frame, "frame", "f", 0, "Keep n pixel of the frame if any")
	flag.BoolVarP(&override, "override", "o", false, "Override original file on save")
	flag.Usage = usage
	flag.Parse()

	for _, file := range flag.Args() {
		if err := processFile(file, frame, override); err != nil {
			log.Fatalln(err)
		}
	}
}

func processFile(file string, frame int, override bool) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	im, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return err
	}

	// find content rect
	size := im.Bounds()
	minY := 0
T1:
	for y := 0; y < size.Max.Y; y++ {
		for x := 0; x < size.Max.X; x++ {
			if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
				minY = y
				break T1
			}
		}
	}
	maxY := 0
T2:
	for y := size.Max.Y; y >= 0; y-- {
		for x := 0; x < size.Max.X; x++ {
			if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
				maxY = y
				break T2
			}
		}
	}
	minX := 0
T3:
	for x := 0; x < size.Max.X; x++ {
		for y := 0; y < size.Max.Y; y++ {
			if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
				minX = x
				break T3
			}
		}
	}
	maxX := 0
T4:
	for x := size.Max.X; x >= 0; x-- {
		for y := 0; y < size.Max.Y; y++ {
			if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
				maxX = x
				break T4
			}
		}
	}

	if maxX+frame > size.Max.X {
		frame = size.Max.X
	}
	if maxY+frame > size.Max.Y {
		frame = size.Max.Y
	}
	if minX-frame < 0 {
		frame = minX
	}
	if minY-frame < 0 {
		frame = minY
	}

	// write rect to new image
	dest := image.NewRGBA(image.Rect(0, 0, maxX-minX+1+frame*2, maxY-minY+1+frame*2))
	draw.Draw(dest, dest.Bounds(), im, image.Pt(minX-frame, minY-frame), draw.Src)

	// add ".cut" to extenstions
	if !override {
		ext := filepath.Ext(file)
		file = file[:len(file)-len(ext)] + ".cut" + ext
	}

	f, err = os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dest)
}
