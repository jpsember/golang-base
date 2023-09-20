package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type mgrState struct {
	Parent        Widget
	StateProvider WidgetStateProvider
	IdPrefix      string
	DebugTag      string
}

type WidgetManagerObj struct {
	BaseObject
	widgetMap           WidgetMap
	stack               []mgrState
	pendingSize         WidgetSize
	pendingAlign        WidgetAlign
	pendingId           string
	pendingLabel        string
	anonymousIdCounter  int
	pendingChildColumns int
}

func NewWidgetManager(session Session) WidgetManager {
	w := WidgetManagerObj{
		widgetMap: make(map[string]Widget),
	}
	w.initStateStack()
	w.SetName("WidgetManager")
	w.resetPendingColumns()
	w.LogCols("Constructed")
	return &w
}

func (m WidgetManager) initStateStack() {
	m.stack = []mgrState{{}}
}

func (m WidgetManager) LogCols(message string) {
	if !Alert("!remove this at some point") {
		Pr("WidgetManager pending child columns:", m.pendingChildColumns)
	}
}

type WidgetManager = *WidgetManagerObj

// Determine if a widget exists
func (m WidgetManager) Exists(id string) bool {
	return m.find(id) != nil
}

func (m WidgetManager) Get(id string) Widget {
	w := m.find(id)
	if w == nil {
		BadState("Can't find widget with id:", Quoted(id))
	}
	return w
}

func (m WidgetManager) Opt(id string) Widget {
	return m.find(id)
}

func (m WidgetManager) find(id string) Widget {
	return m.widgetMap[id]
}

// ------------------------------------------------------------------------------------
// Constructing widgets
// ------------------------------------------------------------------------------------

func (m WidgetManager) Id(id string) WidgetManager {
	m.pendingId = m.IdPrefix() + id
	return m
}

func (m WidgetManager) consumePendingId() string {
	id := m.pendingId
	CheckNonEmpty(id, "no pending id")
	m.pendingId = ""
	return id
}
func (m WidgetManager) consumeOptionalPendingId() string {
	id := m.pendingId
	if id != "" {
		m.pendingId = ""
	} else {
		id = m.AllocateAnonymousId("")
	}
	return id
}

// Set size for next widget (what size means depends upon the widget type).
func (m WidgetManager) Size(size WidgetSize) WidgetManager {
	m.pendingSize = size
	return m
}

// Set horizontal alignment for next widget
func (m WidgetManager) Align(align WidgetAlign) WidgetManager {
	m.pendingAlign = align
	return m
}

// Set number of Bootstrap columns the next widget will occupy within its container.
func (m WidgetManager) Col(columns int) WidgetManager {
	m.pendingChildColumns = columns
	m.LogCols("Set col;")
	return m
}

func (m WidgetManager) Label(value string) WidgetManager {
	CheckState(m.pendingLabel == "")
	m.pendingLabel = value
	return m
}

func (m WidgetManager) consumePendingLabel() string {
	lbl := m.pendingLabel
	m.pendingLabel = ""
	return lbl
}

func (m WidgetManager) consumePendingSize() WidgetSize {
	x := m.pendingSize
	m.pendingSize = SizeDefault
	return x
}

func (m WidgetManager) consumePendingAlign() WidgetAlign {
	x := m.pendingAlign
	m.pendingAlign = AlignDefault
	return x
}

func verifyUsed(flag bool, name string) {
	if flag {
		return
	}
	BadState("unused value:", name)
}

func (m WidgetManager) clearPendingComponentFields() {
	// If some values were not used, issue warnings
	verifyUsed(m.pendingLabel == "", "pendingLabel")
	verifyUsed(m.pendingSize == SizeDefault, "pendingSize")
	verifyUsed(m.pendingAlign == AlignDefault, "pendingAlign")
}

/**
 * Add widget to the hierarchy
 */
