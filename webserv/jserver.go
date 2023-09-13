package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"net/url"

	"log"
	"net/http"
)

type ServerApp interface {
	PrepareSession(s Session)
	HandleRequest(s Session, path string) bool
}

type JServerStruct struct {
	BaseURL        string // e.g. "jeff.org"
	KeyDir         Path
	SessionManager SessionManager
	App            ServerApp
	Resources      Path
}

type JServer = *JServerStruct

func NewJServer() JServer {
	t := &JServerStruct{}
	return t
}

func (s JServer) StartServing() {

	var ourUrl = "jeff.org"

	var keyDir = s.KeyDir //oper.appRoot.JoinM("https_keys")
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
			Todo("!This should be moved to the webserv package, maybe if an initialization parameter was specified")
			s.handle(w, req)
		})

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

// A handler such as this must be thread safe!
func (s JServer) handle(w http.ResponseWriter, req *http.Request) {
	pr := PrIf(false)
	pr("handler, request:", req.RequestURI)

	// We don't know what the session is yet, so we don't have a lock on it...
	sess := DetermineSession(s.SessionManager, w, req, true)

	// Now that we have the session, lock it
	sess.Lock.Lock()
	defer sess.ReleaseLockAndDiscardRequest()
	sess.ResponseWriter = w
	sess.Request = req

	if !sess.prepared {
		sess.prepared = true
		{
			// Assign a widget heirarchy to the session
			m := sess.WidgetManager()
			m.Id(WidgetIdPage)
			widget := m.Open()
			sess.PageWidget = widget
			m.Close()
		}
		s.App.PrepareSession(sess)
	}

	url, err := url.Parse(req.RequestURI)
	if err == nil {

		path := url.Path
		var text string
		var flag bool

		pr("url path:", path)
		if path == "/ajax" {
			Todo("!Use TrimIfPrefix here as well")
			sess.HandleAjaxRequest()
			//} else if text, flag = TrimIfPrefix(path, "/r/"); flag {
			//	pr("handling blob request with:", text)
			//	err = oper.handleBlobRequest(w, req, text)
		} else if text, flag = TrimIfPrefix(path, `/upload/`); flag {
			pr("handling upload request with:", text)
			sess.HandleUploadRequest(text)
		} else {
			result := s.App.HandleRequest(sess, path)
			if !result {
				// If we fail to parse any requests, assume it's a resource, like that stupid favicon
				pr("handling resource request for:", path)
				err = sess.HandleResourceRequest(s.Resources)
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
		Pr("...problem with request, URL:", req.RequestURI, INDENT, p)
	}
}
