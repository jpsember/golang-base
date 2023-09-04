package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var AutoActivateUser = Alert("?Automatically activating user")

type AnimalOperStruct struct {
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	headerMarkup   string
	FullWidth      bool // If true, page occupies full width of screen
	TopPadding     int  // If nonzero, adds padding to top of page
}

type AnimalOper = *AnimalOperStruct

func (oper AnimalOper) UserCommand() string {
	return "widgets"
}

func (oper AnimalOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server with AJAX manipulating Widget UI elements")
}

func (oper AnimalOper) ProcessArgs(c *CmdLineArgs) {
}

func (oper AnimalOper) Perform(app *App) {
	//ClearAlertHistory()

	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())

	// Must 'close rows'?  See https://stackoverflow.com/questions/32479071

	if false && Alert("creating a number of users") {
		mr := NewJSRand().SetSeed(1965).Rand()
		for i := 0; i < 10; i++ {
			u := NewUser()
			u.SetName(RandomText(mr, 3, false))
			Pr("random name:", u.Name())
			Pr("i:", i, "attempting to create user with name:", u.Name())
			result, err := CreateUserWithName(u)
			Pr("create user result:", result.Id(), "err:", err)
			CheckOk(err)
			if result.Id() == 0 {
				Pr("failed to create user, must already exist?", u.Name())
				continue
			}
			Pr("created user:", result.Id(), result.Name())
		}
		Pr("sleeping then quitting")
		SleepMs(2000)
		//return
	}
	Todo("The indexes have the wrong name?")

	if true && Alert("experimenting with iter") {
		//sampleName := `mm`
		////sampleId := 122
		//Pr("Looking for", sampleName, ":", CheckOkWith(ReadUserWithName(sampleName)).Name())

		{
			iter := UserIterator(380)
			i := -1
			for iter.HasNext() {
				i++
				user := iter.Next().(User)
				CheckState(!iter.HasError())
				Pr("i:", i, "id:", user.Id(), "name:", user.Name())

			}
			Halt("done experiment")
		}

		DoIterExperiment(UserIterator(12))
		DoIterExperiment(UserIterator(380))

		//Pr("built an iterator for idMin 12:", INDENT, iter)
		//count := 0
		//for iter.HasNext() {
		//	result := iter.Next().(User)
		//	Pr("Result:", result.Id(), ":", result.Name(), "(count:", count, ")")
		//	if count%8 == 3 {
		//		Pr("Looking for", sampleName, "by id", sampleId, ":", CheckOkWith(ReadUser(sampleId)).Name())
		//	}
		//	//if result.Id() >= 80 {
		//	//	break
		//	//}
		//	count++
		//	if count > 200 {
		//		Pr("count too high")
		//		break
		//	}
		//}
		Halt("done iteration experiment")
	}

	if false && Alert("experiment") {
		b1 := NewBlob().SetName("bravo")
		b, err := CreateBlobWithName(b1)
		CheckOk(err)

		if b.Id() == 0 {
			Pr("failed to create blob:", INDENT, b1)
			Pr("assuming one already exists:")
			b2, err1 := ReadBlobWithName(b1.Name())
			Pr("err:", err1)
			Pr("Existing blob:", INDENT, b2)
		} else {
			var x []byte
			for i := 0; i < 2000; i++ {
				x = append(x, byte(i))
			}

			b = b.ToBuilder().SetData(x)
			Pr("Attempting to write blob:", INDENT, b)
			err2 := UpdateBlob(b)
			Pr("err?", err2)
		}
		Pr("sleeping to allow db flush")
		SleepMs(2000)
		Halt("done experiment")
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
			defer func() {
				if r := recover(); r != nil {
					BadState("<1Panic during http.HandleFunc:", r)
				}
			}()
			oper.handle(w, req)
		})

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func DoIterExperiment(iter DbIter) {

	sampleId := 122
	sampleName := `mm`

	Pr("Performing iter experiment with:", INDENT, iter)
	count := 0
	for iter.HasNext() {
		Todo("can we use generics to have this return a User?")
		result := iter.Next().(User)
		Pr("Result:", result.Id(), ":", result.Name(), "(count:", count, ")")
		if sampleId != 0 && count%8 == 3 {
			Pr("Looking for", sampleName, "by id", sampleId, ":", CheckOkWith(ReadUser(sampleId)).Name())
		}
		count++
		if count > 200 {
			Pr("reached max count")
			break
		}
	}
}

// A handler such as this must be thread safe!
func (oper AnimalOper) handle(w http.ResponseWriter, req *http.Request) {
	pr := PrIf(false)
	pr("handler, request:", req.RequestURI)

	if Alert("!If full page requested, discarding sessions") {
		url, err := url.Parse(req.RequestURI)
		if err == nil && url.Path == "/" {
			sess := DetermineSession(oper.sessionManager, w, req, false)
			if sess != nil {
				DiscardAllSessions(oper.sessionManager)
			}
		}
	}

	sess := DetermineSession(oper.sessionManager, w, req, true)
	if sess.AppData == nil {
		oper.AssignUserToSession(sess)
		oper.constructPageWidget(sess)

		user, ok := sess.AppData.(User)
		CheckState(ok, "no User found in sess AppData:", INDENT, sess.AppData)
		Todo("!have convention of prefixing enums with e.g. 'UserState_'")
		CheckState(
			user.Id() == 0)
		NewLandingPage(sess, sess.PageWidget).Generate()
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

func (oper AnimalOper) processFullPageRequest(sess Session, w http.ResponseWriter, req *http.Request) {
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
func (oper AnimalOper) writeHeader(bp MarkupBuilder) {
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
func (oper AnimalOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseTag() // page container
	bp.CloseTag() // body
	bp.A(`</html>`).Cr()
	WriteResponse(w, "text/html", bp.Bytes())
}

const WidgetIdPage = "main_page"

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

// Assign a widget heirarchy to a session
func (oper AnimalOper) constructPageWidget(sess Session) {
	m := sess.WidgetManager()
	//m.AlertVerbose()

	Todo("?Clarify when we need to *remove* old widgets")
	m.Id(WidgetIdPage)
	widget := m.Open()
	sess.PageWidget = widget
	m.Close()
}

// A new session was created; assign an 'unknown' user to it
func (oper AnimalOper) AssignUserToSession(sess Session) {
	sess.AppData = NewUser().Build()
}
