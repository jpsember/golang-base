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
	auxId := w.AuxId()

	m.A(`<div id="`)
	m.A(w.Id)
	m.A(`">`)

	m.DoIndent()
	m.DebugOpen(w)

	m.Comment("Checkbox").Cr()
	m.A(`<div class="form-check">`).Cr()
	m.DoIndent()
	m.A(`<input class="form-check-input" type="checkbox" value="" id="`).A(auxId).A(`"`)
	if WidgetBooleanValue(state, w.Id) {
		m.A(` checked`)
	}

	m.A(` onclick='jsCheckboxClicked("`).A(w.Id).A(`")'>`).Cr()
	m.Comment("Label").Cr()
	m.A(`<label class="form-check-label" for="`).A(auxId).A(`">`).Escape(w.Label).A(`</label>`).Cr()
	m.DoOutdent()
	m.A(`</div>`).Cr()

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
