package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"net/http"
	"strings"
)

var DebugWidgetRepaint = true && Alert("DebugWidgetRepaint is in effect")

// This function must be threadsafe!
func DetermineSession(manager SessionManager, w http.ResponseWriter, req *http.Request, createIfNone bool) Session {

	pr := PrIf("DetermineSession", false)
	const sessionCookieName = "session_cookie"

	// Determine what session this is, by examining cookies
	var session Session
	cookies := req.Cookies()
	pr("getting cookies for request:", req.URL)
	for _, c := range cookies {
		pr("cookie:", c.Name, "value:", c.Value)
		if c.Name == sessionCookieName {
			sessionId := c.Value
			session = manager.FindSession(sessionId)
		}
		if session != nil {
			pr("found session:", session.SessionId)
			break
		}
	}

	// If no session was found, create one, and send a cookie
	if session == nil && createIfNone {
		session = manager.CreateSession()
		cookie := &http.Cookie{
			Name:   sessionCookieName,
			Value:  session.SessionId,
			MaxAge: 1200, // 20 minutes
			Path:   `/`,
		}
		pr("No cookie found, so creating session:", session.SessionId)
		http.SetCookie(w, cookie)
	}
	return session
}

func WriteResponse(writer http.ResponseWriter, contentType string, response []byte) error {
	if contentType == "" {
		BadArg("<1No response type!")
	}
	writer.Header().Set("Content-Type", contentType)
	_, err := writer.Write(response)
	Todo("!Do I need to explicitly close the writer?")
	return err
}

func InferContentTypeM(path string) string {
	result, found := InferContentType(path)
	if !found {
		BadArg("<1Unknown Content-Type for:", path)
	}
	return result
}

func InferContentType(path string) (string, bool) {
	ext := ExtensionFrom(path)
	result, found := fileExtensionMap[ext]
	return result, found
}

var fileExtensionMap = BuildStringStringMap(strings.Fields(`
ico image/x-icon 
bin application/octet-stream 
css text/css 
jpg image/jpeg 
js text/javascript 
json application/json 
png image/png 
txt text/plain
`)...)
