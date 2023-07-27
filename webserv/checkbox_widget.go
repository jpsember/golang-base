package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays a checkbox
type CheckboxWidgetObj struct {
	BaseWidgetObj
	Label HtmlString
}

type CheckboxWidget = *CheckboxWidgetObj

func NewCheckboxWidget(id string, label HtmlString) CheckboxWidget {
	w := CheckboxWidgetObj{
		BaseWidgetObj{
			Id: id,
		},
		label,
	}
	return &w
}

func (w CheckboxWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	auxId := w.Id + ".aux"

	//<div class="form-check">
	//  <input class="form-check-input" type="checkbox" value="" id="foo">
	//  <label class="form-check-label" for="foo">
	//    Checkbox 'foo'
	//  </label>
	//</div>

	m.A(`<div class "form-check" id="`)
	m.A(w.Id)
	m.A(`">`)

	m.DoIndent()
	m.DebugOpen(w)

	m.Comment("Checkbox").Cr()
	Pr("getting string value for:", Quoted(w.Id), "from:", INDENT, state)
	m.A(`<input class="form-check-input" type="checkbox" value="`).A(boolToHtmlString(WidgetBooleanValue(state, w.Id))).A(`" id="`).A(auxId).A(`"`)
	m.A(` onclick='jsCheckboxClicked("`).A(w.Id).A(`")'>`).Cr()
	m.Comment("Label").Cr()
	m.A(`<label class="form-check-label" for="`).A(auxId).A(`">`).Escape(w.Label).A(`</label>`).Cr()

	m.DebugClose()
	m.DoOutdent()

	m.A(`</div>`)
	m.Cr()
}

func boolToHtmlString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
