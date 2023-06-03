package webserv

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type AjaxOperStruct struct {
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	headerMarkup   string
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

	var url = "animalaid.org"

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
	bp.OpenHtml(`div class="container"`, "body")
}

// Generate the boilerplate footer markup, then write the page to the response
func (oper AjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseHtml("div", "body")
	bp.Br().CloseHtml("body", "")
	bp.A(`</html>`).Cr()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(bp.String()))
}

const WidgetIdPage = "page"

// Assign a widget heirarchy to a session
func (oper AjaxOper) constructPageWidget(sess Session) {
	m := NewWidgetManager()
	//m.AlertVerbose()

	// Page occupies full 12 columns
	m.Col(12)
	widget := m.openFor(WidgetIdPage, "main container")

	m.Col(4)

	m.Listener(birdListener)
	m.Col(6)
	m.AddText("bird")
	m.Col(3)
	m.AddLabel("x52")
	m.AddLabel("x53")

	m.Col(2)
	m.AddLabel("x54")
	m.Col(4)
	m.Listener(zebraListener)
	m.AddText("zebra")
	m.Col(2)
	m.AddLabel("x57")
	m.AddLabel("x58")
	m.AddLabel("x59")

	m.close()

	sess.PageWidget = widget
	sess.WidgetMap = m.widgetMap
}

func birdListener(sess any, widget Widget) {
	Todo("can we have sessions produce listener functions with appropriate handling of sess any?")
	s := sess.(Session)

	newVal := s.GetValueString()
	if !s.Ok() {
		return
	}
	s.State.Put(widget.GetId(), newVal+"<<added for fun")
	Pr("state map now:", INDENT, s.State)
	Pr("repainting widget")
	s.Repaint(widget.GetId())
}

func zebraListener(sess any, widget Widget) {
	s := sess.(Session)
	newVal := s.GetValueString()
	if !s.Ok() {
		return
	}
	s.State.Put(widget.GetId(), newVal)
	s.Repaint(widget.GetId())
}