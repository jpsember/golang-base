package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ServerApp interface {
	PageRequesterInterface
	PrepareSession(s Session)
	Resources() Path // Directory where resources to be served to client can be found, including header.html
	PageTemplates() []Page
}

// Why does leaving the name of the arg off (s) screw things up?
type PathHandler func(s Session, remainingPath string)

type JServerStruct struct {
	App            ServerApp
	FullWidth      bool
	BaseURL        string // e.g. "jeff.org"
	KeyDir         Path
	SessionManager SessionManager
	BlobCache      BlobCache
	resources      Path
	pageRequester  PageRequester
	TopPadding     int
	headerMarkup   string
	handlerMap     map[string]PathHandler
	started        bool
}

type JServer = *JServerStruct

var singletonJServer JServer

func NewJServer(app ServerApp) JServer {
	CheckState(singletonJServer == nil, "singleton server already constructed")

	t := &JServerStruct{
		App:           app,
		pageRequester: NewPageRequester(app),
		handlerMap:    make(map[string]PathHandler),
	}
	singletonJServer = t
	t.resources = app.Resources().AssertNonEmpty()
	t.registerPages()
	t.headerMarkup = t.resources.JoinM("header.html").ReadStringM()

	// Every several runs, remind to discard tabs
	{
		k := ProjectDirM().JoinM("._SKIP_counter")
		m := JSMapFromFileIfExistsM(k)
		count := m.OptInt("", 0) + 1
		m.Put("", count)
		k.WriteStringM(m.CompactString())
		if count >= 10 {
			k.DeleteFileM()
			Pr(VERT_SP, DASHES, CR, "Take a moment and discard all the tabs")
			SleepMs(4000)
		}
	}
	return t
}

func (j JServer) registerPages() {
	pgs := j.App.PageTemplates()
	for _, pg := range pgs {
		j.pageRequester.RegisterPage(pg)
	}
}

const BlobURLPrefix = "~/"

func DebVerifyServerStarted() {
	if singletonJServer == nil || !singletonJServer.started {
		BadState("JServer has not been started yet")
	}
}

func (j JServer) StartServing() {

	CheckState(!j.started, "server has already started")

	if j.BlobCache != nil {
		j.AddResourceHandler(BlobURLPrefix, j.handleBlobRequest)
	}

	j.started = true

	var ourUrl = CheckNonEmpty(j.BaseURL, "BaseURL")

	var keyDir = j.KeyDir //oper.appRoot.JoinM("https_keys")
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
			j.handle(w, req)
		})

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func (j JServer) handleBlobRequest(s Session, blobId string) {
	blob := j.BlobCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}
	err := WriteResponse(s.ResponseWriter, InferContentTypeFromBlob(blob.Data()), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	ReportIfError(err, "Trouble writing blob response")
}

