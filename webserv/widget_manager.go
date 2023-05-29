package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"strings"
)

type WidgetObj struct {
	Id string
}

type Widget = *WidgetObj

func (w Widget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
}

func (w Widget) ReadValue() JSEntity {
	NotImplemented("ReadValue")
	return JBoolFalse
}

type WidgetManagerObj struct {
	pendingColumnWeights []int
	// Note: this was a sorted map in the Java code
	widgetMap   map[string]Widget
	mSpanXCount int

	mGrowXFlag                  int
	mGrowYFlag                  int
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
func (m WidgetManager) SetWidgetValues(js *JSMap) {
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
func (m WidgetManager) ReadWidgetValues() *JSMap {
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
	return m.Get(id).ReadValue().ToString()
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
		result = g.ReadValue().ToBool()
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
	return int(m.Get(id).ReadValue().ToInteger())
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
	return m.Get(id).ReadValue().ToFloat()
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

/**
 * <pre>
 *
 * Set the number of columns, and which ones can grow, for the next view in
 * the hierarchy. The columns expression is a string of column expressions,
 * which may be one of:
 *
 *     "."   a column with weight zero
 *     "x"   a column with weight 100
 *     "\d+" column with integer weight
 *
 * Spaces are ignored, except to separate integer weights from each other.
 * </pre>
 */
func (m WidgetManager) Columns(columnsExpr string) WidgetManager {
	CheckState(m.pendingColumnWeights == nil, "previous column weights were never used")

	columnSizes := NewArray[int]()
	for _, word := range strings.Split(columnsExpr, " ") {
		size := 0
		if digitsExpr.MatchString(word) {
			size = ParseIntM(word)
			columnSizes.Add(size)
		} else {
			w := []byte(word)
			for _, c := range w {
				if c == '.' {
					size = 0
				} else if c == 'x' {
					size = 100
				} else {
					BadArg("Can't parse columns expression:", Quoted(columnsExpr))
				}
				columnSizes.Add(size)
			}
		}
	}
	m.pendingColumnWeights = columnSizes.Array()
	return m
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
 * Make next component added occupy remaining columns in its row
 */
func (m WidgetManager) Spanx() WidgetManager {
	m.mSpanXCount = -1
	return m
}

/**
 * Make next component added occupy some number of columns in its row
 */
func (m WidgetManager) SpanxCount(count int) WidgetManager {
	CheckArg(count > 0)
	m.mSpanXCount = count
	return m
}

/**
 * Skip a single cell
 */
func (m WidgetManager) skip() WidgetManager {
	m.add(m.wrap(nil))
	return m
}

/**
 * Skip one or more cells
 */
func (m WidgetManager) skipN(count int) WidgetManager {
	m.SpanxCount(count)
	m.add(m.wrap(nil))
	return m
}

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
	m.mGrowXFlag = MaxInt(m.mGrowXFlag, weight)
	return m
}

/**
 * Set pending component's vertical weight to a value > 0 (if it is already
 * less than this value)
 */
func (m WidgetManager) GrowYBy(weight int) WidgetManager {
	m.mGrowYFlag = MaxInt(m.mGrowYFlag, weight)
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
	//  return this;
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
	verifyUsed(0 != len(m.pendingColumnWeights), "pending column weights")
	//verifyUsed(mComboChoices, "pending combo choices");
	verifyUsed(m.mPendingDefaultIntValue == 0, "pendingDefaultIntValue")
	verifyUsed(m.mPendingStringDefaultValue != "", "mPendingStringDefaultValue")
	verifyUsed(m.mPendingLabel != "", "mPendingLabel ")
	verifyUsed(!m.mPendingFloatingPointFlag, "mPendingFloatingPoint")

	m.pendingColumnWeights = nil
	m.mSpanXCount = 0
	m.mGrowXFlag = 0
	m.mGrowYFlag = 0
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

/**
func (m WidgetManager)  ( ) WidgetManager {
**/
