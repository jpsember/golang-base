package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

var DummyError = Error("Example error message")

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
	Label    HtmlString
	Password bool
	listener InputWidgetListener
}

type InputWidget = *InputWidgetObj
type InputWidgetListener func(sess Session, widget InputWidget, value string) (string, error)

func NewInputWidget(id string, label HtmlString, listener InputWidgetListener, password bool) InputWidget {
	Todo("?Add multi-line input fields, different font sizes")
	if listener == nil {
		listener = dummyInputWidgetListener
	}
	w := InputWidgetObj{
		Label:    label,
		Password: password,
		listener: listener,
	}
	w.InitBase(id)
	w.LowListen = inputListenWrapper
	return &w
}

func inputListenWrapper(sess Session, widget Widget, value string, args []string) (any, error) {
	pr := PrIf("inputListenWrapper", true)
	inp := widget.(InputWidget)
	value = strings.TrimSpace(value)
	pr("widget id:", widget.Id(), "value:", QUO, value)
	result, err := inp.listener(sess, inp, value)
	pr("after calling client listener, result:", QUO, result, "err:", err)
	pr("callers:", Callers(0, 5))
	return result, err
}

var HtmlStringNbsp = NewHtmlStringEscaped("&nbsp;")

func dummyInputWidgetListener(sess Session, widget InputWidget, value string) (string, error) {
	Alert("#50No InputWidgetListener implemented for id:", widget.Id())
	return "garbage", DummyError
}

func (w InputWidget) RenderTo(s Session, m MarkupBuilder) {

	debug := false
	pr := PrIf("input_widget.RenderTo", debug)
	if debug {
		pr(VERT_SP, "widget id:", w.Id())
		pr("stacked state:", INDENT, s.StateStackToJson())
	}

	// While <input> are span tags, our widget should be considered a block element

	// The outermost element must have id "foo", since we will be replacing that id's outerhtml
	// to perform AJAX updates.
	//
	// The HTML input element has id "foo.aux"
	// If there is a problem with the input, its text will have id "foo.problem"

	id := s.PrependId(w.Id())

	m.TgOpen(`div id=`).A(QUO, id).TgContent()

	auxId := id + `.aux`
	problemText := s.WidgetProblem(w)
	if false && Alert("always problem") {
		problemText = "sample problem information"
	}
	hasProblem := problemText != ""

	labelHtml := w.Label
	if labelHtml != nil {
		m.Comment("Label")
		m.TgOpen(`label class="form-label"`).Style(`font-size:70%`).TgContent()
		m.Escape(labelHtml)
		m.TgClose()
	}

	m.Comment("Input")
	m.TgOpen(`input class="form-control`)
	if hasProblem {
		m.A(` border-danger border-3`) // Adding border-3 makes the text shift a bit on error, maybe not desirable
	}

	m.A(`" type=`, QUO, Ternary(w.Password, "password", "text"), ` id="`, auxId, `" value="`)
	m.A(ESCAPED, s.WidgetStringValue(w))

	m.A(`" onchange="jsVal('`, id, `')"`)
	// Set this id as the element that has the focus.  Note we are using auxId, for the actual <input> element, not its containing widget <div>
	m.A(` onfocus="jsFocus('`, auxId, `')"`)
	m.TgClose()

	if hasProblem {
		m.Comment("Problem")
		m.TgOpen(`div class="form-text text-danger"`).Style(`font-size:70%;`).TgContent().A(ESCAPED, problemText).TgClose()
	}

	m.TgClose()
}

func (w InputWidget) ValidationValue(s Session) (string, bool) {
	return s.WidgetStringValue(w), true
}

func (w InputWidget) ValueAsString(s Session) string {
	return s.WidgetStringValue(w)
}
