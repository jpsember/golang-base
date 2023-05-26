package webserv

import (
	"crypto/rand"
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/gen/webservgen"
	. "github.com/jpsember/golang-base/json"
	"io"
	"sync"
	"time"
)

type SessionMap struct {
	BaseObject
	sessionMap    *JSMap
	lastWrittenMs int64
	modified      bool
	lock          sync.RWMutex
	persistPath   Path
}

func BuildSessionMap() *SessionMap {
	sm := new(SessionMap)
	sm.SetName("SessionMap")
	sm.SetVerbose(true)

	// If there's a file on disk to restore from, do so
	// (in future, use a database or something)
	pth := sm.getPath()
	if pth.Exists() {
		sm.sessionMap = JSMapFromFileIfExistsM(pth)
		sm.lastWrittenMs = time.Now().UnixMilli()
	} else {
		sm.sessionMap = NewJSMap()
	}
	CheckState(sm.sessionMap != nil)
	return sm
}

func (s *SessionMap) FindSession(id string) Session {
	s.Log("FindSession, id:", id)
	s.lock.RLock()
	defer s.lock.RUnlock()
	var result Session
	sm := s.sessionMap.OptMap(id)
	if sm != nil {
		result = DefaultSession.Parse(sm).(Session)
	}
	s.Log("Restult:", INDENT, result)
	return result
}

func randomSessionId() string {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	CheckOk(err)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *SessionMap) CreateSession() Session {
	s.lock.Lock()

	b := NewSession()
	for {
		b.SetId(randomSessionId())
		if s.sessionMap.OptMap(b.Id()) == nil {
			break
		}
	}
	session := b.Build()
	s.sessionMap.Put(session.Id(), session.ToJson())
	s.setModified()
	s.lock.Unlock()

	Todo("have a background task handle flushing any modifications")
	s.flush()

	return session
}

func (s *SessionMap) setModified() {
	s.modified = true
}

func (s *SessionMap) flush() {
	if !s.modified {
		return
	}
	s.lock.Lock()

	if s.modified {
		pth := s.getPath()
		Pr("writing session map to path:", pth, INDENT, s.sessionMap)
		pth.WriteStringM(s.sessionMap.String())
		s.lastWrittenMs = time.Now().UnixMilli()
		Pr("flushed modified session map to:", pth)
		s.modified = false
	}

	defer s.lock.Unlock()
}

func (s *SessionMap) getPath() Path {
	if s.persistPath == "" {
		pth, err := AscendToDirectoryContainingFile("", "go.mod")
		CheckOkWithSkip(1, err)
		pth = pth.JoinM("webserv/cache")
		if !pth.IsDir() {
			pth.MkDirsM()
		}
		s.persistPath = pth.JoinM("session_map.json")
	}
	return s.persistPath
}
