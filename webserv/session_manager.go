package webserv

import (
	"crypto/rand"
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"io"
	"sync"
)

type Session = *SessionStruct

type SessionStruct struct {
	Id string
	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Mutex sync.RWMutex
}

func NewSession() Session {
	s := SessionStruct{}
	return &s
}

func (s Session) ToJson() *JSMap {
	m := NewJSMap()
	m.Put("id", s.Id)
	return m
}

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMap)
	var n = NewSession()
	n.Id = s.OptString("id", "")
	return n
}

type SessionManager interface {
	FindSession(id string) Session
	CreateSession() Session
	SetModified(session Session)
}

func RandomSessionId() string {
	var idLength = 32
	if true {
		// For now, use a much smaller id for legibility
		idLength = 3
	}
	b := make([]byte, idLength)
	_, err := io.ReadFull(rand.Reader, b)
	CheckOk(err)
	return base64.URLEncoding.EncodeToString(b)
}

type inMemorySessionMap struct {
	BaseObject
	sessionMap map[string]Session
	lock       sync.RWMutex
}

func BuildSessionMap() SessionManager {
	sm := new(inMemorySessionMap)
	sm.SetName("inMemorySessionMap")
	sm.SetVerbose(Alert("setting verbosity"))
	sm.sessionMap = make(map[string]Session)
	return sm
}

func (s *inMemorySessionMap) SetModified(session Session) {
}

func (s *inMemorySessionMap) FindSession(id string) Session {
	s.Log("FindSession, id:", id)
	s.lock.RLock()
	defer s.lock.RUnlock()
	var result = s.sessionMap[id]
	s.Log("Result:", INDENT, result)
	return result
}

func (s *inMemorySessionMap) CreateSession() Session {
	s.lock.Lock()

	b := NewSession()
	for {
		b.Id = RandomSessionId()
		// Stop looking for session ids if we've found one that isn't used
		if s.sessionMap[b.Id] == nil {
			break
		}
	}
	s.Log("Creating new session:", INDENT, b)
	s.sessionMap[b.Id] = b
	s.lock.Unlock()
	return b
}
