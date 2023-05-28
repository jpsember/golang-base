package webserv

import (
	"bufio"
	"fmt"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/gen/webservgen"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Define an app with a single operation

type SampleOper struct {
	https        bool
	ticker       *time.Ticker
	sessionMap   *SessionMap
	appRoot      Path
	resources    Path
	uploadedFile Path
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

	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

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

	resource := req.RequestURI[1:]
	if resource != "" {
		if resource == "upload" {
			oper.handleUpload(w, req, resource)
			return
		}
		oper.handleResourceRequest(w, req, resource)
		return
	}

	Todo("the Pr method is not thread safe")

	// Create a buffer to accumulate the response text

	sb := NewBasePrinter()

	sb.Pr(`
<HMTL>

<HEAD>
<TITLE>Example</TITLE>
</HEAD>

<BODY>
`)

	sb.Pr("Request received at:", time.Now().Format(time.ANSIC), CR)
	sb.Pr("URI:", req.RequestURI, CR)

	var session = oper.determineSession(w, req, true)

	sb.Pr("session:", session.Id())

	sb.Pr(`<p>Here is a picture: <img src=picture.jpg alt="Picture"></p>`, CR)

	if oper.uploadedFile != "" {
		sb.Pr(`<p>Here is a recently uploaded image: <img src=recent.jpg></p>`, CR)
	}
	sb.Pr(`<p>Click on the "Choose File" button to upload a file:</p>

<form action="upload" enctype="multipart/form-data" method="post">
    <input type="file" name="file" id="file" />
    <input type="submit" />
</form>

`)
	sb.Pr(`
</BODY>
`)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(sb.String()))
}

func (oper *SampleOper) sendResponseMarkup(w http.ResponseWriter, req *http.Request, content string) {
	sb := NewBasePrinter()

	sb.Pr(`
<HMTL>

<HEAD>
<TITLE>Example</TITLE>
</HEAD>

<BODY>
`)
	sb.Pr("<p>")
	sb.Pr(content)
	sb.Pr("</p>")
	sb.Pr(`</BODY>`)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(sb.String()))
}

func (oper *SampleOper) determineSession(w http.ResponseWriter, req *http.Request, createIfNone bool) *SessionBuilder {
	// Determine what session this is, by examining cookies
	var session *SessionBuilder
	cookies := req.Cookies()
	for _, c := range cookies {
		if c.Name == "session" {
			sessionId := c.Value
			session = oper.sessionMap.FindSession(sessionId)
		}
		if session != nil {
			break
		}
	}

	// If no session was found, create one, and send a cookie
	if session == nil && createIfNone {
		session = oper.sessionMap.CreateSession()
		cookie := &http.Cookie{
			Name:   "session",
			Value:  session.Id(),
			MaxAge: 1200, // 20 minutes
		}
		http.SetCookie(w, cookie)
	}
	return session
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

	var keyDir = oper.appRoot.JoinM("https_keys")
	var certPath = keyDir.JoinM(url + ".crt")
	var keyPath = keyDir.JoinM(url + ".key")

	Pr("Type:", INDENT, "curl -sL https://"+url)

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

func (oper *SampleOper) handleResourceRequest(w http.ResponseWriter, req *http.Request, resource string) {
	if resource == "picture.jpg" {
		picPath := oper.resources.JoinM("picture.jpg")
		content := picPath.ReadBytesM()
		Todo("when does caching come into effect?  Is that a browser thing?")
		w.Header().Set("Content-Type", "image/jpeg ")
		w.Write(content)
		return
	}
	if resource == "recent.jpg" && oper.uploadedFile.NonEmpty() {
		picPath := oper.uploadedFile
		content := picPath.ReadBytesM()
		Todo("just assuming we're sending a jpeg")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(content)
		return
	}
}

func (oper *SampleOper) handleUpload(w http.ResponseWriter, r *http.Request, resource string) {

	// If there is no session, do nothing
	var session = oper.determineSession(w, r, false)
	if session == nil {
		oper.sendResponseMarkup(w, r, "no session, sorry")
		return
	}

	// Relevant: https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad

	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024)
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := reader.NextPart()
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bufio.NewReader(p)
	sniff, _ := buf.Peek(512)
	contentType := http.DetectContentType(sniff)
	Pr("contentType:", contentType)
	if contentType != "image/jpeg" {
		http.Error(w, "file type not allowed", http.StatusBadRequest)
		return
	}

	Todo("not defering closing the file, since we want to copy it immediately")
	f, err := os.CreateTemp("", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))
	written, err := io.Copy(f, lmt)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if written > maxSize {
		os.Remove(f.Name())
		http.Error(w, "file size over limit", http.StatusBadRequest)
		return
	}
	f.Close()

	oldLocation := f.Name()
	newPath := oper.appRoot.JoinM("uploaded/recent.jpg")
	if newPath.Exists() {
		newPath.DeleteFileM()
	}

	err = os.Rename(oldLocation, newPath.String())
	CheckOk(err)

	oper.uploadedFile = newPath
	oper.sendResponseMarkup(w, r, "Successfully uploaded: "+newPath.String())
}
