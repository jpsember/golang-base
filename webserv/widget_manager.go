package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type mgrState struct {
	Parent        Widget
	StateProvider JSMap
	// Not to be confused with stackedStateProvider.Prefix, this is an optional prefix to be prepended to the
	// id when adding widgets:
	IdPrefix            string
	DebugTag            string
	pendingChildColumns int
	clickTargetPrefix   string // prefix to add to onpress arguments
}

var mgrStateDefault = mgrState{
	DebugTag: "<bottom of stack>",
}

type WidgetManagerObj struct {
	widgetMap            WidgetMap
	stack                []mgrState
	pendingSize          WidgetSize
	pendingAlign         WidgetAlign
	pendingHeight        int
	pendingId            string
	pendingLabel         string
	pendingClickListener ButtonWidgetListener
	anonymousIdCounter   int
}

func (m WidgetManager) InitializeWidgetManager() {
	DebVerifyServerStarted()
	m.widgetMap = make(map[string]Widget)
	m.initStateStack()
	m.resetPendingColumns()
}

func (m WidgetManager) initStateStack() {
	// Add a sentinel value so the stack is never empty
	m.stack = []mgrState{mgrStateDefault}
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
	Todo("!this name conflicts with Session.SessionId")
	v := m.IdPrefix() + id
	AssertNoDots(v)
	m.pendingId = v
	return m
}

