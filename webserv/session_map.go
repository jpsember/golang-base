package webserv

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/jpsember/golang-base/base"
	"io"
	"sync"
)

type SessionMap struct {
	sessionMap map[string]*Session
	lock       sync.RWMutex
}

func BuildSessionMap() *SessionMap {
	sm := new(SessionMap)
	sm.sessionMap = make(map[string]*Session)
	return sm
}

func (s *SessionMap) FindSession(id string) *Session {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.sessionMap[id]
}

func randomSessionId() string {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	base.CheckOk(err)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *SessionMap) CreateSession() *Session {
	s.lock.Lock()
	defer s.lock.Unlock()

	session := new(Session)
	for {
		session.Id = randomSessionId()
		if s.sessionMap[session.Id] == nil {
			break
		}
	}
	s.sessionMap[session.Id] = session
	return session
}
