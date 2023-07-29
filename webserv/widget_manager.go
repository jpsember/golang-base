package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"math/rand"
	"strings"
)

type WidgetManagerObj struct {
	BaseObject
	rand                        *rand.Rand
	widgetMap                   WidgetMap
	GrowXWeight                 int
	GrowYWeight                 int
	mPendingMinWidthEm          float64
	mPendingMinHeightEm         float64
	mPendingMonospaced          bool
	mLineCount                  int
	mComboChoices               *Array[string]
	mPendingBooleanDefaultValue bool
	mPendingStringDefaultValue  string
	mPendingTabTitle            string
	mPendingFloatingPointFlag   bool
	mPendingDefaultFloatValue   float64
	mPendingDefaultIntValue     int
	pendingListener             WidgetListener
	parentStack                 *Array[ContainerWidget]
	pendingSize                 int
	//pendingColumns              int
	pendingText        string
	pendingId          string
	pendingLabel       string
	anonymousIdCounter int
}

func NewWidgetManager() WidgetManager {
	w := WidgetManagerObj{
		parentStack: NewArray[ContainerWidget](),
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

func (m WidgetManager) find(id string) Widget {
	return m.widgetMap[id]
}

// ------------------------------------------------------------------
// Accessing widget values
// ------------------------------------------------------------------

/**
 * Set widgets' values. Used to restore app widgets to a previously saved
 * state
 */
func (m WidgetManager) SetWidgetValues(js *JSMapStruct) {
	for id, val := range js.WrappedMap() {
		if m.Exists(id) {
			m.Get(id).WriteValue(val)
		}
	}
}

/**
 * Read widgets' values. Doesn't include widgets that have no ids, or whose
 * ids start with "."
 */
func (m WidgetManager) ReadWidgetValues() *JSMapStruct {
	mp := NewJSMap()

	for id, w := range m.widgetMap {
		if strings.HasPrefix(id, ".") {
			continue
		}
		v := w.ReadValue()
		if v != nil {
			mp.Put(id, v)
		}
	}
	return mp
}

/**
 * Get value of string-valued widget
 */
func (m WidgetManager) Vs(id string) string {
	return m.Get(id).ReadValue().AsString()
}

/**
 * Set value of string-valued widget
 */
func (m WidgetManager) Sets(id string, v string) {
	m.Get(id).WriteValue(JString(v))
}

/**
 * Get value of boolean-valued widget
 */
func (m WidgetManager) Vb(id string) bool {
	result := false
	g := m.Get(id)
	if g != nil {
		result = g.ReadValue().AsBool()
	}
	return result
}

/**
 * Set value of boolean-valued widget
 */
func (m WidgetManager) Setb(id string, boolValue bool) bool {
	m.Get(id).WriteValue(JBool(boolValue))
	return boolValue
}

/**
 * Toggle value of boolean-valued widget
 *
 */
func (m WidgetManager) Toggle(id string) bool {
	return m.Setb(id, !m.Vb(id))
}

/**
 * Get value of integer-valued widget
 */
func (m WidgetManager) Vi(id string) int {
	return int(m.Get(id).ReadValue().AsInteger())
}

/**
 * Set value of integer-valued widget
 */
func (m WidgetManager) Seti(id string, v int) int {
	m.Get(id).WriteValue(JInteger(v))
	return v
}

/**
 * Get value of float-valued widget
 */
func (m WidgetManager) Vf(id string) float64 {
	return m.Get(id).ReadValue().AsFloat()
}

/**
 * Set value of double-valued widget
 */
func (m WidgetManager) SetF(id string, v float64) float64 {
	m.Get(id).WriteValue(JFloat(v))
	return v
}

// ------------------------------------------------------------------------------------

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

const (
	SIZE_DEFAULT = iota
	SIZE_TINY
	SIZE_SMALL
	SIZE_LARGE
	SIZE_HUGE
	SIZE_MEDIUM
)

const (
	ALIGNMENT_DEFAULT = iota
	ALIGNMENT_LEFT
	ALIGNMENT_CENTER
	ALIGNMENT_RIGHT
)

/**
 * Set pending component, and the column it occupies, as 'growable'. This can
 * also be accomplished by using an 'x' when declaring the columns.
 * <p>
 * Calls growX(100)...
 */
func (m WidgetManager) GrowX() WidgetManager {
	return m.GrowXBy(100)
}

/**
 * Set pending component, and the column it occupies, as 'growable'. This can
 * also be accomplished by using an 'x' when declaring the columns.
 * <p>
 * Calls growY(100)...
 */
func (m WidgetManager) GrowY() WidgetManager {
	return m.GrowYBy(100)
}

/**
 * Set pending component's horizontal weight to a value > 0 (if it is already
 * less than this value)
 */
func (m WidgetManager) GrowXBy(weight int) WidgetManager {
	m.GrowXWeight = MaxInt(m.GrowXWeight, weight)
	return m
}

/**
 * Set pending component's vertical weight to a value > 0 (if it is already
 * less than this value)
 */
func (m WidgetManager) GrowYBy(weight int) WidgetManager {
	m.GrowYWeight = MaxInt(m.GrowYWeight, weight)
	return m
}

/**
 * Specify the component to use for the next open() call, instead of
 * generating one
 */
func (m WidgetManager) SetPendingContainer(component any) WidgetManager {
	//
	//public WidgetManager setPendingContainer(JComponent component) {
	//  checkState(mPanelStack.isEmpty(), "current panel stack isn't empty");
	//  mPendingContainer = component;
	//  return m
	return m
}

// Set size for next widget (what size means depends upon the widget type).
func (m WidgetManager) Size(size int) WidgetManager {
	m.pendingSize = 1 + size
	return m
}

/**
 * Have next widget use a monospaced font
 */
func (m WidgetManager) Monospaced() WidgetManager {
	m.mPendingMonospaced = true
	return m
}

// Set number of Bootstrap columns for next widget
func (m WidgetManager) Col(columns int) WidgetManager {
	m.currentPanel().SetColumns(columns)
	return m
}

func (m WidgetManager) MinWidth(ems float64) WidgetManager {
	m.mPendingMinWidthEm = ems
	return m
}

func (m WidgetManager) MinHeight(ems float64) WidgetManager {
	m.mPendingMinHeightEm = ems
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

func (m WidgetManager) Text(value string) WidgetManager {
	m.pendingText = value
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

func (m WidgetManager) consumePendingText() string {
	lbl := m.pendingText
	m.pendingText = ""
	return lbl
}

func (m WidgetManager) consumePendingLabel() string {
	lbl := m.pendingLabel
	m.pendingLabel = ""
	return lbl
}

func (m WidgetManager) ConsumePendingSize() int {
	CheckState(m.pendingSize > 0, "no pending Size")
	size := m.pendingSize - 1
	m.pendingSize = 0
	return size
}

func (m WidgetManager) ConsumePendingStringDefaultValue() string {
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
	Todo("!incorporate skip values into 'BadState', other assertions")
	// If some values were not used, issue warnings
	verifyUsed(m.mPendingDefaultIntValue == 0, "pendingDefaultIntValue")
	verifyUsed(m.mPendingStringDefaultValue == "", "mPendingStringDefaultValue")
	verifyUsed(m.pendingText == "", "pendingText")
	verifyUsed(m.pendingLabel == "", "pendingLabel")
	verifyUsed(!m.mPendingFloatingPointFlag, "mPendingFloatingPoint")
	verifyUsed(m.pendingListener == nil, "pendingListener")
	verifyUsed(m.pendingSize == 0, "pendingSize")

	m.GrowXWeight = 0
	m.GrowYWeight = 0
	m.mPendingMinWidthEm = 0
	m.mPendingMinHeightEm = 0
	m.mPendingMonospaced = false
	m.mComboChoices = nil
	m.mPendingDefaultIntValue = 0
	m.mPendingBooleanDefaultValue = false
	m.mPendingStringDefaultValue = ""
	m.pendingText = ""
	m.mPendingFloatingPointFlag = false
}

/**
 * Add widget to the hierarchy
 */
func (m WidgetManager) Add(widget Widget) WidgetManager {
	b := widget.GetBaseWidget()
	id := b.GetId()
	if id != "" {
		if m.Exists(id) {
			BadState("Attempt to add widget with duplicate id:", id)
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

/**
 * Create a child container widget and push onto stack
 */
func (m WidgetManager) Open() Widget {
	m.Log("open")
	// the number of columns a widget is to occupy should be sent to the *parent*...
	widget := NewContainerWidget(m.consumeOptionalPendingId())
	m.Log("Adding container widget")
	m.Add(widget)
	m.parentStack.Add(widget)
	m.Log("added container to stack")
	return widget
}

/**
 * Pop view from the stack
 */
func (m WidgetManager) Close() WidgetManager {
	return m.CloseFor("<no context>")
}

// Pop view from stack
// Deprecated. Get rid of debugContext.
func (m WidgetManager) CloseFor(debugContext string) WidgetManager {
	m.Log("Close", debugContext)
	parent := m.parentStack.Pop()
	//m.EndRow()
	parent.LayoutChildren(m)
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

func (m WidgetManager) currentPanel() ContainerWidget {
	if m.parentStack.IsEmpty() {
		BadState("no current panel")
	}
	return m.parentStack.Last()
}

func (m WidgetManager) AddInput(id string) WidgetManager {
	t := NewInputWidget(id, NewHtmlString(m.consumePendingLabel()))
	m.assignPendingListener(t)
	return m.Add(t)
}

func (m WidgetManager) AddHeading(id string) WidgetManager {
	t := NewHeadingWidget(id, m.ConsumePendingSize())
	return m.Add(t)
}

func (m WidgetManager) assignPendingListener(widget Widget) {
	if m.pendingListener != nil {
		b := widget.GetBaseWidget()
		CheckState(widget.GetBaseWidget().Listener == nil, "Widget", b.GetId(), "already has a listener")
		widget.GetBaseWidget().Listener = m.pendingListener
		m.pendingListener = nil
	}
}

func (m WidgetManager) AddText() WidgetManager {

	var w TextWidget
	// The text can either be expressed as a string (static content),
	// or an id (dynamic content, read from session state)
	staticContent := m.consumePendingText()
	hasStaticContent := staticContent != ""
	if hasStaticContent {
		CheckState(m.pendingId == "", "specify id OR static content")
	}
	id := m.consumeOptionalPendingId()
	w = NewTextWidget(id)
	if hasStaticContent {
		w.SetStaticContent(staticContent)
	}
	m.Log("Adding text, id:", w.Id)
	return m.Add(w)
}

func (m WidgetManager) AddButton() ButtonWidget {
	w := NewButtonWidget()
	w.Id = m.consumePendingId()
	m.assignPendingListener(w)
	m.Log("Adding button, id:", w.Id)
	w.Label = NewHtmlString(m.consumePendingText())
	m.Add(w)
	return w
}

func (m WidgetManager) AddDebug() WidgetManager {
	Alert("!<1 Adding DebugWidget")
	w := NewDebugWidget(m.consumeOptionalPendingId())
	m.Add(w)

	return m
}

func (m WidgetManager) AddCheckbox() CheckboxWidget {
	return m.checkboxHelper(false)
}

func (m WidgetManager) AddSwitch() CheckboxWidget {
	return m.checkboxHelper(true)
}

func (m WidgetManager) checkboxHelper(switchFlag bool) CheckboxWidget {
	w := NewCheckboxWidget(switchFlag, m.consumePendingId(), NewHtmlString(m.consumePendingText()))
	m.assignPendingListener(w)
	m.Add(w)
	return w
}

func (m WidgetManager) Listener(listener WidgetListener) WidgetManager {
	m.pendingListener = listener
	return m
}

func (m WidgetManager) AllocateAnonymousId() string {
	m.anonymousIdCounter++
	return "." + IntToString(m.anonymousIdCounter)
}

var DebugColorsFlag bool
var DebugWidgetBounds = false

// Deprecated. To have uses show up in editor as a warning.
func SetDebugColors() {
	Alert("<1 Setting debug colors")
	DebugColorsFlag = true
	Todo("!<1 Have debug colors affect all widgets, not just debug_widget")
}

// Deprecated. To have uses show up in editor as a warning.
func SetDebugWidgetBounds() {
	Alert("<1 Setting debug widget bounds")
	if !Alert("probably no longer useful") {
		DebugWidgetBounds = true
	}
}