func (m WidgetManager) Add(widget Widget) WidgetManager {
	id := widget.Id()
	if id != "" {
		if m.Exists(id) {
			BadState("<1Attempt to add widget with duplicate id:", id)
		}
		m.widgetMap[id] = widget
	}
	// Set its state provider, if it doesn't already have one
	if widget.StateProvider() == nil {
		widget.SetStateProvider(m.StateProvider())
	}

	m.Log("addWidget, id:", id, "panel stack size:", len(m.stack))
	state := m.stackedState()
	parent := state.Parent
	if parent != nil {
		parent.AddChild(widget, m)
	}
	m.clearPendingComponentFields()

	// Ask widget to add any children that it may need
	widget.AddChildren(m)
	return m
}

func (m WidgetManager) stackedState() *mgrState {
	return &m.stack[len(m.stack)-1]
}

// Detach a widget that has just been constructed from the WidgetManager and its container.
func (m WidgetManager) Detach(widget Widget) Widget {
	result := m.Opt(widget.Id())
	if result == nil {
		BadArg("Cannot detach widget; not in manager set:", widget.Id())
	}
	container := m.parentWidget()
	container.RemoveChild(widget)

	// Delete all widgets within the detached widget's tree from the widget map
	idList := getWidgetsInTree(widget)
	for _, id := range idList {
		delete(m.widgetMap, id)
	}
	return result
}

func getWidgetsInTree(widget Widget) []string {
	return auxGetWidgetsInTree(widget, nil)
}

func auxGetWidgetsInTree(widget Widget, list []string) []string {
	list = append(list, widget.Id())
	for _, c := range widget.Children() {
		list = auxGetWidgetsInTree(c, list)
	}
	return list
}

// Have subsequent WidgetManager operations operate on a particular container widget.
// The container is marked for repainting.
func (m WidgetManager) With(container Widget) WidgetManager {
	cont := container.(GridWidget)
	id := cont.Id()

	CheckState(m.Exists(id))

	// Discard any existing child widgets
	m.removeWidgets(cont.Children())
	cont.ClearChildren()
	m.initStateStack()
	m.PushContainer(container)
	m.resetPendingColumns()
	return m
}

func (m WidgetManager) resetPendingColumns() {
	m.pendingChildColumns = MaxColumns
}

// Add a child GridContainerWidget, and push onto stack as active container
func (m WidgetManager) Open() Widget {
	m.Log("open")
	widget := NewContainerWidget(m.consumeOptionalPendingId())
	m.Add(widget)
	return m.OpenContainer(widget)
}

// Push a container widget onto the stack as an active container
func (m WidgetManager) OpenContainer(widget Widget) Widget {
	m.Log("Adding container widget")
	itm := *m.stackedState()
	itm.Parent = widget
	m.pushState(itm, tag_container)
	m.Log("added container to stack")
	return widget
}

func (m WidgetManager) pushState(state mgrState, tag string) {
	state.DebugTag = tag
	m.stack = append(m.stack, state)
}

const (
	tag_container = "container"
	tag_prefix    = "prefix"
	tag_provider  = "provider"
)

// Pop the active container from the stack.
func (m WidgetManager) Close() WidgetManager {
	m.Log("Close")
	m.popStack(tag_container)
	return m
}

func (m WidgetManager) popStack(tag string) {
	top := m.stackedState()
	if top.DebugTag != tag {
		BadState("attempt to pop state stack, tag is:", top.DebugTag, "but expected:", tag)
	}
	_, m.stack = PopLast(m.stack)
}

//// Verify that no unused 'pending' arguments exist, calls are balanced, etc
//func (m WidgetManager) finish() WidgetManager {
//	m.clearPendingComponentFields()
//	if len(m.stack) != 1 {
//		sb := strings.Builder{}
//		for _, x := range m.stack {
//			sb.WriteByte(' ')
//			sb.WriteString(x.DebugTag)
//		}
//		BadState("state stack nonempty! Tags:", m.dumpStateStack(-1))
//	}
//	return m
//}

