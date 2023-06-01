package webserv

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"log"
	"net/http"
	"strings"
	"time"
)

type AjaxOper struct {
	sessioinManager SessionManager
	appRoot         Path
	resources       Path
	headerMarkup    string
}

func (oper *AjaxOper) UserCommand() string {
	return "widgets"
}

func (oper *AjaxOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server with AJAX manipulating Widget UI elements")
}

func (oper *AjaxOper) ProcessArgs(c *CmdLineArgs) {
}

func (oper *AjaxOper) Perform(app *App) {

	oper.sessioinManager = BuildFileSystemSessionMap()
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

	http.HandleFunc("/", oper.handler())

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

// A handler such as this must be thread safe!
func (oper *AjaxOper) handle(w http.ResponseWriter, req *http.Request) {

	// These are a pain in the ass
	if req.RequestURI == "/favicon.ico" {
		return
	}

	Pr("handler, request:", req.RequestURI)

	resource := req.RequestURI[1:]

	// If the request is "ajax?...", it is an Ajax request.
	if strings.HasPrefix(resource, "ajax?") {
		sess := DetermineSession(oper.sessioinManager, w, req, false)
		// Ignore if there is no session
		if sess == nil {
			return
		}
		oper.sendAjaxMarkup(w, req)
		return
	}

	// Otherwise, assume a full page refresh
	oper.processFullPageRequest(w, req)
}

func (oper *AjaxOper) handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		oper.handle(w, req)
	}
}

func (oper *AjaxOper) sendAjaxMarkup(w http.ResponseWriter, req *http.Request) {
	sb := NewBasePrinter()
	sb.Pr(`<h3> This was changed via an AJAX call without using JQuery at ` +
		time.Now().Format(time.ANSIC) + `</h3>`)
	Pr("sending markup back to Ajax caller:", INDENT, sb.String())
	w.Write([]byte(sb.String()))
}

func (oper *AjaxOper) processFullPageRequest(w http.ResponseWriter, req *http.Request) {
	// Construct a session if none found, and a widget for a full webpage
	sess := DetermineSession(oper.sessioinManager, w, req, true)
	if sess.PageWidget == nil {
		oper.constructPageWidget(sess)
	}
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb)
	oper.writeFooter(w, sb)
}

// Generate the biolerplate header and scripts markup
func (oper *AjaxOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.headerMarkup)
	bp.OpenHtml("body", "").Br()
	bp.OpenHtml(`div class="container-fluid"`, "body")
}

// Generate the boilerplate footer markup, then write the page to the response
func (oper *AjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseHtml("div", "body")
	bp.Br().CloseHtml("body", "")
	bp.A(`</html>`).Cr()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(bp.String()))
}

// Assign a widget heirarchy to a session
func (oper *AjaxOper) constructPageWidget(sess Session) {
	m := NewWidgetManager()
	m.SetVerbose(true)

	m.Columns("..x")
	widget := m.openFor("main container")
	m.AddLabel("x51")
	m.AddLabel("x52")
	m.AddLabel("x53")

	m.AddLabel("x54")
	m.AddText("zebra")
	m.AddLabel("x56")

	m.close()

	sess.PageWidget = widget
}
