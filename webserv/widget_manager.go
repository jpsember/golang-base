package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type mgrState struct {
	Parent              Widget
	StateProvider       WidgetStateProvider
	IdPrefix            string
	DebugTag            string
	pendingChildColumns int
}

type WidgetManagerObj struct {
	widgetMap          WidgetMap
	stack              []mgrState
	pendingSize        WidgetSize
	pendingAlign       WidgetAlign
	pendingHeight      int
	pendingId          string
	pendingLabel       string
	anonymousIdCounter int
}

func (m WidgetManager) InitializeWidgetManager() {
	DebVerifyServerStarted()
	m.widgetMap = make(map[string]Widget)
	m.initStateStack()
	m.resetPendingColumns()
}

func (m WidgetManager) initStateStack() {
	m.stack = []mgrState{{}}
}

type WidgetManager = *WidgetManagerObj

func (m WidgetManager) Opt(id string) Widget {
	return m.widgetMap[id]
}

func (m WidgetManager) exists(id string) bool {
	return m.Opt(id) != nil
}

func (m WidgetManager) Get(id string) Widget {
	w := m.Opt(id)
	if w == nil {
		BadState("Can't find widget with id:", Quoted(id))
	}
	return w
}

// ------------------------------------------------------------------------------------
// Constructing widgets
// ------------------------------------------------------------------------------------

func (m WidgetManager) Id(id string) WidgetManager {
	v := m.IdPrefix() + id
	AssertNoDots(v)
	m.pendingId = v
	return m
}

func (m WidgetManager) consumePendingId() string {
	id := m.pendingId
	CheckNonEmpty(id, "no pending id")
	m.pendingId = ""
	return id
}
func (m WidgetManager) ConsumeOptionalPendingId() string {
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

// Set height for next widget (for text, this is e.g. 5em).
func (m WidgetManager) Height(ems int) WidgetManager {
	m.pendingHeight = ems
	return m
}

// Set horizontal alignment for next widget
func (m WidgetManager) Align(align WidgetAlign) WidgetManager {
	m.pendingAlign = align
	return m
}

// Set number of Bootstrap columns the next widget will occupy within its container.
func (m WidgetManager) Col(columns int) WidgetManager {
	m.stackedState().pendingChildColumns = columns
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

func (m WidgetManager) consumePendingHeight() int {
	x := m.pendingHeight
	m.pendingHeight = 0
	return x
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
	verifyUsed(m.pendingHeight == 0, "pendingHeight")
}

/**
 * Add widget to the hierarchy
 */
func (m WidgetManager) Add(widget Widget) WidgetManager {
	id := widget.Id()
	if id != "" {
		if m.exists(id) {
			BadState("<1Attempt to add widget with duplicate id:", id, "widget ids:", m.IdSummary())
		}
		m.widgetMap[id] = widget
	}
	// Set its state provider, if it doesn't already have one
	if widget.StateProvider() == nil {
		widget.setStateProvider(m.StateProvider())
	}

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

// Have subsequent WidgetManager operations operate on a particular container widget.
// The container is marked for repainting.
func (m WidgetManager) With(container Widget) WidgetManager {
	cont := container.(GridWidget)
	id := cont.Id()

	CheckState(m.exists(id), "There is no widget with id:", id)

	// Discard any existing child widgets
	m.removeWidgets(cont.Children())
	cont.ClearChildren()
	m.initStateStack()
	m.PushContainer(container)
	m.resetPendingColumns()
	return m
}

func (m WidgetManager) resetPendingColumns() {
	m.stackedState().pendingChildColumns = MaxColumns
}

// Add a child GridContainerWidget, and push onto stack as active container
func (m WidgetManager) Open() Widget {
	widget := NewContainerWidget(m.ConsumeOptionalPendingId())
	m.Add(widget)
	return m.OpenContainer(widget)
}

// Push a container widget onto the stack as an active container
func (m WidgetManager) OpenContainer(widget Widget) Widget {
	itm := *m.stackedState()
	itm.Parent = widget
	m.pushState(itm, tag_container)
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

// The ButtonWidgetListener will receive message USER_HEADER_ACTION_xxxx.
func (m WidgetManager) AddUserHeader(listener ButtonWidgetListener) UserHeaderWidget {
	w := NewUserHeaderWidget(m.ConsumeOptionalPendingId(), listener)
	w.BgndImageMarkup = `style=" height:50px; background-image:url('app_header.jpg'); background-repeat: no-repeat;"`
	m.Add(w)
	return w
}

func (m WidgetManager) auxAddInput(listener InputWidgetListener, password bool) InputWidget {
	id := m.ConsumeOptionalPendingId()
	t := NewInputWidget(id, NewHtmlString(m.consumePendingLabel()), listener, password)
	m.Add(t)
	return t
}

func (m WidgetManager) AddPassword(listener InputWidgetListener) InputWidget {
	return m.auxAddInput(listener, true)
}

func (m WidgetManager) AddList(list ListInterface, itemWidget Widget) ListWidget {
	if !itemWidget.Visible() {
		BadArg("widget is not visible (detaching will happen by us)")
	}
	itemWidget.SetDetached(true)
	id := m.ConsumeOptionalPendingId()
	t := NewListWidget(id, list, itemWidget)
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
	id := m.ConsumeOptionalPendingId()
	return staticContent, id
}

func (m WidgetManager) AddHeading() HeadingWidget {
	staticContent, id := m.getStaticContentAndId()
	w := NewHeadingWidget(id)
	w.SetSize(m.consumePendingSize())
	Todo("!Setting WidgetSize seems to have no effect on headings")
	w.SetAlign(m.consumePendingAlign())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	m.Add(w)
	return w
}

func (m WidgetManager) AddText() TextWidget {
	staticContent, id := m.getStaticContentAndId()
	w := NewTextWidget(id, m.consumePendingSize(), m.consumePendingHeight())
	w.setStateProvider(m.StateProvider())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	m.Add(w)
	return w
}

func (m WidgetManager) AddButton(listener ButtonWidgetListener) ButtonWidget {
	w := NewButtonWidget(m.ConsumeOptionalPendingId(), listener)
	w.SetSize(m.consumePendingSize())
	w.SetAlign(m.consumePendingAlign())
	w.Label = NewHtmlString(m.consumePendingLabel())
	m.Add(w)
	return w
}

func (m WidgetManager) AddSpace() WidgetManager {
	return m.Add(NewBaseWidget(m.ConsumeOptionalPendingId()))
}

func doNothingFileUploadListener(s Session, widget FileUpload, value []byte) error {
	Pr("'do nothing' FileUploadListener called with bytes:", len(value))
	return nil
}

func (m WidgetManager) AddFileUpload(listener FileUploadWidgetListener) FileUpload {
	if listener == nil {
		listener = doNothingFileUploadListener
	}
	w := NewFileUpload(m.ConsumeOptionalPendingId(), NewHtmlString(m.consumePendingLabel()), listener)
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
	result := m.IdPrefix() + "z" + IntToString(m.anonymousIdCounter)
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
	if m.exists(id) {
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

func (m WidgetManager) PushContainer(container Widget) WidgetManager {
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

func (m WidgetManager) IdSummary() JSList {
	js := NewJSList()
	for k := range m.widgetMap {
		js.Add(k)
	}
	return js
}

func (m WidgetManager) RebuildPageWidget() Widget {
	m.widgetMap = make(map[string]Widget)
	m.Id(WidgetIdPage)
	widget := m.Open()
	m.Close()
	return widget
}