// A handler such as this must be thread safe!
func (j JServer) handle(w http.ResponseWriter, req *http.Request) {
	pr := PrIf("JServer.handle", false)
	pr("JServer handler, request:", req.RequestURI)

	// We don't know what the session is yet, so we don't have a lock on it...
	sess := DetermineSession(j.SessionManager, w, req, true)
	// It seems awkward to initialize this at session construction time, so set it here...
	sess.app = j.App

	// Now that we have the session, lock it
	sess.lock.Lock()
	defer sess.ReleaseLockAndDiscardRequest()

	sess.PrepareForHandlingRequest(w, req)
	pr("session id:", sess.Id, "page widget:", sess.PageWidget)

	// If session hasn't been prepared yet, do so.
	if sess.PageWidget == nil {
		pr(VERT_SP, "constructing page widget")

		// Open a container for the entire page
		sess.PageWidget = sess.RebuildPageWidget()
		j.App.PrepareSession(sess)
	}

	url, err := url.Parse(req.RequestURI)
	if err == nil {

		path := url.Path
		if !strings.HasPrefix(path, "/") {
			Alert("#50path didn't have expected prefix:", VERT_SP, Quoted(path), VERT_SP)
		} else {
			path = strings.TrimPrefix(path, "/")
		}
		var text string
		var flag bool

		pr("JServer, url path:", path)
		if path == "ajax" {
			sess.HandleAjaxRequest()
		} else if text, flag = TrimIfPrefix(path, `upload/`); flag {
			pr("handling upload request with:", text)
			sess.HandleUploadRequest(text)
		} else {

			result := false
			for key, handler := range j.handlerMap {
				if text, flag = TrimIfPrefix(path, key); flag {
					handler(sess, text)
					result = true
					break
				}
			}

			if !result {
				result = j.processPageRequest(sess, path)
			}

			if !result {
				// If we fail to parse any requests, assume it's a resource, like that stupid favicon
				pr("JServer handling resource request for:", path)
				err = sess.HandleResourceRequest(j.resources)
				if err != nil {
					Todo("Issue a 404")
					Alert("<1#50Cannot handle request:", Quoted(path))
				}
			}
		}
	}

	if err != nil {
		sess.SetRequestProblem(err)
	}

	if p := sess.GetRequestProblem(); p != nil {
		Pr("jserver:...problem with request, URL:", req.RequestURI, INDENT, p)
	}
}

// Generate the boilerplate header and scripts markup
func (j JServer) writeHeader(bp MarkupBuilder) {
	bp.A(j.headerMarkup)
	bp.TgOpen(`body`).TgContent()
	containerClass := "container"
	if j.FullWidth {
		containerClass = "container-fluid"
	}
	if false && j.TopPadding != 0 {
		containerClass += "  pt-" + IntToString(j.TopPadding)
	}
	bp.Comments("page container").TgOpen(`div class=`).A(QUO, containerClass).TgContent()
}

func (j JServer) sendFullPage(sess Session) {
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sb := NewMarkupBuilder()
	j.writeHeader(sb)
	RenderWidget(sess.PageWidget, sess, sb)
	sess.RequestClientInfo(sb)
	j.writeFooter(sess, sb)
	WriteResponse(sess.ResponseWriter, "text/html", sb.Bytes())
}

// Generate the boilerplate footer markup, then write the page to the response
func (j JServer) writeFooter(s Session, bp MarkupBuilder) {
	bp.TgClose() // page container

	// Add a bit of javascript that will change the url to what we want
	expr := s.NewBrowserPath()
	if expr != "" {
		code := `
<script type="text/javascript">
var url = location.origin+'` + expr + `'
history.replaceState(null, null, url)
</script>
`
		// ^^^I suspect we don't want to do pushState if we got here due to user pressing the back button.
		bp.WriteString(code)
	}
	bp.TgClose() // body

	bp.A(`</html>`).Cr()
}

// Strings that start with {zero or more lowercase letters}/
var looksLikePageRexEx = CheckOkWith(regexp.Compile(`^[a-z]*\/`))

// Parse URL requested by client, and serve up an appropriate page.
func (j JServer) processPageRequest(s Session, path string) bool {

	// If path is NOT one of "", "pagename", or "pagename[non-alpha]..., exit with false immediately
	{
		// Add a trailing / to make this logic simpler
		modifiedPath := path + `/`
		if !looksLikePageRexEx.MatchString(modifiedPath) {
			return false
		}
		// Extract page name
		i := strings.IndexByte(modifiedPath, '/')
		pageName := modifiedPath[0:i]

		// If what remains is a nonempty string that isn't the name of a page, exit
		if pageName != "" && j.pageRequester.PageWithName(pageName) == nil {
			return false
		}
	}

	j.pageRequester.Process(s, path)
	j.sendFullPage(s)
	return true
}

func (j JServer) AddResourceHandler(pathPrefix string, handler PathHandler) {
	CheckState(!j.started, "server has already started")
	CheckState(!HasKey(j.handlerMap, pathPrefix), "duplicate handler for prefix:", pathPrefix)
	j.handlerMap[pathPrefix] = handler
}
