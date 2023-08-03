package main

import (
	"bufio"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var app = NewApp()
	app.SetName("nonajax_demo")
	app.Version = "1.0"
	app.CmdLineArgs().Add("insecure").Desc("insecure (http) mode")
	app.RegisterOper(&NonAjaxOperStruct{})
	app.Start()
}

// Define an app with a single operation

type NonAjaxOperStruct struct {
	insecure       bool
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	uploadedFile   Path
	headerMarkup   string
}

type NonAjaxOper = *NonAjaxOperStruct

func (oper NonAjaxOper) UserCommand() string {
	return "sample"
}

func (oper NonAjaxOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server")
}

func (oper NonAjaxOper) ProcessArgs(c *CmdLineArgs) {
}

func (oper NonAjaxOper) Perform(app *App) {
	oper.sessionManager = BuildFileSystemSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")
	var insecure = app.CmdLineArgs().Get("insecure")
	if insecure {
		oper.doHttp()
	} else {
		oper.doHttps()
	}
}

func (oper NonAjaxOper) getHeaderMarkup() string {
	if oper.headerMarkup == "" {
		s := strings.Builder{}
		s.WriteString(oper.resources.JoinM("header.html").ReadStringM())
		s.WriteString(oper.resources.JoinM("base.js").ReadStringM())
		s.WriteString(`
</script>                                                                      +.                                                                               
</head>                                                                        +.                                                                               
`)
		oper.headerMarkup = s.String()
	}
	return oper.headerMarkup
}

func (oper NonAjaxOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.getHeaderMarkup())
	bp.OpenHtml("body", "").Br()
	bp.OpenHtml(`div class="container-fluid"`, "body")
}

// Write footer markup, then write the page to the response
func (oper NonAjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseHtml("div", "body")
	bp.Br().CloseHtml("body", "")
	bp.A(`</html>`).Cr()
	// See https://stackoverflow.com/questions/64294267/
	WriteResponse(w, "text/html", []byte(bp.String()))
}

// A handler such as this must be thread safe!
func (oper NonAjaxOper) handle(w http.ResponseWriter, req *http.Request) {

	// These are a pain in the ass
	if req.RequestURI == "/favicon.ico" {
		return
	}

	Pr("handler, request:", req.RequestURI)

	resource := req.RequestURI[1:]

	if resource != "" {

		if resource == "view" {
			oper.processViewRequest(w, req)
		}

		if resource == "robot" {
			oper.sendResponseMarkup(w, req, "Hi, Robot")
			return
		}

		if resource == "upload" {
			oper.handleUpload(w, req, resource)
			return
		}
		oper.handleResourceRequest(w, req, resource)
		return
	}

	// Create a buffer to accumulate the response text

	sb := NewMarkupBuilder()
	oper.writeHeader(sb)

	sb.Pr("Request received at:", time.Now().Format(time.ANSIC), CR)
	sb.Pr("URI:", req.RequestURI, CR)

	var session = DetermineSession(oper.sessionManager, w, req, true)

	sb.Pr("session:", session.Id)

	sb.A(`<p>Here is a picture: <img src=picture.jpg alt="Picture"></p>`)

	if oper.uploadedFile != "" {
		sb.Pr(`<p>Here is a recently uploaded image: <img src=recent.jpg></p>`, CR)
	}
	sb.A(`<p>Click on the "Choose File" button to upload a file:</p>

<form action="upload" enctype="multipart/form-data" method="post">
    <input type="file" name="file" id="file">
    <input type="submit">
</form>
`)

	sb.Pr(`<div id="div1"><h2>Let AJAX Change This Text</h2></div><button onclick="ajax('div1')">Get External Text</button>`)
	sb.Pr(`<div id="div2"><h2>Another independent element</h2></div><button onclick="ajax('div2')">Button 2</button>`)

	oper.writeFooter(w, sb)
}

// Send a simple web page back with a message
func (oper NonAjaxOper) sendResponseMarkup(w http.ResponseWriter, req *http.Request, content string) {
	sb := NewMarkupBuilder()

	oper.writeHeader(sb)

	sb.Pr("<p>")
	sb.Pr(content)
	sb.Pr("</p>")

	oper.writeFooter(w, sb)
	Todo("Have an HTML string class that handles escaping")
	w.Write([]byte(sb.String()))
}

func (oper NonAjaxOper) handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		oper.handle(w, req)
	}
}

// ------------------------------------------------------------------------------------

func (oper NonAjaxOper) doHttp() {
	http.HandleFunc("/", oper.handler())
	Pr("Type:", INDENT, "curl -sL http://localhost:8090/hello")
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// ------------------------------------------------------------------------------------

func (oper NonAjaxOper) doHttps() {

	var url = "jeff.org"

	var keyDir = oper.appRoot.JoinM("https_keys")
	var certPath = keyDir.JoinM(url + ".crt")
	var keyPath = keyDir.JoinM(url + ".key")

	Pr("curl -sL https://" + url)

	http.HandleFunc("/", oper.handler())

	if false {
		var robot = NewRobotRequester("https://animalaid.org/robot")
		robot.SetVerbose(true)
		robot.IntervalMS = 5000
		robot.Start()
	}

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (oper NonAjaxOper) handleResourceRequest(w http.ResponseWriter, req *http.Request, resource string) {
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

func (oper NonAjaxOper) handleUpload(w http.ResponseWriter, r *http.Request, resource string) {

	// If there is no session, do nothing
	var session = DetermineSession(oper.sessionManager, w, r, false)
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

	CheckOk(os.Rename(oldLocation, newPath.String()))

	oper.uploadedFile = newPath
	oper.sendResponseMarkup(w, r, "Successfully uploaded: "+newPath.String())
}

func (oper NonAjaxOper) sendAjaxMarkup(w http.ResponseWriter, req *http.Request) {
	sb := NewBasePrinter()
	sb.Pr(`<h3> This was changed via an AJAX call without using JQuery at ` +
		time.Now().Format(time.ANSIC) + `</h3>`)
	Pr("sending markup back to Ajax caller:", INDENT, sb.String())
	w.Write([]byte(sb.String()))
}

func (oper NonAjaxOper) processViewRequest(w http.ResponseWriter, req *http.Request) {
	sb := NewMarkupBuilder()
	sess := DetermineSession(oper.sessionManager, w, req, true)
	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb, sess.State)
	oper.writeFooter(w, sb)
}
