package webserv

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"log"
	"net/http"
	"time"
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

func handler(msg ...any) func(http.ResponseWriter, *http.Request) {
	var m = ToString(msg...)
	return func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(time.Now().Format(time.ANSIC) + ": " + m + "\n\nRequest URI:" + req.RequestURI))

	}
}

// ------------------------------------------------------------------------------------

func (oper *SampleOper) doHttp() {
	Todo("I don't think I need separate handlers for secure vs non")
	http.HandleFunc("/hello", handler("hello"))
	http.HandleFunc("/", handler("home page"))
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

	var url = "animalaid.org"

	var keyDir = NewPathM("webserv/https_keys")
	var certPath = keyDir.JoinM(url + ".crt")
	var keyPath = keyDir.JoinM(url + ".key")

	Pr("Type:", INDENT, "curl -sL https://"+url+"/hello")

	// This handles xxx.org
	//
	http.HandleFunc("/", handler("home page"))

	// xxx.org/hello
	//
	http.HandleFunc("/hello/", handler("hello"))

	// xxx.org/hey???
	//
	http.HandleFunc("/hey/", handler("hey"))

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
