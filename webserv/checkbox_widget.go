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

	m.TgOpen(`div id="`).A(w.BaseId, `"`).TgContent()
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

		m.Comment("checkbox").TgOpen(`div class=`).Quote().A(cbClass).TgContent()
		{
			m.TgOpen(`input class="form-check-input" type="checkbox" id=`).A(QUOTED, auxId, role,
				Ternary(s.WidgetBoolValue(w), ` checked`, ``),
				` onclick='jsCheckboxClicked("`, s.baseIdPrefix, w.BaseId, `")'`).TgClose()
			{
				m.Comment("Label").TgOpen(`label class="form-check-label" for=`).A(QUOTED, auxId).TgContent().Escape(w.Label).TgClose()
			}
		}
		m.TgClose()
	}
	m.TgClose()
}

func boolToHtmlString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