func (m WidgetManager) ConsumeOptionalPendingClickListener() ButtonWidgetListener {
	listener := m.pendingClickListener
	m.pendingClickListener = nil
	return listener
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

func (m WidgetManager) Listener(listener ButtonWidgetListener) WidgetManager {
	m.pendingClickListener = listener
	return m
}

func (m WidgetManager) clearPendingComponentFields() {
	// If some values were not used, issue warnings
	verifyUsed(m.pendingLabel == "", "pendingLabel")
	verifyUsed(m.pendingSize == SizeDefault, "pendingSize")
	verifyUsed(m.pendingAlign == AlignDefault, "pendingAlign")
	verifyUsed(m.pendingHeight == 0, "pendingHeight")
	verifyUsed(m.pendingClickListener == nil, "pendingClickListener")
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

	widget.setStateProvider(m.stackedStateProvider())

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

func (m WidgetManager) DebugStackedState() *mgrState { return m.stackedState() }

func (m WidgetManager) stackedState() *mgrState {
	return &m.stack[len(m.stack)-1]
}

func (s *mgrState) String() string {
	return NewJSMap().Put("IdPrefix", s.IdPrefix).Put("StateProvider", s.StateProvider.String()).Put("clickpref", s.clickTargetPrefix).CompactString()
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
	widget := NewContainerWidget(m.ConsumeOptionalPendingId(), m.ConsumeOptionalPendingClickListener())
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

const (
	tag_container   = "container"
	tag_prefix      = "prefix"
	tag_provider    = "provider"
	tag_editor      = "editor"
	tag_clickprefix = "clickprefix"
)

var debStack = false && Alert("debug stack")

// Pop the active container from the stack.
func (m WidgetManager) Close() WidgetManager {
	m.popStack(tag_container)
	return m
}

func (m WidgetManager) pushState(state mgrState, tag string) {
	if debStack {
		Pr(Callers(1, 4))
		Pr("pushState:", tag, INDENT, m.dumpStateStack(len(m.stack)))
	}
	state.DebugTag = tag
	m.stack = append(m.stack, state)
}

func (m WidgetManager) popStack(tag string) {
	if debStack {
		Pr(Callers(1, 4))
		Pr("popStack:", tag, INDENT, m.dumpStateStack(len(m.stack)))
	}
	top := m.stackedState()
	if top.DebugTag != tag {
		BadState("attempt to pop state stack, tag is:", top.DebugTag, "but expected:", tag)
	}
	_, m.stack = PopLast(m.stack)
}

func (m WidgetManager) StateStackToJson() JSMap {
	result := NewJSMap()
	for _, x := range m.stack {
		mp := NewJSMap()
		if x.IdPrefix != "" {
			mp.Put("id_prefix", x.IdPrefix)
		}
		mp.Put("debug_tag", x.DebugTag)
		if x.clickTargetPrefix != "" {
			mp.Put("click_pref", x.clickTargetPrefix)
		}
		m2 := x.StateProvider
		if m2 != nil {
			mp.Put("state_prov", m2)
		}
		result.PutNumbered(mp)
	}
	return result
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
	m.PushStateMap(NewJSMap())
	w := NewUserHeaderWidget(m.ConsumeOptionalPendingId(), listener)
	w.BgndImageMarkup = `style=" height:50px; background-image:url('app_header.jpg'); background-repeat: no-repeat;"`
	m.Add(w)
	m.PopStateProvider()
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

func (m WidgetManager) AddList(list ListInterface, itemWidget Widget, listener ListWidgetListener) ListWidget {
	if !itemWidget.Visible() {
		BadArg("widget is not visible (detaching will happen by us)")
	}
	itemWidget.SetDetached(true)
	Alert("!The list item subwidgets are not being detached along with the item widget; but maybe we don't care")
	id := m.ConsumeOptionalPendingId()
	t := NewListWidget(id, list, itemWidget, listener)
	m.Add(t)
	return t
}

// Utility method to determine the label and id for text fields (text fields, headings).
// The label can either be expressed as a string (static content),
// or an id (dynamic content, read from session state).  If static, there should *not* be
// a pending id.
func (m WidgetManager) getStaticContentAndId() (string, string, bool) {
	staticContent := m.consumePendingLabel()
	hasStaticContent := staticContent != ""
	if hasStaticContent {
		CheckState(m.pendingId == "", "specify id OR static content")
	}
	id := m.ConsumeOptionalPendingId()
	return staticContent, id, hasStaticContent
}

func (m WidgetManager) AddHeading() HeadingWidget {
	staticContent, id, wasStatic := m.getStaticContentAndId()
	w := NewHeadingWidget(id)
	w.SetSize(m.consumePendingSize())
	Todo("!Setting WidgetSize seems to have no effect on headings")
	w.SetAlign(m.consumePendingAlign())
	if wasStatic {
		w.SetStaticContent(staticContent)
	}
	m.Add(w)
	return w
}

func (m WidgetManager) AddText() TextWidget {
	staticContent, id, wasStatic := m.getStaticContentAndId()
	w := NewTextWidget(id, m.consumePendingSize(), m.consumePendingHeight())
	if wasStatic {
		w.SetStaticContent(staticContent)
	}
	m.Add(w)
	return w
}

func (m WidgetManager) AddBtn() ButtonWidget {
	w := NewButtonWidget(m.ConsumeOptionalPendingId(), m.ConsumeOptionalPendingClickListener())
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

func (m WidgetManager) AddImage(urlProvider ImageURLProvider) ImageWidget {
	w := NewImageWidget(m.consumePendingId(), urlProvider, m.ConsumeOptionalPendingClickListener())
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
	result := m.IdPrefix() + "z_" + IntToString(m.anonymousIdCounter)
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

func (m WidgetManager) PushStateMap(jsmap JSMap) {
	m.PushStateProvider(jsmap)
}

func (m WidgetManager) PushEditor(editor DataEditor) {
	itm := *m.stackedState()
	itm.IdPrefix = editor.Prefix
	itm.StateProvider = editor.JSMap
	m.pushState(itm, tag_editor)
}

func (m WidgetManager) PopEditor() {
	m.popStack(tag_editor)
}

// Deprecated.
func (m WidgetManager) PushStateProvider(p JSMap) {
	Pr("Use PushStateMap instead")
	itm := *m.stackedState()
	itm.StateProvider = p
	m.pushState(itm, tag_provider)
}

func (m WidgetManager) PopStateProvider() {
	m.popStack(tag_provider)
}

func (m WidgetManager) PopStateMap() {
	m.PopStateProvider()
}

func (m WidgetManager) stackedStateProvider() JSMap {
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

func (m WidgetManager) PushClickPrefix(prefix string) {
	itm := *m.stackedState()
	itm.clickTargetPrefix = prefix + itm.clickTargetPrefix
	m.pushState(itm, tag_clickprefix)
}

func (m WidgetManager) PopClickPrefix() {
	m.popStack(tag_clickprefix)
}

func (m WidgetManager) ClickPrefix() string {
	return m.stackedState().clickTargetPrefix
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
