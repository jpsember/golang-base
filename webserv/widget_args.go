// A data structure for parsing widget arguments (colon-separated strings, e.g. "hotel_list:page:3")

package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type WidgetArgsStruct struct {
	text   string
	delim  []int
	cursor int
}

type WidgetArgs = *WidgetArgsStruct

func NewWidgetArgs(text string) WidgetArgs {
	t := &WidgetArgsStruct{
		text: text,
	}
	d := []int{-1}

	i := 0
	for i < len(text) {
		if text[i] == ':' {
			d = append(d, i)
		}
		i++
	}
	t.delim = append(d, i)
	return t
}

func (w WidgetArgs) Count() int {
	return len(w.delim) - 1
}

func (w WidgetArgs) Done() bool { return w.cursor == w.Count() }

func (w WidgetArgs) Arg(i int) string {
	if i < 0 || i >= w.Count() {
		BadArg("illegal arg number:", i, "in:", QUO, w)
	}
	return w.arg(i)
}

// An internal method that WON'T make recursive calls as a result of logging
func (w WidgetArgs) arg(i int) string {
	return w.text[1+w.delim[i] : w.delim[i+1]]
}

func (w WidgetArgs) String() string {
	s := strings.Builder{}
	s.WriteString("[")
	for i := 0; i < w.Count(); i++ {
		s.WriteByte(':')
		if i == w.cursor {
			s.WriteByte('>')
		}
		s.WriteString(w.arg(i))
	}
	s.WriteString(":]")
	return s.String()
}

func (w WidgetArgs) Range(i int, j int) string {
	if i < 0 || j > w.Count() || i >= j {
		BadArg("bad range:", i, j, "in:", w)
	}
	return w.text[1+w.delim[i] : w.delim[j]]
}

func (w WidgetArgs) SetCursor(position int) {
	CheckArg(position >= 0 && position <= w.Count())
	w.cursor = position
}

func (w WidgetArgs) Peek() (bool, string) {
	if !w.Done() {
		return true, w.Arg(w.cursor)
	}
	return false, ""
}

func (w WidgetArgs) ReadIf(s string) bool {
	exists, value := w.Peek()
	if exists && value == s {
		w.cursor++
		return true
	}
	return false
}

func (w WidgetArgs) ReadIntWithinRange(minValue int, maxValue int) (bool, int) {
	exists, arg := w.Peek()
	if exists {
		value, err := ParseInt(arg)
		if err == nil && value >= minValue && value < maxValue {
			w.cursor++
			return true, value
		}
	}
	return false, -1
}
