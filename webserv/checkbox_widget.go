package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strconv"
)

type CheckboxWidgetListener func(sess Session, widget CheckboxWidget, state bool) (bool, error)

// A Widget that displays a checkbox
type CheckboxWidgetObj struct {
	BaseWidgetObj
	Label      HtmlString
	listener   CheckboxWidgetListener
	switchFlag bool
}

type CheckboxWidget = *CheckboxWidgetObj

func doNothingCheckboxListener(sess Session, widget CheckboxWidget, state bool) (bool, error) {
	Pr("'do nothing' doNothingCheckboxListener called")
	return state, nil
}

func NewCheckboxWidget(switchFlag bool, id string, label HtmlString, listener CheckboxWidgetListener) CheckboxWidget {
	if listener == nil {
		listener = doNothingCheckboxListener
	}
	w := CheckboxWidgetObj{
		Label:      label,
		switchFlag: switchFlag,
		listener:   listener,
	}
	w.InitBase(id)
	w.LowListen = checkboxListenWrapper
	return &w
}

func checkboxListenWrapper(sess Session, widget Widget, value string) (any, error) {
	highLevelListener := widget.(CheckboxWidget)
	boolValue := false
	if b, err := strconv.ParseBool(value); err != nil {
		Alert("trouble parsing bool from:", Quoted(value))
	} else {
		boolValue = b
	}
	return highLevelListener.listener(sess, highLevelListener, boolValue)
}

func (w CheckboxWidget) RenderTo(s Session, m MarkupBuilder) {
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
				Ternary(s.WidgetBoolValue(w), ` checked`, ``),
				` onclick='jsCheckboxClicked("`, s.baseIdPrefix+w.BaseId, `")'`)

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
