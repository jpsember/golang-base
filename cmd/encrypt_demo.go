package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/gen/sample"
)

func main() {
	var oper = &EncryptOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("encrypt_demo")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs()
	app.AddTestArgs("--verbose")
	app.Start()
}

type EncryptOper struct {
	BaseObject
}

func (oper *EncryptOper) UserCommand() string {
	return "encrypt"
}

func (oper *EncryptOper) Perform(app *App) {

	pth := NewPathM("a/b/c")
	pth.ReadBytesM()

	oper.SetVerbose(true)
	pr := oper.Log

	password := "thatwaseasy"
	bytes := []byte{
		135, 120, 92, 46, 178, 72, 115, 146, 187, 200, 150, 249, 46, 22, 193,
		253, 108, 81, 238, 165, 135, 186, 254, 91, 115, 17, 59, 62, 189, 19, 21,
		29, 165, 33, 228, 169, 64, 12, 185, 40, 104, 18, 153, 64, 168, 29, 124,
		135, 219, 53, 45, 177, 28, 196, 238, 103, 202,
	}

	if false {
		message := "Hello, World!"
		encrypted := CheckOkWith(EncryptBytes([]byte(message), password))

		pr("encrypted:", INDENT, encrypted)
		pr("expected :", INDENT, bytes)

		bytes = encrypted
	}

	decrypted := CheckOkWith(DecryptBytes(bytes, password))

	Pr("Decrypted:", decrypted)
	Pr("Message  :", string(decrypted))

	h := sample.DefaultDemoConfig
	Pr(h)
	j := h.ToBuilder()
	j.SetName("alpha")
	Pr(j)
	Pr(h)
}

func (oper *EncryptOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Performs AES encryption/decryption.")
}

func (oper *EncryptOper) ProcessArgs(c *CmdLineArgs) {
	for c.HasNextArg() {
		var arg = c.NextArg()
		switch arg {
		default:
			c.SetError("extraneous argument:", arg)
		}
	}
}
