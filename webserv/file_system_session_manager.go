package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"sync"
	"time"
)

type FileSystemSessionManager struct {
	BaseObject
	sessionMap    map[string]Session
	lastWrittenMs int64
	modified      bool
	lock          sync.RWMutex
	persistPath   Path
}

func BuildFileSystemSessionMap() *FileSystemSessionManager {
	sm := new(FileSystemSessionManager)
	sm.SetName("FileSystemSessionManager")
	sm.sessionMap = make(map[string]Session)

	// If there's a file on disk to restore from, do so
	// (in future, use a database or something)
	pth := sm.getPath()
	if pth.Exists() {
		json := JSMapFromFileIfExistsM(pth)
		for k, v := range json.WrappedMap() {
			s := ParseSession(v)
			sm.sessionMap[k] = s
		}
		sm.lastWrittenMs = time.Now().UnixMilli()
	}
	return sm
}

func (s *FileSystemSessionManager) SetModified(session Session) {
	Todo("Have some process flush changes periodically")
}

func (s *FileSystemSessionManager) FindSession(id string) Session {
	s.Log("FindSession, id:", id)
	s.lock.RLock()
	defer s.lock.RUnlock()
	var result = s.sessionMap[id]
	s.Log("Result:", INDENT, result)
	return result
}

func (s *FileSystemSessionManager) CreateSession() Session {
	s.lock.Lock()

	b := NewSession()

	for {
		b.Id = RandomSessionId()
		// Stop looking for session ids if we've found one that isn't used
		if s.sessionMap[b.Id] == nil {
			break
		}
	}
	s.sessionMap[b.Id] = b
	s.setModified()
	s.lock.Unlock()

	Todo("have a background task handle flushing any modifications")
	s.flush()
	return b
}

func (s *FileSystemSessionManager) setModified() {
	s.modified = true
}

func (s *FileSystemSessionManager) flush() {
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

func (s *FileSystemSessionManager) getPath() Path {
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
