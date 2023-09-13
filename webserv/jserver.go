package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"net/url"

	"log"
	"net/http"
)

type ServerApp interface {
	PrepareSession(s Session)
	HandleRequest(w http.ResponseWriter, s Session, path string) bool
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
	pr := PrIf(true)
	pr("handler, request:", req.RequestURI)

	// We don't know what the session is yet, so we don't have a lock on it...
	sess := DetermineSession(s.SessionManager, w, req, true)

	Todo("This shouldn't be done until we have a lock on the session; maybe lock the session throughout the handler?  Or do we already have it?")

	// Now that we have the session, lock it
	Todo("But when we kill the session, i.e. logging out, do we still have the lock?")
	sess.Lock.Lock()
	defer sess.ReleaseLockAndDiscardRequest()
	Todo("We can (temporarily) store the ResponseWriter, Request in the session for simplicity")

	if !sess.prepared {
		sess.prepared = true
		s.App.PrepareSession(sess)
	}

	url, err := url.Parse(req.RequestURI)
	if err == nil {

		Todo("!Move as much of this as possible to the webserv package")
		path := url.Path
		var text string
		var flag bool

		pr("url path:", path)
		if path == "/ajax" {
			Todo("!Use TrimIfPrefix here as well")
			sess.HandleAjaxRequest(w, req)
			//} else if text, flag = TrimIfPrefix(path, "/r/"); flag {
			//	pr("handling blob request with:", text)
			//	err = oper.handleBlobRequest(w, req, text)
		} else if text, flag = TrimIfPrefix(path, `/upload/`); flag {
			pr("handling upload request with:", text)
			sess.HandleUploadRequest(w, req, text)
		} else {
			result := s.App.HandleRequest(w, sess, path)
			//result := oper.animalURLRequestHandler(w, req, sess, path)
			if !result {
				// If we fail to parse any requests, assume it's a resource, like that stupid favicon
				pr("handling resource request for:", path)
				err = sess.HandleResourceRequest(w, req, s.Resources)
			}
		}
	}

	if err != nil {
		sess.SetRequestProblem(err)
	}

	Todo("This code should be done while the lock is still held")
	if p := sess.GetRequestProblem(); p != nil {
		Pr("...problem with request, URL:", req.RequestURI, INDENT, p)
	}
}
