package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"log"
	"os"
	"runtime/debug"

	. "github.com/jpsember/golang-base/webserv"
	"net/http"
	"net/url"
	"strings"
)

type AnimalOperStruct struct {
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	headerMarkup   string
	FullWidth      bool // If true, page occupies full width of screen
	TopPadding     int  // If nonzero, adds padding to top of page
}
type AjaxOper = *AnimalOperStruct

func (oper AjaxOper) UserCommand() string {
	return "widgets"
}

func (oper AjaxOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server with AJAX manipulating Widget UI elements")
}

func (oper AjaxOper) ProcessArgs(c *CmdLineArgs) {
}

func ShowStackTrace() {
	//Pr("printing stack:")
	//debug.PrintStack()
	Pr("generating stack:")
	y := debug.Stack()
	x := string(y)
	lns := strings.Split(x, "\n")

	prefix := ""
	for _, val := range lns {
		result := val
		for {
			if strings.HasPrefix(val, "goroutine ") {
				result = ""
				break
			}

			if strings.HasPrefix(val, "\t") {
				val := strings.TrimSpace(val)
				cols := strings.Fields(val)
				if len(cols) != 2 {
					break
				}
				result = cols[0]
				break
			}
			j := strings.LastIndex(val, "(")
			if j < 0 {
				break
			}
			q := strings.LastIndex(val[0:j], ".")
			if q < 0 {
				break
			}
			prefix = val[q+1 : j]
			result = ""
			break
		}
		if result != "" {
			Pr(result + " " + prefix)
		}
	}
	//Pr(CurrentDirectory())
	Pr("animal_app.go:89")
	Pr("zero.go:89")

	//Pr(lns)
	os.Exit(1)
}

func (oper AjaxOper) Perform(app *App) {
	if false && Alert("Performing sql experiment") {
		SQLiteExperiment()
		return
	}

	db := CreateDatabase()
	db.SetDataSourceName("../sqlite/jeff_experiment.db")
	db.Open()

	if false && Alert("blob experiment") {
		PerformBlobExperiment(db)
		Halt()
	}

	oper.sessionManager = BuildSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	{
		s := strings.Builder{}
		s.WriteString(oper.resources.JoinM("header.html").ReadStringM())
		oper.headerMarkup = s.String()
	}

	var ourUrl = "jeff.org"

	var keyDir = oper.appRoot.JoinM("https_keys")
	var certPath = keyDir.JoinM(ourUrl + ".crt")
	var keyPath = keyDir.JoinM(ourUrl + ".key")
	Pr("URL:", INDENT, `https://`+ourUrl)

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
	pr("handler, request:", req.RequestURI)

	sess := DetermineSession(oper.sessionManager, w, req, true)
	if sess.AppData == nil {
		oper.AssignUserToSession(sess)
		oper.constructPageWidget(sess)
	}

	url, err := url.Parse(req.RequestURI)
	if err == nil {
		path := url.Path
		pr("url path:", path)
		if path == "/ajax" {
			sess.HandleAjaxRequest(w, req)
		} else if path == "/" {
			oper.processFullPageRequest(sess, w, req)
		} else {
			pr("handling resource request for:", path)
			err = sess.HandleResourceRequest(w, req, oper.resources)
		}
	}

	if err != nil {
		sess.SetRequestProblem(err)
	}

	if p := sess.GetRequestProblem(); p != "" {
		Pr("...problem with request, URL:", req.RequestURI, INDENT, p)
	}
}

func (oper AjaxOper) processFullPageRequest(sess Session, w http.ResponseWriter, req *http.Request) {
	// Construct a session if none found, and a widget for a full webpage
	//sess := DetermineSession(oper.sessionManager, w, req, true)
	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()

	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb, sess.State)
	sess.RequestClientInfo(sb)
	oper.writeFooter(w, sb)
}

