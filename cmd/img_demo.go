package main

import (
	"bufio"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/img"
	"github.com/sunshineplan/imgconv"
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

	// Read the original image
	originalImage := CheckOkWith(imgconv.Open("img/resources/balloons.jpg"))

	// Resize the image to a particular width, preserving the aspect ratio.
	modifiedImage := imgconv.Resize(originalImage, &imgconv.ResizeOption{Width: 40})

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
