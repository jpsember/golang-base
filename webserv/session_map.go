package webserv

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type SessionMap struct {
	sessionMap  map[string]*Session
	sessionLock sync.RWMutex

	uniqueSessionId atomic.Int64
}

func BuildSessionMap() *SessionMap {
	sm := new(SessionMap)
	sm.sessionMap = make(map[string]*Session)
	return sm
}

func (s *SessionMap) FindSession(id string) *Session {
	s.sessionLock.RLock()
	session := s.sessionMap[id]
	s.sessionLock.RUnlock()
	return session
}

func (s *SessionMap) CreateSession() *Session {
	s.sessionLock.Lock()
	ourId := s.uniqueSessionId.Add(1)
	session := new(Session)
	session.Id = fmt.Sprintf("%v", ourId)
	s.sessionMap[session.Id] = session
	s.sessionLock.Unlock()
	return session
}
