package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/gen/webservgen"
	"reflect"

	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Define an app with a single operation

type SampleOper struct {
	https  bool
	ticker *time.Ticker

	sessionMap *SessionMap
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
	oper.sessionMap = BuildSessionMap()
	var app = NewApp()
	app.SetName("WebServer")
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

func (oper *SampleOper) handle(w http.ResponseWriter, req *http.Request) {

	// Create a buffer to accumulate the response text

	sb := NewBasePrinter()

	sb.Pr("Request received at:", time.Now().Format(time.ANSIC), CR)
	sb.Pr("URI:", req.RequestURI, CR)

	// Determine what session this is, by examining cookies

	// Can a session be non-nil and its builder nil?

	//var session *SessionBuilder
	var session Session
	{

		cookies := req.Cookies()
		sb.Pr("Cookies received:", len(cookies), CR)

		for i, c := range cookies {
			sb.Pr("Cookie #", i, "name:", c.Name, CR)
			Pr("cookie", i, "name:", c.Name)
			if c.Name == "session" {
				sessionId := c.Value
				session = oper.sessionMap.FindSession(sessionId)
				Pr("session id:", sessionId, "non nil:", session != nil, "found:", INDENT, session)
				sb.Pr("sessionId:", sessionId, INDENT, "found:", INDENT, session)
			}
			sb.Cr()
			if session != nil {
				sb.Pr("found a non-nil session:", INDENT, session)
				// This is a non-nil pointer to null? wtf
				// https://codefibershq.com/blog/golang-why-nil-is-not-always-nil

				Pr("session type:", reflect.TypeOf(session))
				fmt.Println(session)
				st := session.Build()
				Pr("built:", st)
				Pr("Stringer session:", session.String())
				break
			}
		}

		// If no session was found, create one, and send a cookie
		if session == nil {
			sb.Pr("creating session")
			session = oper.sessionMap.CreateSession()

			cookie := &http.Cookie{
				Name:   "session",
				Value:  session.Id(),
				MaxAge: 1200, // 20 minutes
			}
			sb.Pr("...no session cookie found, created one:", INDENT, session, CR)
			http.SetCookie(w, cookie)
		}
	}
	w.Header().Set("Content-Type", "text/plain")

	w.Write([]byte(sb.String()))
}

func (oper *SampleOper) handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		oper.handle(w, req)
	}
}

// ------------------------------------------------------------------------------------

func (oper *SampleOper) doHttp() {
	http.HandleFunc("/", oper.handler())
	Pr("Type:", INDENT, "curl -sL http://localhost:8090/hello")
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// ------------------------------------------------------------------------------------

func (oper *SampleOper) doHttps() {

	var url = "animalaid.org"

	var keyDir = NewPathM("webserv/https_keys")
	var certPath = keyDir.JoinM(url + ".crt")
	var keyPath = keyDir.JoinM(url + ".key")

	Pr("Type:", INDENT, "curl -sL https://"+url+"/hello")

	http.HandleFunc("/", oper.handler())

	if false {
		oper.startTicker()
	}

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (oper *SampleOper) startTicker() {
	oper.ticker = time.NewTicker(5 * time.Second)

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-oper.ticker.C:
				oper.makeRequest()
			case <-quit:
				oper.ticker.Stop()
				oper.ticker = nil
				return
			}
		}
	}()

}

func (oper *SampleOper) makeRequest() {
	resp, err := http.Get("https://animalaid.org/hey/joe")
	if err != nil {
		log.Fatalln(err)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	Pr("client: response body:", INDENT, string(resBody))
}
