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

	Todo("can we do OpenTag here?")
	m.A(`<div id="`, w.Id, `">`)
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
	m.A(`<div class=`, cbClass, `>`).Cr()
	m.DoIndent()
	m.A(`<input class="form-check-input" type="checkbox" id="`, auxId, `"`, role)
	if WidgetBooleanValue(state, w.Id) {
		m.A(` checked`)
	}

	m.A(` onclick='jsCheckboxClicked("`, w.Id, `")'>`).Cr()
	m.Comment("Label").Cr()
	m.A(`<label class="form-check-label" for="`, auxId, `">`).Escape(w.Label).A(`</label>`).Cr()
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
