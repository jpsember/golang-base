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
	sessionMap   SessionManager
	appRoot      Path
	resources    Path
	headerMarkup string
}

func (oper *AjaxOper) UserCommand() string {
	return "sample"
}

func (oper *AjaxOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server")
}

func (oper *AjaxOper) ProcessArgs(c *CmdLineArgs) {
}

func (oper *AjaxOper) Perform(app *App) {

	oper.sessionMap = BuildFileSystemSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	{
		s := strings.Builder{}
		s.WriteString(oper.resources.JoinM("header.html").ReadStringM())
		s.WriteString(oper.resources.JoinM("base.js").ReadStringM())
		s.WriteString(`
</script>                                                                      +.                                                                               
</head>                                                                        +.                                                                               
`)
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

func (oper *AjaxOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.headerMarkup)
	bp.OpenHtml("body", "").Br()
	bp.OpenHtml(`div class="container-fluid"`, "body")
}

// Write footer markup, then write the page to the response
func (oper *AjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseHtml("div", "body")
	bp.Br().CloseHtml("body", "")
	bp.A(`</html>`).Cr()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(bp.String()))
}

// A handler such as this must be thread safe!
func (oper *AjaxOper) handle(w http.ResponseWriter, req *http.Request) {

	// These are a pain in the ass
	if req.RequestURI == "/favicon.ico" {
		return
	}

	Pr("handler, request:", req.RequestURI)

	resource := req.RequestURI[1:]

	if resource == "" {
		oper.processFullPageRequest(w, req)
		return
	}

	if strings.HasPrefix(resource, "ajax?") {
		sess := oper.determineSession(w, req, false)
		if sess != nil {
			oper.sendAjaxMarkup(w, req)
		}
		return
	}
}

func (oper *AjaxOper) determineSession(w http.ResponseWriter, req *http.Request, createIfNone bool) Session {

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
			Value:  session.Id,
			MaxAge: 1200, // 20 minutes
		}
		http.SetCookie(w, cookie)
	}
	return session
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
	sb := NewMarkupBuilder()
	sess := oper.determineSession(w, req, true)
	if sess.PageWidget == nil {
		oper.constructPageWidget(sess)
	}
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb)
	oper.writeFooter(w, sb)
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
