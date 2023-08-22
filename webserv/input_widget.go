package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
	Label    HtmlString
	Password bool
}

type InputWidget = *InputWidgetObj

func NewInputWidget(id string, label HtmlString, password bool) InputWidget {
	w := InputWidgetObj{
		BaseWidgetObj: BaseWidgetObj{
			Id: id,
		},
		Label:    label,
		Password: password,
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

	m.A(`<div id="`, w.Id, `">`)

	m.DoIndent()

	problemId := WidgetIdWithProblem(w.Id)
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

	m.A(`" type="`, Ternary(w.Password, "password", "text"), `" id="`, w.Id, `.aux" value="`)
	value := WidgetStringValue(state, w.Id)
	m.Escape(value)
	m.A(`" onchange='jsVal("`, w.Id, `")'>`).Cr()

	if hasProblem {
		m.Comment("Problem")
		m.A(`<div class="form-text text-danger" style="font-size:  70%">`)
		m.Escape(problemText).A(`</div>`)
	}

	m.DoOutdent()

	m.A(`</div>`)
	m.Cr()
}