func (m WidgetManager) dumpStateStack(cursor int) string {
	sb := strings.Builder{}
	for index, x := range m.stack {
		sb.WriteByte(' ')
		if index == cursor {
			sb.WriteByte('>')
		}
		sb.WriteString(x.DebugTag)
	}
	return sb.String()
}

func (m WidgetManager) parentWidget() Widget {
	return m.stackedState().Parent
}

func (m WidgetManager) AddInput(listener InputWidgetListener) InputWidget {
	return m.auxAddInput(listener, false)
}

func (m WidgetManager) AddUserHeader() UserHeaderWidget {
	w := NewUserHeaderWidget(m.consumeOptionalPendingId())
	w.BgndImageMarkup = `style=" height:50px; background-image:url('app_header.jpg'); background-repeat: no-repeat;"`
	m.Add(w)
	return w
}

func (m WidgetManager) auxAddInput(listener InputWidgetListener, password bool) InputWidget {
	id := m.consumeOptionalPendingId()
	t := NewInputWidget(id, NewHtmlString(m.consumePendingLabel()), listener, password)
	m.Add(t)
	return t
}

func (m WidgetManager) AddPassword(listener InputWidgetListener) InputWidget {
	return m.auxAddInput(listener, true)
}

func (m WidgetManager) AddList(list ListInterface, itemWidget Widget, provider ListItemStateProvider) ListWidget {
	id := m.consumeOptionalPendingId()
	t := NewListWidget(id, list, itemWidget, provider)
	m.Add(t)
	return t
}

// Utility method to determine the label and id for text fields (text fields, headings).
// The label can either be expressed as a string (static content),
// or an id (dynamic content, read from session state).  If static, there should *not* be
// a pending id.
func (m WidgetManager) getStaticContentAndId() (string, string) {
	staticContent := m.consumePendingLabel()
	hasStaticContent := staticContent != ""
	if hasStaticContent {
		CheckState(m.pendingId == "", "specify id OR static content")
	}
	id := m.consumeOptionalPendingId()
	return staticContent, id
}

func (m WidgetManager) AddHeading() WidgetManager {
	staticContent, id := m.getStaticContentAndId()
	w := NewHeadingWidget(id)
	w.SetSize(m.consumePendingSize())
	Todo("!Setting WidgetSize seems to have no effect on headings")
	w.SetAlign(m.consumePendingAlign())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	return m.Add(w)
}

func (m WidgetManager) AddText() WidgetManager {
	staticContent, id := m.getStaticContentAndId()
	w := NewTextWidget(id, m.consumePendingSize())
	w.SetStateProvider(m.StateProvider())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	m.Log("Adding text, id:", w.BaseId)
	return m.Add(w)
}

func (m WidgetManager) AddButton(listener ButtonWidgetListener) ButtonWidget {
	w := NewButtonWidget(m.consumeOptionalPendingId(), listener)
	w.SetSize(m.consumePendingSize())
	w.SetAlign(m.consumePendingAlign())
	m.Log("Adding button, id:", w.BaseId)
	w.Label = NewHtmlString(m.consumePendingLabel())
	m.Add(w)
	return w
}

func (m WidgetManager) AddSpace() WidgetManager {
	return m.Add(NewBaseWidget(m.consumeOptionalPendingId()))
}

func doNothingFileUploadListener(sess Session, widget FileUpload, value []byte) error {
	Pr("'do nothing' FileUploadListener called with bytes:", len(value))
	return nil
}

func (m WidgetManager) AddFileUpload(listener FileUploadWidgetListener) FileUpload {
	if listener == nil {
		listener = doNothingFileUploadListener
	}
	w := NewFileUpload(m.consumePendingId(), NewHtmlString(m.consumePendingLabel()), listener)
	m.Add(w)
	return w
}

func (m WidgetManager) AddImage() ImageWidget {
	w := NewImageWidget(m.consumePendingId())
	m.Add(w)
	return w
}

