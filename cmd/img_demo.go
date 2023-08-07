package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/img"
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
	originalImage := CheckOkWith(img.DecodeImage(bytes))
	Pr("original (std lib), info:", originalImage.ToJson())

	{
		targ, err := originalImage.AsType(img.TypeNRGBA)
		CheckOk(err)

		targetPath := NewPathM("_SKIP_image_of_type.png")

		pngBytes := CheckOkWith(targ.ToPNG())
		targetPath.WriteBytesM(pngBytes)

		Pr("converted:", img.GetImageInfo(targ.Image()))
		return
	}

}
