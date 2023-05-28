package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/gen/webservgen"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Define an app with a single operation

type SampleOper struct {
	https      bool
	ticker     *time.Ticker
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

	resource := req.RequestURI[1:]
	if resource != "" {
		if strings.HasPrefix(resource, "upload?") {
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

	var session = oper.determineSession(w, req)

	sb.Pr("session:", session.Id())

	sb.Pr(`<p>Here is a picture: <img src=picture.jpg alt="Picture"></p>`, CR)

	sb.Pr(`<p>Click on the "Choose File" button to upload a file:</p>

<form action="upload">
  <input type="file" id="myFile" name="filename">
  <input type="submit">
</form>
`)

	sb.Pr(`
</BODY>
`)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(sb.String()))
}

func (oper *SampleOper) determineSession(w http.ResponseWriter, req *http.Request) *SessionBuilder {
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
	if session == nil {
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

	var keyDir = NewPathM("webserv/https_keys")
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
	Pr("resource:", resource)
	if resource == "picture.jpg" {
		pth, err := AscendToDirectoryContainingFile("", "go.mod")
		CheckOk(err)
		resourcesPath := pth.JoinM("webserv/resources")

		picPath := resourcesPath.JoinM("picture.jpg")
		content := picPath.ReadBytesM()

		Todo("when does caching come into effect?  Is that a browser thing?")

		Pr("sending image")
		w.Header().Set("Content-Type", "image/jpeg ")
		w.Write(content)
	}
}

func (oper *SampleOper) handleUpload(w http.ResponseWriter, r *http.Request, resource string) {

	Pr("request:", INDENT, r)

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	err := r.ParseMultipartForm(10 << 20)
	Pr("parse multipart form produced:", err)

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("filename")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	Pr("read:", len(fileBytes), "bytes")
	//// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	//// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")

}
