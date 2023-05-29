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
	widgetMap map[string]Widget
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
