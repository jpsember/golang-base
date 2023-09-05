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
	w.Base().BaseId = id
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

	m.Comment("CheckboxWidget")
	m.OpenTag(`div id="`, w.BaseId, `"`)
	{
		var cbClass string
		var role string
		if w.switchFlag {
			cbClass = `"form-check form-switch"`
			role = ` role="switch"`
		} else {
			cbClass = "form-check"
			role = ``
		}

		m.Comment("checkbox").OpenTag(`div class=`, cbClass)
		{
			m.VoidTag(
				`input class="form-check-input" type="checkbox" id="`, auxId, `"`, role,
				Ternary(WidgetBooleanValue(state, w.BaseId), ` checked`, ``),
				` onclick='jsCheckboxClicked("`, w.BaseId, `")'`)

			{
				m.Comment("Label").OpenTag(`label class="form-check-label" for="`, auxId, `"`).Escape(w.Label).CloseTag() //.A(`</label>`).Cr()
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func boolToHtmlString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
