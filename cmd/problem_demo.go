package main

import (
	"bufio"
	"bytes"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
)

func fillImage(m *image.NRGBA, clr color.NRGBA) {
	b := m.Pix
	imax := len(b)
	for i := 0; i < imax; i += 4 {
		b[i] = clr.R
		b[i+1] = clr.G
		b[i+2] = clr.B
		b[i+3] = clr.A
	}
}

func main() {

	sourceRect := image.Rect(0, 0, 420, 315)
	sourceImage := image.NewNRGBA(sourceRect)
	fillImage(sourceImage, color.NRGBA{
		R: 255,
		G: 100,
		B: 100,
		A: 255,
	})

	for i := 0; i < 40; i++ {
		tgtSize := image.Pt(140, 80-i)

		destImage := image.NewNRGBA(image.Rect(0, 0, tgtSize.X, tgtSize.Y))
		fillImage(destImage, color.NRGBA{
			R: 100,
			G: 100,
			B: 255,
			A: 255,
		})
		plotRect := image.Rect(0, -26, 140, -26+105)

		draw.BiLinear.Scale(destImage, plotRect, sourceImage, sourceRect, draw.Src, nil)

		path := strconv.Itoa(i) + ".png"

		w := bytes.Buffer{}
		writer := bufio.NewWriter(&w)
		png.Encode(writer, destImage)
		writer.Flush()
		os.WriteFile(path, w.Bytes(), 0644)
	}

}
