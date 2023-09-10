package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type WidgetManagerObj struct {
	BaseObject
	widgetMap                   WidgetMap
	mComboChoices               *Array[string]
	mPendingBooleanDefaultValue bool
	mPendingStringDefaultValue  string
	mPendingTabTitle            string
	mPendingFloatingPointFlag   bool
	mPendingDefaultFloatValue   float64
	mPendingDefaultIntValue     int
	//pendingListener             WidgetListener
	parentStack        *Array[Widget]
	pendingSize        WidgetSize
	pendingAlign       WidgetAlign
	pendingId          string
	pendingLabel       string
	anonymousIdCounter int
	repaintSet         StringSet
}

func NewWidgetManager(session Session) WidgetManager {
	w := WidgetManagerObj{
		parentStack: NewArray[Widget](),
		widgetMap:   make(map[string]Widget),
	}
	w.SetName("WidgetManager")
	return &w
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
	m.pendingId = id
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
		id = m.AllocateAnonymousId()
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

// Set number of Bootstrap columns for next widget
func (m WidgetManager) Col(columns int) WidgetManager {
	w := m.currentPanel()
	c, ok := w.(ContainerWidget)
	CheckState(ok)
	c.SetColumns(columns)
	return m
}

func (m WidgetManager) Floats() WidgetManager {
	m.mPendingFloatingPointFlag = true
	return m
}

/**
 * Set default value for next boolean-valued control
 */
func (m WidgetManager) DefaultBool(value bool) WidgetManager {
	m.mPendingBooleanDefaultValue = value
	return m
}

func (m WidgetManager) DefaultString(value string) WidgetManager {
	m.mPendingStringDefaultValue = value
	return m
}

func (m WidgetManager) Label(value string) WidgetManager {
	CheckState(m.pendingLabel == "")
	m.pendingLabel = value
	return m
}

/**
 * Set default value for next double-valued control
 */
func (m WidgetManager) defaultFloat(value float64) WidgetManager {
	m.Floats()
	m.mPendingDefaultFloatValue = value
	return m
}

/**
 * Set default value for next integer-valued control
 */
func (m WidgetManager) defaultInt(value int) WidgetManager {
	m.mPendingDefaultIntValue = value
	return m
}

/**
 * Append some choices for the next ComboBox
 */
func (m WidgetManager) Choices(choices ...string) WidgetManager {
	for _, s := range choices {
		if m.mComboChoices == nil {
			m.mComboChoices = NewArray[string]()
		}
		m.mComboChoices.Add(s)
	}
	return m
}

func (m WidgetManager) ConsumePendingBooleanDefaultValue() bool {
	v := m.mPendingBooleanDefaultValue
	m.mPendingBooleanDefaultValue = false
	return v
}

func (m WidgetManager) ConsumePendingFloatingPointFlag() bool {
	v := m.mPendingFloatingPointFlag
	m.mPendingFloatingPointFlag = false
	return v
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

func (m WidgetManager) consumePendingStringDefaultValue() string {
	s := m.mPendingStringDefaultValue
	m.mPendingStringDefaultValue = ""
	return s
}

func (m WidgetManager) consumePendingTabTitle() string {
	tabNameExpression := m.mPendingTabTitle
	m.mPendingTabTitle = ""
	CheckState(tabNameExpression != "", "no pending tab title")
	return tabNameExpression
}

func verifyUsed(flag bool, name string) {
	if flag {
		return
	}
	BadState("unused value:", name)
}

func (m WidgetManager) clearPendingComponentFields() {
	// If some values were not used, issue warnings
	verifyUsed(m.mPendingDefaultIntValue == 0, "pendingDefaultIntValue")
	verifyUsed(m.mPendingStringDefaultValue == "", "mPendingStringDefaultValue")
	verifyUsed(m.pendingLabel == "", "pendingLabel")
	verifyUsed(!m.mPendingFloatingPointFlag, "mPendingFloatingPoint")
	//verifyUsed(m.pendingListener == nil, "pendingListener")
	verifyUsed(m.pendingSize == SizeDefault, "pendingSize")
	verifyUsed(m.pendingAlign == AlignDefault, "pendingAlign")

	m.mComboChoices = nil
	m.mPendingDefaultIntValue = 0
	m.mPendingBooleanDefaultValue = false
	m.mPendingStringDefaultValue = ""
	m.mPendingFloatingPointFlag = false
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

	m.Log("addWidget, id:", id, "panel stack size:", m.parentStack.Size())
	if !m.parentStack.IsEmpty() {
		m.parentStack.Last().AddChild(widget, m)
	}
	m.clearPendingComponentFields()
	return m
}

// Have subsequent WidgetManager operations operate on a particular container widget.
// The container is marked for repainting.
func (m WidgetManager) With(container Widget) WidgetManager {
	pr := PrIf(false)
	id := container.Id()
	pr(VERT_SP, "With:", id, "at:", CallerLocation(1))

	CheckState(m.Exists(id))

	m.Repaint(container)

	pr("current widget map:", INDENT, m.WidgetMapSummary())
	pr("removing all child widgets")
	// Discard any existing child widgets
	m.removeWidgets(container.Children())
	cont := container.(ContainerWidget)
	cont.ClearChildren()

	pr("after removal, current widget map:", INDENT, m.WidgetMapSummary())

	m.parentStack.Clear()
	m.parentStack.Add(container)
	return m
}

// Create a child container widget and push onto stack
func (m WidgetManager) Open() Widget {
	m.Log("open")
	// the number of columns a widget is to occupy should be sent to the *parent*...
	widget := NewContainerWidget(m.consumeOptionalPendingId())
	m.OpenContainer(widget)
	return widget
}

// Push a container widget onto the stack
func (m WidgetManager) OpenContainer(widget Widget) {
	m.Log("Adding container widget")
	m.Add(widget)
	m.parentStack.Add(widget)
	m.Log("added container to stack")
}

// Pop a container widget from the stack.
func (m WidgetManager) Close() WidgetManager {
	m.Log("Close")
	m.parentStack.Pop()
	return m
}

// Verify that no unused 'pending' arguments exist, calls are balanced, etc
func (m WidgetManager) finish() WidgetManager {
	m.clearPendingComponentFields()
	if !m.parentStack.IsEmpty() {
		BadState("panel stack nonempty; size:", m.parentStack.Size())
	}
	return m
}

func (m WidgetManager) currentPanel() Widget {
	if m.parentStack.IsEmpty() {
		BadState("no current panel")
	}
	return m.parentStack.Last()
}

func (m WidgetManager) AddInput(listener InputWidgetListener) WidgetManager {
	return m.auxAddInput(listener, false)
}

func (m WidgetManager) auxAddInput(listener InputWidgetListener, password bool) WidgetManager {
	id := m.consumeOptionalPendingId()
	t := NewInputWidget(id, NewHtmlString(m.consumePendingLabel()), listener, password)
	m.assignPendingListener(t)
	return m.Add(t)
}

func (m WidgetManager) AddPassword(listener InputWidgetListener) WidgetManager {
	return m.auxAddInput(listener, true)
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
	w.SetAlign(m.consumePendingAlign())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	return m.Add(w)
}

func (m WidgetManager) assignPendingListener(widget Widget) {
	Alert("<1#50No longer supported, assignPendingListener() method")
	//
	//if m.pendingListener != nil {
	//	m.assignRequiredPendingListener(widget)
	//}
}

//
//func (m WidgetManager) assignRequiredPendingListener(widget Widget) {
//	if m.pendingListener == nil {
//		BadState("no listener found for widget", widget.Id())
//	}
//	widget.SetListener(m.pendingListener)
//	m.pendingListener = nil
//}

func (m WidgetManager) AddText() WidgetManager {
	staticContent, id := m.getStaticContentAndId()
	w := NewTextWidget(id, m.consumePendingSize())
	if staticContent != "" {
		w.SetStaticContent(staticContent)
	}
	m.Log("Adding text, id:", w.BaseId)
	return m.Add(w)
}

func (m WidgetManager) AddButton(listener ButtonWidgetListener) ButtonWidget {
	w := NewButtonWidget(listener)
	w.BaseId = m.consumeOptionalPendingId()
	w.SetSize(m.consumePendingSize())
	w.SetAlign(m.consumePendingAlign())
	m.assignPendingListener(w)
	m.Log("Adding button, id:", w.BaseId)
	w.Label = NewHtmlString(m.consumePendingLabel())
	m.Add(w)
	return w
}

func (m WidgetManager) AddSpace() WidgetManager {
	return m.Add(NewBaseWidget(m.consumeOptionalPendingId()))
}

func (m WidgetManager) AddFileUpload() FileUpload {
	Todo("add listener to FileUpload Widget")
	w := NewFileUpload(m.consumePendingId(), NewHtmlString(m.consumePendingLabel()))
	//m.assignRequiredPendingListener(w)
	m.Add(w)
	return w
}

func (m WidgetManager) AddImage() ImageWidget {
	w := NewImageWidget(m.consumePendingId())
	Pr("created image widget, id:", w.Id())
	m.assignPendingListener(w)
	m.Add(w)
	return w
}

func (m WidgetManager) AddCheckbox() CheckboxWidget {
	return m.checkboxHelper(false)
}

func (m WidgetManager) AddSwitch() CheckboxWidget {
	return m.checkboxHelper(true)
}

func (m WidgetManager) checkboxHelper(switchFlag bool) CheckboxWidget {
	w := NewCheckboxWidget(switchFlag, m.consumePendingId(), NewHtmlString(m.consumePendingLabel()))
	m.assignPendingListener(w)
	m.Add(w)
	return w
}

func (m WidgetManager) Listener(listener any) WidgetManager {
	Alert("<1#50No longer supported, Listener() method")
	//m.pendingListener = listener
	return m
}

//func (m WidgetManager) Listener(listener WidgetListener) WidgetManager {
//	m.pendingListener = listener
//	return m
//}

func (m WidgetManager) AllocateAnonymousId() string {
	m.anonymousIdCounter++
	return "." + IntToString(m.anonymousIdCounter)
}

func (m WidgetManager) removeWidgets(widgets *Array[Widget]) {
	for _, widget := range widgets.Array() {
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

func GetStaticOrDynamicLabel(widget Widget, state JSMap) (string, bool) {
	Todo("?we are assuming static content is a string here")
	sc := widget.StaticContent()
	if sc != nil {
		return sc.(string), true
	} else {
		return WidgetStringValue(state, widget.Id()), false
	}
}

// Mark a widget for repainting.  Does nothing if there is no repaintSet (i.e., it is not being done within
// an AJAX call)
func (m WidgetManager) Repaint(w Widget) WidgetManager {
	if m.repaintSet != nil {

		pr := PrIf(debRepaint)
		pr("Repaint:", w)
		if m.repaintSet.Add(w.Id()) {
			pr("...adding to set")
		}
	}
	return m
}

// Mark widgets for repainting (if they exist).  Does nothing if there is no repaintSet.
func (m WidgetManager) RepaintIds(ids ...string) WidgetManager {
	for _, id := range ids {
		w := m.Opt(id)
		Pr("repaint:", id, "found:", w)
		if w != nil {
			m.Repaint(w)
		} else {
			Alert("#50<1Can't find widget to repaint with id:", id)
		}
	}
	return m
}

func (m WidgetManager) clearRepaintSet() {
	m.repaintSet = NewStringSet()
}
