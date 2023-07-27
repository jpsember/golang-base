package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
	Label HtmlString
}

type InputWidget = *InputWidgetObj

func NewInputWidget(id string, label HtmlString) InputWidget {
	w := InputWidgetObj{
		BaseWidgetObj{
			Id: id,
		},
		label,
	}
	return &w
}

var HtmlStringNbsp = NewHtmlStringEscaped("&nbsp;")

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

	m.A(`<div id="`)
	m.A(w.Id)
	m.A(`">`)

	m.DoIndent()
	m.DebugOpen(w)

	problemId := w.Id + ".problem"
	problemText := state.OptString(problemId, "")
	if false && Alert("always problem") {
		problemText = "sample problem information"
	}
	hasProblem := problemText != ""

	labelHtml := w.Label
	if labelHtml != nil {
		m.HtmlComment("Label")
		m.A(`<label class="form-label" style="font-size:70%">`).Cr()
		m.H(labelHtml)
		m.A(`</label>`).Cr()
	}

	m.HtmlComment("Input")
	m.A(`<input class="form-control`)
	if hasProblem {
		m.A(` border-danger border-3`) // Adding border-3 makes the text shift a bit on error, maybe not desirable
	}
	m.A(`" type="text" id="`)
	m.A(w.Id)
	m.A(`.aux" value="`)
	value := WidgetStringValue(state, w.Id)
	m.H(NewHtmlString(value))
	m.A(`" onchange='jsVal("`)
	m.A(w.Id)
	m.A(`")'>`).Cr()

	Todo("Simplify m.HtmlComment, m.H, m.A(NewHtmlString)...")

	if hasProblem {
		m.HtmlComment("Problem")
		m.A(`<div class="form-text`)
		m.A(` text-danger" style="font-size:  70%">`)
		problemHtml := NewHtmlString(problemText)
		Todo("Have MarkupBuilder method to write escaped form of argument")
		m.H(problemHtml).A(`</div>`).Cr()
	}

	m.DebugClose()
	m.DoOutdent()

	m.A(`</div>`)
	m.Cr()
}
