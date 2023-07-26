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

	m.A(`<div class="container" id='`)
	m.A(w.Id)
	m.A(`'>`)

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
		m.A(`<div class="row-text-sm" style="font-size: 70%">`).Cr()
		m.H(labelHtml)
		m.A("</div>").Cr()
	}

	m.HtmlComment("Input")
	m.A(`<div class="row">`).Cr()
	value := WidgetStringValue(state, w.Id)
	m.A(`<input `)
	if hasProblem {
		m.A(`class="text-bg-danger" `)
	}
	m.A(`type='text' id='`)
	m.A(w.Id)
	m.A(`.aux`)
	m.A(`' value='`)
	m.H(NewHtmlString(value))
	m.A(`' onchange='jsVal("`)
	m.A(w.Id)
	m.A(`")'>`)
	m.Cr()
	m.A("</div>").Cr()
	Todo("Simplify m.HtmlComment, m.H, m.A(NewHtmlString)...")

	m.HtmlComment("Problem")
	m.A(`<div class="row text-danger" style="font-size:  60%" class="text-danger">`).Cr()
	problemHtml := HtmlStringNbsp
	if hasProblem {
		problemHtml = NewHtmlString(problemText)
	}
	m.H(problemHtml).A(`</div>`).Cr()

	m.DebugClose()
	m.DoOutdent()

	m.A(`</container>`)
	m.Cr()
}
