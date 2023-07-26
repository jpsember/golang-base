package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	//ClearAlertHistory()
	SetDebugWidgetBounds()

	var app = NewApp()
	app.SetName("WebServer")
	app.Version = "1.0"
	app.CmdLineArgs().Add("insecure").Desc("insecure (http) mode")

	app.RegisterOper(&AjaxOperStruct{
		//FullWidth: true,
		TopPadding: 5,
	})
	app.Start()
}

type AjaxOperStruct struct {
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	headerMarkup   string
	FullWidth      bool // If true, page occupies full width of screen
	TopPadding     int  // If nonzero, adds padding to top of page
}
type AjaxOper = *AjaxOperStruct

func (oper AjaxOper) UserCommand() string {
	return "widgets"
}

func (oper AjaxOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server with AJAX manipulating Widget UI elements")
}

func (oper AjaxOper) ProcessArgs(c *CmdLineArgs) {
}

func (oper AjaxOper) Perform(app *App) {
	oper.sessionManager = BuildFileSystemSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	{
		s := strings.Builder{}
		s.WriteString(oper.resources.JoinM("header.html").ReadStringM())
		s.WriteString(oper.resources.JoinM("base.js").ReadStringM())
		s.WriteString("</script>\n</head>\n")
		oper.headerMarkup = s.String()
	}

	var url = "zebra.org"

	var keyDir = oper.appRoot.JoinM("https_keys")
	var certPath = keyDir.JoinM(url + ".crt")
	var keyPath = keyDir.JoinM(url + ".key")
	Pr("URL:", INDENT, `https://`+url)

	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			oper.handle(w, req)
		})

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

// A handler such as this must be thread safe!
func (oper AjaxOper) handle(w http.ResponseWriter, req *http.Request) {

	pr := PrIf(false)

	// These are a pain in the ass
	if req.RequestURI == "/favicon.ico" {
		return
	}

	pr("handler, request:", req.RequestURI)

	url, err := url.Parse(req.RequestURI)
	if err != nil {
		Pr("Error parsing RequestURI:", Quoted(req.RequestURI), INDENT, err)
		return
	}
	pr("url path:", url.Path)
	if url.Path == "/ajax" {
		sess := DetermineSession(oper.sessionManager, w, req, false)
		if sess != nil {
			sess.HandleAjaxRequest(w, req)
			return
		}
		// ...if no session, we fall back on a full page request
	}
	oper.processFullPageRequest(w, req)
}

func (oper AjaxOper) processFullPageRequest(w http.ResponseWriter, req *http.Request) {
	// Construct a session if none found, and a widget for a full webpage
	sess := DetermineSession(oper.sessionManager, w, req, true)
	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()
	// If this is a new session, store our operation within it
	if sess.AppData == nil {
		sess.AppData = oper
		sess.State.Put("header_text", "This is ajax_demo.go").
			Put("header_text_2", "8 columns").Put("header_text_3", "4 columns")
	}

	if sess.PageWidget == nil {
		oper.constructPageWidget(sess)
	}
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb, sess.State)
	oper.writeFooter(w, sb)
}

// Generate the biolerplate header and scripts markup
func (oper AjaxOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.headerMarkup)
	bp.OpenHtml("body", "").Br()
	containerClass := "container"
	if oper.FullWidth {
		containerClass = "container-fluid"
	}
	if oper.TopPadding != 0 {
		containerClass += "  pt-" + IntToString(oper.TopPadding)
	}
	bp.OpenHtml(`div class='`+containerClass+`'`, "page container")
}

// Generate the boilerplate footer markup, then write the page to the response
func (oper AjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseHtml("div", "page container")
	bp.Br().CloseHtml("body", "")
	bp.A(`</html>`).Cr()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(bp.String()))
}

const WidgetIdPage = "page"

var alertWidget AlertWidget
var myRand = rand.New(rand.NewSource(1234))

func GetOperFromSession(session Session) AjaxOper {
	return session.AppData.(AjaxOper)
}

// Assign a widget heirarchy to a session
func (oper AjaxOper) constructPageWidget(sess Session) {
	m := sess.WidgetManager()
	//m.AlertVerbose()

	// Page occupies full 12 columns
	m.Col(12)
	widget := m.OpenFor(WidgetIdPage, "main container")

	alertWidget = NewAlertWidget("sample_alert", AlertInfo)
	alertWidget.SetVisible(false)
	m.Add(alertWidget)

	heading := NewHeadingWidget("header_text", 1)
	m.Add(heading)

	m.Col(4)
	m.Text("uniform delta").AddText()
	m.Id("x58").Text(`X58`).Listener(buttonListener).AddButton().SetEnabled(false)
	m.Id("x59").Text(`X59`).Listener(buttonListener).AddButton()

	m.Col(8)
	m.Listener(birdListener)
	m.AddInput("bird")
	m.Col(4)
	m.Id("launch").Text(`Launch`).Listener(buttonListener).AddButton()

	m.Col(8)
	m.Text(`Sample text; is 5 < 26? A line feed
"Quoted string"
Multiple line feeds:


   an indented final line`)
	m.AddText()

	m.Col(4)
	m.Listener(zebraListener)
	m.AddInput("zebra")

	m.Close()

	sess.PageWidget = widget
}

func birdListener(sess any, widget Widget) {
	// Todo("can we have sessions produce listener functions with appropriate handling of sess any?")
	s := sess.(Session)
	newVal := s.GetValueString()
	if !s.Ok() {
		return
	}
	s.State.Put(widget.GetId(), newVal+" at "+time.Now().Format(time.ANSIC))
	//Pr("state map now:", INDENT, s.State)
	//Pr("repainting widget")
	s.Repaint(widget)
}

func zebraListener(sess any, widget Widget) {

	s := sess.(Session)

	// Get the requested new value for the widget
	newVal := s.GetValueString()
	if !s.Ok() {
		return
	}

	// Store this as the new value for this widget within the session state map
	s.State.Put(widget.GetBaseWidget().Id, newVal)
	s.Repaint(widget.GetBaseWidget())

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal

	alertWidget.SetVisible(newVal != "")

	s.State.Put(alertWidget.Id,
		strings.TrimSpace(newVal+" "+
			RandomText(myRand, 55, false)))
	s.Repaint(alertWidget)
}

func buttonListener(sess any, widget Widget) {
	s := sess.(Session)
	wid := s.GetWidgetId()
	Pr("Button click on id:", wid)
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.Id,
		strings.TrimSpace(newVal))
	s.Repaint(alertWidget)
}
