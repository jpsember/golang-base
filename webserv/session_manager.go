package webserv

import (
	"crypto/rand"
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	"io"
	"strings"
)

type SessionManager interface {
	FindSession(id string) Session
	CreateSession() Session
	SetModified(session Session)
	DiscardAllSessions()
}

func RandomSessionId() string {
	debug := Alert("!using smaller session ids for development")
	var idLength = 32
	if debug {
		idLength = 3
	}
	b := make([]byte, idLength)
	CheckOkWith(io.ReadFull(rand.Reader, b))
	result := base64.URLEncoding.EncodeToString(b)
	Todo("!what are legal characters in session id?  is base64 overkill?")
	if debug {
		result = strings.ToUpper(result)
	}
	return result
}

type inMemorySessionMap struct {
	BaseObject
	sessionMap *ConcurrentMap[string, Session]
}

func BuildSessionMap() SessionManager {
	sm := new(inMemorySessionMap)
	sm.SetName("inMemorySessionMap")
	//sm.SetVerbose(Alert("setting verbosity"))
	sm.sessionMap = NewConcurrentMap[string, Session]()
	return sm
}

func (s *inMemorySessionMap) SetModified(session Session) {
}

func (s *inMemorySessionMap) FindSession(id string) Session {
	s.Log("FindSession, id:", id)
	result := s.sessionMap.Get(id)
	s.Log("Result:", INDENT, result)
	return result
}

func (s *inMemorySessionMap) CreateSession() Session {

	b := NewSession()
	for {
		b.SessionId = RandomSessionId()
		_, ok := s.sessionMap.Provide(b.SessionId, b)
		// Stop looking for session ids if we've found one that isn't used
		if !ok {
			break
		}
	}
	s.Log("Created new session:", INDENT, b)
	return b
}

func (s *inMemorySessionMap) DiscardAllSessions() {
	s.sessionMap.Clear()
}

// Get a string value from session state map
func WidgetStringValue(state JSMap, id string) string {
	return state.OptString(id, "")
}

const clientKeyWidget = "w"
const clientKeyValue = "v"
const clientKeyInfo = "i"