func (m WidgetManager) AddCheckbox(listener CheckboxWidgetListener) CheckboxWidget {
	return m.checkboxHelper(listener, false)
}

func (m WidgetManager) AddSwitch(listener CheckboxWidgetListener) CheckboxWidget {
	return m.checkboxHelper(listener, true)
}

func (m WidgetManager) checkboxHelper(listener CheckboxWidgetListener, switchFlag bool) CheckboxWidget {
	w := NewCheckboxWidget(switchFlag, m.consumePendingId(), NewHtmlString(m.consumePendingLabel()), listener)
	m.Add(w)
	return w
}

func (m WidgetManager) AllocateAnonymousId(debugInfo string) string {
	m.anonymousIdCounter++
	result := "z" + IntToString(m.anonymousIdCounter)
	if debugInfo != "" {
		result += "_" + debugInfo + "_"
	}
	return result
}

func (m WidgetManager) removeWidgets(widgets []Widget) {
	for _, widget := range widgets {
		m.Remove(widget)
	}
}

// Remove widget (if it exists), and the subtree of widgets it may contain.
func (m WidgetManager) Remove(widget Widget) WidgetManager {
	id := widget.Id()
	if m.Exists(id) {
		delete(m.widgetMap, id)
		m.removeWidgets(widget.Children())
	}
	return m
}

func (m WidgetManager) WidgetMapSummary() JSMap {
	mp := NewJSMap()
	for id, widget := range m.widgetMap {
		mp.Put(id, TypeOf(widget))
	}
	return mp
}

var WidgetDebugRenderingFlag bool

// Deprecated. To have uses show up in editor as a warning.
func SetWidgetDebugRendering() {
	Alert("<1 Setting widget debug rendering")
	WidgetDebugRenderingFlag = true
}

// Mark widgets for repainting (if they exist).  Does nothing if there is no repaintSet.
func (s Session) RepaintIds(ids ...string) WidgetManager {
	m := s.WidgetManager()
	for _, id := range ids {
		w := m.Opt(id)
		if w != nil {
			s.Repaint(w)
		} else {
			Alert("#50<1Can't find widget to repaint with id:", id)
		}
	}
	return m
}

func (m WidgetManager) PushContainer(container Widget) WidgetManager {
	Todo("!this is a lot like OpenContainer, but without the adding")
	// Push a container widget onto the stack
	itm := *m.stackedState()
	itm.Parent = container
	m.pushState(itm, tag_container)
	return m
}

func (m WidgetManager) PushStateProvider(p WidgetStateProvider) {
	itm := *m.stackedState()
	itm.StateProvider = p
	m.pushState(itm, tag_provider)
}

func (m WidgetManager) PopStateProvider() {
	m.popStack(tag_provider)
}

func (m WidgetManager) StateProvider() WidgetStateProvider {
	return m.stackedState().StateProvider
}

func (m WidgetManager) PushIdPrefix(prefix string) {
	itm := *m.stackedState()
	itm.IdPrefix = prefix
	m.pushState(itm, tag_prefix)
}

func (m WidgetManager) PopIdPrefix() {
	m.popStack(tag_prefix)
}

func (m WidgetManager) IdPrefix() string {
	return m.stackedState().IdPrefix
}

// Debug method to verify that various push/pop operations of the state stack are balanced.
// Call EndConstruction() with the value that this returns to confirm.
func (m WidgetManager) StartConstruction() int {
	return len(m.stack)
}

// Debug method to verify that various push/pop operations of the state stack are balanced.
// Call EndConstruction() with the value that StartConstruction() returned.
func (m WidgetManager) EndConstruction(expectedStackSize int) {
	if len(m.stack) != expectedStackSize {
		BadState("expected state stack to be at", expectedStackSize, "but is at", len(m.stack), INDENT, m.dumpStateStack(expectedStackSize))
	}
}
