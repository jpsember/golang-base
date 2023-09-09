package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

const (
	id_animal_name    = "a_name"
	id_animal_summary = "a_summary"
	id_animal_details = "a_details"
	id_add            = "a_add"
	id_animal_photo   = "a_photo"
)

type CreateAnimalPageStruct struct {
	BasicPage
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) AbstractPage {
	t := &CreateAnimalPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "create_animal_page"
	return t
}

func (p CreateAnimalPage) Generate() {
	//SetWidgetDebugRendering()

	m := p.GenerateHeader()

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")
	m.Label("Create New Animal Record").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("Name").Id(id_animal_name).Listener(ValidateAnimalName).AddInput()

		m.Label("Summary").Id(id_animal_summary).AddInput()
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		m.Label("Details").Id(id_animal_details).AddInput()
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()

		m.Listener(p.addListener)
		m.Id(id_add).Label("Create").AddButton()
	}
	m.Close()

	m.Open()
	m.Id(id_animal_photo).Label("Photo").AddFileUpload()
	m.Close()
}

func (p CreateAnimalPage) addListener(sess Session, widget Widget) {
	if Todo("CreateAnimal") {

	}
}

func ValidateAnimalName(s Session, widget Widget) {
	errStr := ""
	n := s.GetValueString()
	n = strings.TrimSpace(n)
	for {
		ln := len(n)
		if ln < 3 || ln > 20 {
			errStr = "Length should be 3...20 characters"
			break
		}
		break
	}
	if errStr != "" {
		s.SetWidgetProblem(widget, errStr)
	}
}

//// Define an app with a single operation
//
//type SampleOper struct {
//	//https        bool
//	//ticker       *time.Ticker
//	//sessionMap   *SessionMap
//	//appRoot      Path
//	//resources    Path
//	//uploadedFile Path
//}
//
//func (oper *SampleOper) handle(w http.ResponseWriter, req *http.Request) {
//
//	resource := req.RequestURI[1:]
//	if resource != "" {
//		if resource == "upload" {
//			oper.handleUpload(w, req, resource)
//			return
//		}
//		oper.handleResourceRequest(w, req, resource)
//		return
//	}
//
//	Todo("the Pr method is not thread safe")
//
//	// Create a buffer to accumulate the response text
//
//	sb := NewBasePrinter()
//
//	sb.Pr(`
//<HMTL>
//
//<HEAD>
//<TITLE>Example</TITLE>
//</HEAD>
//
//<BODY>
//`)
//
//	sb.Pr("Request received at:", time.Now().Format(time.ANSIC), CR)
//	sb.Pr("URI:", req.RequestURI, CR)
//
//	var session = oper.determineSession(w, req, true)
//
//	sb.Pr("session:", session.Id())
//
//	sb.Pr(`<p>Here is a picture: <img src=picture.jpg alt="Picture"></p>`, CR)
//
//	//if oper.uploadedFile != "" {
//	//	sb.Pr(`<p>Here is a recently uploaded image: <img src=recent.jpg></p>`, CR)
//	//}
//	sb.Pr(`<p>Click on the "Choose File" button to upload a file:</p>
//
//<form action="upload" enctype="multipart/form-data" method="post">
//    <input type="file" name="file" id="file" />
//    <input type="submit" />
//</form>
//
//`)
//	sb.Pr(`
//</BODY>
//`)
//
//	w.Header().Set("Content-Type", "text/html")
//
//	w.Write([]byte(sb.String()))
//}
//
//func (oper *SampleOper) handler() func(http.ResponseWriter, *http.Request) {
//	return func(w http.ResponseWriter, req *http.Request) {
//		oper.handle(w, req)
//	}
//}
//
//// ------------------------------------------------------------------------------------
//
//func (oper *SampleOper) doHttp() {
//	http.HandleFunc("/", oper.handler())
//	Pr("Type:", INDENT, "curl -sL http://localhost:8090/hello")
//	err := http.ListenAndServe(":8090", nil)
//	if err != nil {
//		log.Fatal("ListenAndServe: ", err)
//	}
//}
//
//// ------------------------------------------------------------------------------------
//
//func (oper *SampleOper) makeRequest() {
//	resp, err := http.Get("https://animalaid.org/hey/joe")
//	if err != nil {
//		log.Fatalln(err)
//	}
//	resBody, err := io.ReadAll(resp.Body)
//	if err != nil {
//		fmt.Printf("client: could not read response body: %s\n", err)
//		os.Exit(1)
//	}
//	Pr("client: response body:", INDENT, string(resBody))
//}
//
//func (oper *SampleOper) handleUpload(w http.ResponseWriter, r *http.Request, resource string) {
//
//	// If there is no session, do nothing
//	var session = oper.determineSession(w, r, false)
//	if session == nil {
//		oper.sendResponseMarkup(w, r, "no session, sorry")
//		return
//	}
//
//	// Relevant: https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
//
//	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024)
//	reader, err := r.MultipartReader()
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	p, err := reader.NextPart()
//	if err != nil && err != io.EOF {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	buf := bufio.NewReader(p)
//	sniff, _ := buf.Peek(512)
//	contentType := http.DetectContentType(sniff)
//	Pr("contentType:", contentType)
//	if contentType != "image/jpeg" {
//		http.Error(w, "file type not allowed", http.StatusBadRequest)
//		return
//	}
//
//	Todo("not defering closing the file, since we want to copy it immediately")
//	f, err := os.CreateTemp("", "")
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	var maxSize int64 = 32 << 20
//	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))
//	written, err := io.Copy(f, lmt)
//	if err != nil && err != io.EOF {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	if written > maxSize {
//		os.Remove(f.Name())
//		http.Error(w, "file size over limit", http.StatusBadRequest)
//		return
//	}
//	f.Close()
//
//	oldLocation := f.Name()
//	newPath := oper.appRoot.JoinM("uploaded/recent.jpg")
//	if newPath.Exists() {
//		newPath.DeleteFileM()
//	}
//
//	err = os.Rename(oldLocation, newPath.String())
//	CheckOk(err)
//
//	oper.uploadedFile = newPath
//	oper.sendResponseMarkup(w, r, "Successfully uploaded: "+newPath.String())
//}
