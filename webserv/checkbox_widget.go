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

	m.Comment("CheckboxWidget")
	m.OpenTag(`div id="`, w.Id, `"`)
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
				Ternary(WidgetBooleanValue(state, w.Id), ` checked`, ``),
				` onclick='jsCheckboxClicked("`, w.Id, `")'`)

			{
				m.Comment("Label").OpenTag(`label class="form-check-label" for="`, auxId, `"`).Escape(w.Label).CloseTag() //.A(`</label>`).Cr()
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func Ternary[V any](flag bool, ifTrue V, ifFalse V) V {
	if flag {
		return ifTrue
	}
	return ifFalse
}

func boolToHtmlString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
