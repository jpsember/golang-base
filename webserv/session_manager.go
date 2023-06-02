package webserv

import (
	"crypto/rand"
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"io"
	"net/http"
	"sync"
)

type Session = *SessionStruct

type SessionStruct struct {
	Id string
	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Mutex sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap
	// Map of widgets for this session
	WidgetMap  map[string]Widget
	repaintMap *Set[string]
	// TODO: we might want the repaintMap to be ephemeral, only alive while serving the request
	// We also might want to have a singleton, global widget map, since the state is stored here in the session
}

func NewSession() Session {
	s := SessionStruct{
		State:      NewJSMap(),
		repaintMap: NewSet[string](),
	}
	Todo("Restore user session from filesystem/database")
	return &s
}

func (s Session) ToJson() *JSMapStruct {
	m := NewJSMap()
	m.Put("id", s.Id)
	return m
}

// Mark a widget for repainting
func (s Session) Repaint(id string) {
	s.repaintMap.Add(id)
}

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMapStruct)
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

// Get a string value from session state map
func WidgetStringValue(state JSMap, id string) string {
	return state.OptString(id, "")
}

func (s Session) sendAjaxMarkup(w http.ResponseWriter, req *http.Request) {
	pr := PrIf(true)

	jsmap := NewJSMap()

	// TODO: there might be a more efficient way to do the repainting

	// Determine which widgets need repainting
	if s.repaintMap.Size() != 0 {

		// refmap will be the map sent to the client with the widgets

		refmap := NewJSMap()
		jsmap.Put("w", refmap)

		painted := NewSet[string]()

		for k, _ := range s.repaintMap.WrappedMap() {
			w := s.WidgetMap[k]
			addSubtree(painted, w)
		}

		// Do a depth first search of the widget map, sending widgets that have been marked for painting
		stack := NewArray[string]()
		stack.Add(s.PageWidget.GetId())
		for stack.NonEmpty() {
			widgetId := stack.Pop()
			widget := s.WidgetMap[widgetId]
			if painted.Contains(widgetId) {
				m := NewMarkupBuilder()
				widget.RenderTo(m, s.State)
				refmap.Put(widgetId, m.String())
			}
			for _, child := range widget.GetChildren() {
				stack.Add(child.GetId())
			}
		}
	}

	content := jsmap.CompactString()

	pr("sending back to Ajax caller:", INDENT, content)
	w.Write([]byte(content))
}
