package webserv

import (
	"bufio"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Define an app with a single operation

type SampleOper struct {
	insecure     bool
	sessionMap   SessionManager
	appRoot      Path
	resources    Path
	uploadedFile Path
}

func (oper *SampleOper) UserCommand() string {
	return "sample"
}

func (oper *SampleOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server")
}

func (oper *SampleOper) ProcessArgs(c *CmdLineArgs) {
}

func WebServerDemo() {
	var oper = &SampleOper{}
	oper.sessionMap = BuildFileSystemSessionMap()

	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	var app = NewApp()
	app.SetName("WebServer")
	app.Version = "1.0"
	app.CmdLineArgs().Add("insecure").Desc("insecure (http) mode")
	app.RegisterOper(oper)
	//app.SetTestArgs("--insecure")
	app.Start()
}

func (oper *SampleOper) Perform(app *App) {
	var insecure = app.CmdLineArgs().Get("insecure")
	if insecure {
		oper.doHttp()
	} else {
		oper.doHttps()
	}
}

// A handler such as this must be thread safe!
func (oper *SampleOper) handle(w http.ResponseWriter, req *http.Request) {

	resource := req.RequestURI[1:]
	if resource != "" {

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

	sb := NewBasePrinter()

	sb.Pr(`
<!DOCTYPE html>
<html lang="en">
<head>
<title>Example</title>
<script src="https://code.jquery.com/jquery-1.12.4.min.js" integrity="sha256-ZosEbRLbNQzLpnKIkEdrPv7lOy9C27hHQ+Xp8a4MxAQ=" crossorigin="anonymous"></script>
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
</head>
<body>
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
    <input type="file" name="file" id="file">
    <input type="submit">
</form>

`)
	sb.Pr(`
</body></html>
`)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(sb.String()))
}

// Send a simple web page back with a message
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
	Todo("Have an HTML string class that handles escaping")
	w.Write([]byte(sb.String()))
}

func (oper *SampleOper) determineSession(w http.ResponseWriter, req *http.Request, createIfNone bool) Session {

	const sessionCookieName = "session_cookie"

	// Determine what session this is, by examining cookies
	var session Session
	cookies := req.Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName {
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
			Name:   sessionCookieName,
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
