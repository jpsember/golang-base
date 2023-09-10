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
type InputWidgetListener func(sess Session, widget Widget, value string) (string, error)

func NewInputWidget(id string, label HtmlString, listener InputWidgetListener, password bool) InputWidget {
	if listener == nil {
		listener = dummyInputWidgetListener
	}
	w := InputWidgetObj{
		BaseWidgetObj: BaseWidgetObj{
			BaseId: id,
		},
		Label:    label,
		Password: password,
		listener: listener,
	}
	w.Base().LowListen = inputListenWrapper
	return &w
}

func inputListenWrapper(sess Session, widget Widget, value string) (string, error) {
	inp := widget.(InputWidget)
	value = strings.TrimSpace(value)
	result, err := inp.listener(sess, inp, value)
	return result, err
}

var HtmlStringNbsp = NewHtmlStringEscaped("&nbsp;")

func dummyInputWidgetListener(sess Session, widget Widget, value string) (string, error) {
	Alert("#50No InputWidgetListener implemented for id:", widget.Id())
	return "garbage", DummyError
}

func (w InputWidget) RenderTo(m MarkupBuilder, state JSMap) {

	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	// While <input> are span tags, our widget should be considered a block element

	// The outermost element must have id "foo", since we will be replacing that id's outerhtml
	// to perform AJAX updates.
	//
	// The HTML input element has id "foo.aux"
	// If there is a problem with the input, its text will have id "foo.problem"

	m.A(`<div id="`, w.BaseId, `">`)

	m.DoIndent()

	problemId := WidgetIdWithProblem(w.BaseId)
	problemText := state.OptString(problemId, "")
	if false && Alert("always problem") {
		problemText = "sample problem information"
	}
	hasProblem := problemText != ""

	labelHtml := w.Label
	if labelHtml != nil {
		m.Comment("Label")
		m.OpenTag(`label class="form-label" style="font-size:70%"`)
		m.Escape(labelHtml)
		m.CloseTag()
	}

	m.Comment("Input")
	m.A(`<input class="form-control`)
	if hasProblem {
		m.A(` border-danger border-3`) // Adding border-3 makes the text shift a bit on error, maybe not desirable
	}

	m.A(`" type="`, Ternary(w.Password, "password", "text"), `" id="`, w.BaseId, `.aux" value="`)
	value := WidgetStringValue(state, w.BaseId)
	m.Escape(value)
	m.A(`" onchange='jsVal("`, w.BaseId, `")'>`).Cr()

	if hasProblem {
		m.Comment("Problem")
		m.A(`<div class="form-text text-danger" style="font-size:  70%">`)
		m.Escape(problemText).A(`</div>`)
	}

	m.DoOutdent()

	m.A(`</div>`)
	m.Cr()
}
