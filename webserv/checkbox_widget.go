package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays a checkbox
type CheckboxWidgetObj struct {
	BaseWidgetObj
	Label      HtmlString
	switchFlag bool
}

type CheckboxWidget = *CheckboxWidgetObj

func NewCheckboxWidget(switchFlag bool, id string, label HtmlString) CheckboxWidget {
	w := CheckboxWidgetObj{}
	w.GetBaseWidget().Id = id
	w.Label = label
	w.switchFlag = switchFlag
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

	m.Comment("Checkbox").Cr()

	var cbClass string
	var role string
	if w.switchFlag {
		cbClass = `"form-check form-switch"`
		role = ` role="switch"`
	} else {
		cbClass = "form-check"
		role = ``
	}
	m.A(`<div class=`)
	m.A(cbClass)
	m.A(`>`).Cr()
	m.DoIndent()
	m.A(`<input class="form-check-input" type="checkbox" id="`).A(auxId).A(`"`)
	m.A(role)
	if WidgetBooleanValue(state, w.Id) {
		m.A(` checked`)
	}

	m.A(` onclick='jsCheckboxClicked("`).A(w.Id).A(`")'>`).Cr()
	m.Comment("Label").Cr()
	m.A(`<label class="form-check-label" for="`).A(auxId).A(`">`).Escape(w.Label).A(`</label>`).Cr()
	m.DoOutdent()
	m.A(`</div>`).Cr()

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
