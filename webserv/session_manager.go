package webserv

import (
	"crypto/rand"
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	"io"
)

type SessionManager interface {
	FindSession(id string) Session
	CreateSession() Session
	SetModified(session Session)
	DiscardAllSessions()
}

func RandomSessionId() string {
	var idLength = 32
	b := make([]byte, idLength)
	CheckOkWith(io.ReadFull(rand.Reader, b))
	return base64.URLEncoding.EncodeToString(b)
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
		b.Id = RandomSessionId()
		_, ok := s.sessionMap.Provide(b.Id, b)
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
const clientKeyURL = "u"
