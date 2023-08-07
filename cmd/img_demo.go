package main

import (
	"bufio"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/img"
	"github.com/sunshineplan/imgconv"
	"image"
	"image/draw"
	"os"
)

func main() {
	var oper = &ImgOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("encrypt_demo")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs()
	app.Start()
}

type ImgOper struct {
	BaseObject
}

func (oper *ImgOper) UserCommand() string {
	return "img"
}

func (oper *ImgOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Image manipulation.")
}

func (oper *ImgOper) ProcessArgs(c *CmdLineArgs) {
	for c.HasNextArg() {
		var arg = c.NextArg()
		switch arg {
		default:
			c.SetError("extraneous argument:", arg)
		}
	}
}

func (oper *ImgOper) Perform(app *App) {

	p := NewPathM("img/resources/balloons.jpg")

	// Reading this image using go's standard image library produces a strange format:
	bytes := p.ReadBytesM()
	stdLibraryImage := CheckOkWith(img.DecodeImage(bytes))
	Pr("original (std lib), info:", stdLibraryImage.ToJson())

	// Reading it with the 3rd party produces the same format:
	originalImage := CheckOkWith(imgconv.Open(p.String()))
	Pr("original (3rd party), info:", img.GetImageInfo(originalImage))

	{

		// We construct a target image of our desired format, and redraw the source image into it;
		// This hopefully yields an image of the type we want.

		// See:
		// https://stackoverflow.com/questions/47535474/convert-image-from-image-ycbcr-to-image-rgba
		b := stdLibraryImage.Image().Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(m, m.Bounds(), stdLibraryImage.Image(), b.Min, draw.Src)
		redrawnImage := img.JImageOf(m)
		Pr("redrawn using img/draw, info:", redrawnImage.ToJson())

		// Write the resulting image as PNG.
		targetFile := CheckOkWith(os.Create("_SKIP_redrawn.png"))
		defer targetFile.Close()

		writer := bufio.NewWriter(targetFile)
		CheckOk(imgconv.Write(writer, redrawnImage.Image(), &imgconv.FormatOption{Format: imgconv.PNG}))
		writer.Flush()
	}

	// Resize the image to a particular width, preserving the aspect ratio.
	modifiedImage := imgconv.Resize(originalImage, &imgconv.ResizeOption{Width: originalImage.Bounds().Dx()})

	// The resize operation has changed the image format!
	Pr("modified, info:", img.GetImageInfo(modifiedImage))

	// Write the resulting image as PNG.
	targetFile := CheckOkWith(os.Create("_SKIP_result.png"))
	defer targetFile.Close()

	writer := bufio.NewWriter(targetFile)
	CheckOk(imgconv.Write(writer, modifiedImage, &imgconv.FormatOption{Format: imgconv.PNG}))
	writer.Flush()

	convertedImage := CheckOkWith(imgconv.Open("_SKIP_result.png"))
	Pr("converted, info:", img.GetImageInfo(convertedImage))

}
