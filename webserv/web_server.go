package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Define an app with a single operation

type SampleOper struct {
	https  bool
	ticker *time.Ticker

	sessionMap  map[string]*Session
	sessionLock sync.RWMutex

	uniqueSessionId atomic.Int64
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
	oper.sessionMap = make(map[string]*Session)
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

func (oper *SampleOper) handle(w http.ResponseWriter, req *http.Request, msg string) {
	// Determine what session this is, by examining cookies
	var session *Session
	{

		cookies := req.Cookies()
		Pr("received", len(cookies), "cookies")
		for _, c := range cookies {
			Pr("Cookie:", c)
			if c.Name == "session" {
				sessionId := c.Value
				session = oper.findSession(sessionId)
				Pr("session id in cookie is", sessionId, "finding:", session)
			}
		}

		// If no session was found, create one, and send a cookie
		if session == nil {
			session = oper.createSession()

			cookie := &http.Cookie{
				Name:   "session",
				Value:  session.Id,
				MaxAge: 1200, // 20 minutes
			}
			Pr("created session with id", session.Id, "and storing in response")
			http.SetCookie(w, cookie)
		}

	}
	w.Header().Set("Content-Type", "text/plain")

	s := ToString(time.Now().Format(time.ANSIC)+":", msg, CR, "Request URI:", req.RequestURI, CR, "Session:", session)
	w.Write([]byte(s))
}

type Session struct {
	Id string
}

func (session *Session) String() string {
	return session.Id
}

func (oper *SampleOper) findSession(key string) *Session {
	oper.sessionLock.RLock()
	session := oper.sessionMap[key]
	oper.sessionLock.RUnlock()
	return session
}

func (oper *SampleOper) createSession() *Session {
	oper.sessionLock.Lock()
	ourId := oper.uniqueSessionId.Add(1)
	session := new(Session)
	session.Id = fmt.Sprintf("%v", ourId)
	oper.sessionMap[session.Id] = session
	oper.sessionLock.Unlock()
	return session
}

func (oper *SampleOper) handler(msg ...any) func(http.ResponseWriter, *http.Request) {
	var m = ToString(msg...)
	return func(w http.ResponseWriter, req *http.Request) {
		oper.handle(w, req, m)

	}
}

// ------------------------------------------------------------------------------------

func (oper *SampleOper) doHttp() {
	Todo("I don't think I need separate handlers for secure vs non")
	http.HandleFunc("/hello", oper.handler("hello"))
	http.HandleFunc("/", oper.handler("home page"))
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
	http.HandleFunc("/", oper.handler("home page"))

	// xxx.org/hello
	//
	http.HandleFunc("/hello/", oper.handler("hello"))

	// xxx.org/hey???
	//
	http.HandleFunc("/hey/", oper.handler("hey"))

	oper.startTicker()

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
