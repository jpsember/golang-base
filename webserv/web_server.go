package webserv

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"log"
	"net/http"
)

// Define an app with a single operation

type SampleOper struct {
	https bool
}

func (oper *SampleOper) UserCommand() string {
	return "sample"
}

func (oper *SampleOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates an http or https server")
}

func (oper *SampleOper) ProcessArgs(c *CmdLineArgs) {
}

func Demo() {
	var oper = &SampleOper{}
	var app = NewApp()
	app.Version = "1.0"
	app.CmdLineArgs().Add("https").Desc("secure mode")
	app.RegisterOper(oper)
	app.SetTestArgs("--https")
	app.Start()
}

func (oper *SampleOper) Perform(app *App) {
	var secure = app.CmdLineArgs().Get("https")
	if secure {
		oper.doHttps()
	} else {
		oper.doHttp()
	}
}

func handleHello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

// ------------------------------------------------------------------------------------

func (oper *SampleOper) doHttp() {
	Todo("I don't think I need separate handlers for secure vs non")
	http.HandleFunc("/hello", handleHello)
	Pr("Type:", INDENT, "curl -sL http://localhost:8090/hello")
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// ------------------------------------------------------------------------------------

// https://github.com/denji/golang-tls

// How to get a certificate: server.crt
// https://www.vultr.com/docs/secure-a-golang-web-server-with-a-selfsigned-or-lets-encrypt-ssl-certificate/

func (oper *SampleOper) doHttps() {

	var keyDir = NewPathM("webserv/keys")
	var certPath = keyDir.JoinM("server.crt")
	var keyPath = keyDir.JoinM("server.key")

	Pr("Type:", INDENT, "curl -sL https://localhost/hello")

	http.HandleFunc("/hello", handleHello)
	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
