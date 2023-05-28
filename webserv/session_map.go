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

type Session = *SessionDataBuilder

type SessionMap struct {
	BaseObject
	sessionMap    map[string]Session
	lastWrittenMs int64
	modified      bool
	lock          sync.RWMutex
	persistPath   Path
}

func BuildSessionMap() *SessionMap {
	sm := new(SessionMap)
	sm.SetName("SessionMap")
	sm.sessionMap = make(map[string]Session)

	// If there's a file on disk to restore from, do so
	// (in future, use a database or something)
	pth := sm.getPath()
	if pth.Exists() {
		json := JSMapFromFileIfExistsM(pth)
		for k, v := range json.WrappedMap() {
			s := ParseSessionData(v).ToBuilder()
			sm.sessionMap[k] = s
		}
		sm.lastWrittenMs = time.Now().UnixMilli()
	}
	return sm
}

func (s *SessionMap) FindSession(id string) Session {
	s.Log("FindSession, id:", id)
	s.lock.RLock()
	defer s.lock.RUnlock()
	var result = s.sessionMap[id]
	s.Log("Result:", INDENT, result)
	return result
}

func randomSessionId() string {
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

func (s *SessionMap) CreateSession() Session {
	s.lock.Lock()

	b := NewSessionData().ToBuilder()

	for {
		b.SetId(randomSessionId())
		// Stop looking for session ids if we've found one that isn't used
		if s.sessionMap[b.Id()] == nil {
			break
		}
	}
	s.sessionMap[b.Id()] = b
	s.setModified()
	s.lock.Unlock()

	Todo("have a background task handle flushing any modifications")
	s.flush()
	return b
}

func (s *SessionMap) setModified() {
	s.modified = true
}

func (s *SessionMap) flush() {
	if !s.modified {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.modified {
		pth := s.getPath()
		Pr("writing session map to path:", pth)
		jsm := NewJSMap()
		for k, v := range s.sessionMap {
			jsm.Put(k, v.ToJson())
		}
		pth.WriteStringM(PrintJSEntity(jsm, false))
		s.lastWrittenMs = time.Now().UnixMilli()
		Pr("flushed modified session map to:", pth)
		s.modified = false
	}

}

func (s *SessionMap) getPath() Path {
	if s.persistPath == "" {
		pth := AscendToDirectoryContainingFileM("", "go.mod")
		pth = pth.JoinM("webserv/cache")
		if !pth.IsDir() {
			pth.MkDirsM()
		}
		s.persistPath = pth.JoinM("session_map.json")
	}
	return s.persistPath
}
