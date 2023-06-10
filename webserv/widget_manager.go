package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"math/rand"
	"strings"
)

type WidgetManagerObj struct {
	BaseObject
	rand                        *rand.Rand
	widgetMap                   WidgetMap
	GrowXWeight                 int
	GrowYWeight                 int
	mPendingSize                int
	mPendingAlignment           int
	mPendingGravity             int
	mPendingMinWidthEm          float64
	mPendingMinHeightEm         float64
	mPendingMonospaced          bool
	mLineCount                  int
	mComboChoices               *Array[string]
	mPendingBooleanDefaultValue bool
	mPendingStringDefaultValue  string
	mPendingLabel               string
	mPendingTabTitle            string
	mPendingFloatingPointFlag   bool
	mPendingDefaultFloatValue   float64
	mPendingDefaultIntValue     int
	pendingListener             WidgetListener
	parentStack                 *Array[ContainerWidget]

	pendingColumns int
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

var digitsExpr = Regexp(`^\d+$`)

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

func (m WidgetManager) setPendingSize(value int) WidgetManager {
	m.mPendingSize = value
	return m
}

func (m WidgetManager) setPendingAlignment(value int) WidgetManager {
	m.mPendingAlignment = value
	return m
}

func (m WidgetManager) Small() WidgetManager {
	return m.setPendingSize(SIZE_SMALL)
}

func (m WidgetManager) Large() WidgetManager {
	return m.setPendingSize(SIZE_LARGE)
}

func (m WidgetManager) medium() WidgetManager {
	return m.setPendingSize(SIZE_MEDIUM)
}

func (m WidgetManager) tiny() WidgetManager {
	return m.setPendingSize(SIZE_TINY)
}

func (m WidgetManager) huge() WidgetManager {
	return m.setPendingSize(SIZE_HUGE)
}

func (m WidgetManager) left() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_LEFT)
}

func (m WidgetManager) right() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_RIGHT)
}

func (m WidgetManager) center() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_CENTER)
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
	m.pendingColumns = columns
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

func (m WidgetManager) gravity(gravity int) WidgetManager {
	m.mPendingGravity = gravity
	return m
}

func (m WidgetManager) LineCount(numLines int) WidgetManager {
	CheckArg(numLines > 0)
	m.mLineCount = numLines
	return m
}