// Generate the biolerplate header and scripts markup
func (oper AjaxOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.headerMarkup)
	bp.OpenTag("body")
	containerClass := "container"
	if oper.FullWidth {
		containerClass = "container-fluid"
	}
	if oper.TopPadding != 0 {
		containerClass += "  pt-" + IntToString(oper.TopPadding)
	}
	bp.Comments("page container").OpenTag(`div class='` + containerClass + `'`)
}

// Generate the boilerplate footer markup, then write the page to the response
func (oper AjaxOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseTag() // page container
	bp.CloseTag() // body
	bp.A(`</html>`).Cr()
	WriteResponse(w, "text/html", bp.Bytes())
}

const WidgetIdPage = "page"

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

// Assign a widget heirarchy to a session
func (oper AjaxOper) constructPageWidget(sess Session) {
	m := sess.WidgetManager()
	//m.AlertVerbose()

	Todo("?Clarify when we need to *remove* old widgets")
	m.Id(WidgetIdPage)
	widget := m.Open()
	sess.PageWidget = widget

	user, ok := sess.AppData.(webapp_data.User)
	CheckState(ok, "no User found in sess AppData:", INDENT, sess.AppData)

	Todo("!have convention of prefixing enums with e.g. 'UserState_'")
	if user.State() == webapp_data.UnknownUser {
		Pr("the user is unknown")
		CreateLandingPage(sess)
		return
	}

	alertWidget = NewAlertWidget("sample_alert", AlertInfo)
	alertWidget.SetVisible(false)
	m.Add(alertWidget)

	m.Size(SizeLarge).Text("This is the header text").AddHeading()

	heading := NewHeadingWidget("header_text", 1)
	m.Add(heading)

	m.Col(4)
	for i := 1; i < 12; i++ {
		anim, err := Db().GetAnimal(i)
		if err != nil {
			Pr("what do we do with unexpected errors?", INDENT, err)
		}
		if anim == nil {
			continue
		}
		cardId := "animal_" + IntToString(int(anim.Id()))
		OpenAnimalCardWidget(m, cardId, anim, buttonListener)
	}

	m.Col(4)
	m.Text("uniform delta").AddText()
	m.Col(8)
	m.Id("x58").Text(`X58`).Listener(buttonListener).AddButton().SetEnabled(false)

	m.Col(2).AddSpace()
	m.Col(3).AddSpace()
	m.Col(3).AddSpace()
	m.Col(4).AddSpace()

	m.Col(6)
	m.Listener(birdListener)
	m.Label("Bird").Id("bird")
	m.AddInput()

	m.Col(6)
	m.Open()
	m.Id("x59").Text(`Label for X59`).Listener(checkboxListener).AddCheckbox()
	m.Id("x60").Text(`With fruit`).Listener(checkboxListener).AddSwitch()
	m.Close()

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
	m.Label("Animal").Id("zebra").AddInput()

	m.Close()
}

// A new session was created; assign an 'unknown' user to it
func (oper AjaxOper) AssignUserToSession(sess Session) {
	sess.AppData = webapp_data.NewUser().Build()
}

func birdListener(sess any, widget Widget) {
	Todo("?can we have sessions produce listener functions with appropriate handling of sess any?")
	s := sess.(Session)
	newVal := s.GetValueString()
	if !s.Ok() {
		return
	}
	b := widget.GetBaseWidget()
	s.ClearWidgetProblem(widget)
	s.State.Put(b.Id, newVal)
	Todo("!do validation as a global function somewhere")
	if newVal == "parrot" {
		s.SetWidgetProblem(widget, "No parrots, please!")
	}
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
			RandomText(myRand.Rand(), 55, false)))
	s.Repaint(alertWidget)
}

func buttonListener(sess any, widget Widget) {
	s := sess.(Session)
	wid := s.GetWidgetId()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.Id,
		strings.TrimSpace(newVal))
	s.Repaint(alertWidget)
}

func checkboxListener(sess any, widget Widget) {
	s := sess.(Session)
	wid := s.GetWidgetId()

	// Get the requested new value for the widget
	newVal := s.GetValueBoolean()
	if !s.Ok() {
		return
	}

	s.State.Put(wid, newVal)
	// Repainting isn't necessary, as the web page has already done this
}