func (m WidgetManager) addLabel(id string) WidgetManager {
	text := m.ConsumePendingLabel()
	Todo("addLabel", text)
	//add(new LabelWidget(id, mPendingGravity, mLineCount, text, mPendingSize, mPendingMonospaced,
	//    mPendingAlignment));
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
	m.mPendingLabel = value
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

func (m WidgetManager) ConsumePendingLabel() string {
	lbl := m.mPendingLabel
	m.mPendingLabel = ""
	return lbl
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
	// If some values were not used, issue warnings
	//verifyUsed(mComboChoices, "pending combo choices");
	verifyUsed(m.mPendingDefaultIntValue == 0, "pendingDefaultIntValue")
	verifyUsed(m.mPendingStringDefaultValue == "", "mPendingStringDefaultValue")
	verifyUsed(m.mPendingLabel == "", "mPendingLabel ")
	verifyUsed(!m.mPendingFloatingPointFlag, "mPendingFloatingPoint")
	verifyUsed(m.pendingListener == nil, "pendingListener")

	m.GrowXWeight = 0
	m.GrowYWeight = 0
	m.mPendingSize = SIZE_DEFAULT
	m.mPendingAlignment = ALIGNMENT_DEFAULT
	m.mPendingGravity = 0
	m.mPendingMinWidthEm = 0
	m.mPendingMinHeightEm = 0
	m.mPendingMonospaced = false
	m.mLineCount = 0
	m.mComboChoices = nil
	m.mPendingDefaultIntValue = 0
	m.mPendingBooleanDefaultValue = false
	m.mPendingStringDefaultValue = ""
	m.mPendingLabel = ""
	m.mPendingFloatingPointFlag = false
}

func (m WidgetManager) Open(id string) Widget {
	return m.OpenFor(id, "<no context>")
}

/**
 * Add widget to the hierarchy
 */
func (m WidgetManager) Add(widget Widget) WidgetManager {
	id := widget.GetId()
	if id != "" {
		if m.Exists(id) {
			BadState("Attempt to add widget with duplicate id:", id)
		}
		m.widgetMap[id] = widget
	}

	m.Log("addWidget, id:", widget.GetId(), "panel stack size:", m.parentStack.Size())
	if !m.parentStack.IsEmpty() {
		m.parentStack.Last().AddChild(widget, m)

	}
	m.clearPendingComponentFields()
	return m
}

/**
 * Create a child widget and push onto stack
 */
func (m WidgetManager) OpenFor(id string, debugContext string) Widget {
	Todo("support ids for these containers")
	m.Log("openFor:", debugContext)

	if m.pendingColumns == 0 {
		Todo("default to previous columns?")
		m.pendingColumns = 4
	}

	// the number of columns a widget is to occupy should be sent to the *parent*...

	widget := NewContainerWidget(id, m)
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

/**
 * Pop view from the stack
 */
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

func (m WidgetManager) AddText(id string) WidgetManager {
	t := NewInputWidget(id, m.mPendingSize)
	m.assignPendingListener(t)
	//TextWidget t = new TextWidget(consumePendingListener(), id, consumePendingStringDefaultValue(),
	//    mLineCount, mEditableFlag, mPendingSize, mPendingMonospaced, mPendingMinWidthEm, mPendingMinHeightEm);
	//consumeTooltip(t);
	return m.Add(t)
}

func (m WidgetManager) assignPendingListener(widget Widget) {
	if m.pendingListener != nil {
		CheckState(widget.GetBaseWidget().Listener == nil, "Widget", widget.GetId(), "already has a listener")
		widget.GetBaseWidget().Listener = m.pendingListener
		m.pendingListener = nil
	}
}

//func (m WidgetManager)  AddHeader(text string ) WidgetManager {
//  m.spanx();
//  JLabel label = new JLabel(text);
//  label.setBorder(
//      new CompoundBorder(buildStandardBorderWithZeroBottom(), BorderFactory.createEtchedBorder()));
//  label.setHorizontalAlignment(SwingConstants.CENTER);
//  add(wrap(label));
//  return m;
//}

/**
* Add a horizontal space to occupy cell(s) in place of other widgets
 */
func (m WidgetManager) AddHorzSpace() WidgetManager {
	return m.Add(NewPanelWidget())
}

///**
// * Add a horizontal separator that visually separates components above from
// * below
// */
//func (m WidgetManager)  AddHorzSep( ) WidgetManager {
//  m.spanx();
//  m.add(wrap(new JSeparator(JSeparator.HORIZONTAL)));
//  return m
//}

///**
// * Add a vertical separator that visually separates components left from right
// */
//func (m WidgetManager)  AddVertSep( ) WidgetManager {
// m. spanx();
// m. growY();
// m. add(m.wrap(new JSeparator(JSeparator.VERTICAL)));
//  return m
//}

///**
// * Add a row that can stretch vertically to occupy the available space
// */
//func (m WidgetManager)  AddVertGrow( ) WidgetManager {
//  //JComponent panel;
//  //if (verbose())
//  //  panel = colorPanel();
//  //else
//  //  panel = new JPanel();
//  //spanx().growY();
//  //add(wrap(panel));
//  return m
//}

//func (m WidgetManager) AddButton ( id string) WidgetManager {
//  ButtonWidget button = new ButtonWidget(consumePendingListener(), id, consumePendingLabel(true));
//  return add(button);
//}

//func (m WidgetManager) AddToggleButton (id string ) WidgetManager {
//  ToggleButtonWidget button = new ToggleButtonWidget(consumePendingListener(), id,
//      consumePendingLabel(true), consumePendingBooleanDefaultValue());
//  return add(button);
//}

func (m WidgetManager) AddLabel(id string) WidgetManager {
	text := m.ConsumePendingLabel()
	w := NewLabelWidget()
	w.Id = id
	w.LineCount = m.mLineCount
	w.Text = text
	w.Size = m.mPendingSize
	w.Monospaced = m.mPendingMonospaced
	w.Alignment = m.mPendingAlignment
	m.Log("Adding label, id:", id)
	return m.Add(w)
}

//
//
//func (m WidgetManager) AddChoiceBox (id string ) WidgetManager {
//  ComboBoxWidget c = new ComboBoxWidget(consumePendingListener(), id, mComboChoices);
//  return add(c);
//}

func (m WidgetManager) Listener(listener WidgetListener) WidgetManager {
	m.pendingListener = listener
	return m
}
