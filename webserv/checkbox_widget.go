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
	w.SetLowListener(checkboxListenWrapper)
	return &w
}

func checkboxListenWrapper(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
	pr := PrIf("checkboxListenWrapper", false)
	highLevelListener := widget.(CheckboxWidget)
	boolValue := false
	pr("widget id:", widget.Id(), "value:", Quoted(value))
	if b, err := strconv.ParseBool(value); err != nil {
		Alert("trouble parsing bool from:", Quoted(value))
	} else {
		boolValue = b
	}
	return highLevelListener.listener(sess, highLevelListener, boolValue)
}

func (w CheckboxWidget) RenderTo(s Session, m MarkupBuilder) {
	auxId := s.PrependId(w.AuxId())

	id := s.PrependId(w.Id())
	m.TgOpen(`div id="`).A(id, `"`).TgContent()
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

		m.Comment("checkbox").TgOpen(`div class=`).A(QUO, cbClass).TgContent()
		{
			m.TgOpen(`input class="form-check-input" type="checkbox" id='`).A(auxId, `'`, role,
				Ternary(s.WidgetBoolValue(w), ` checked`, ``),
				` onclick="jsCheckboxClicked('`, s.PrependId(w.baseId), `')"`).TgClose()
			{
				m.Comment("Label").TgOpen(`label class="form-check-label" for=`).A(QUO, auxId).TgContent().Escape(w.Label).TgClose()
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

func (w CheckboxWidget) ValueAsString(s Session) string {
	return Ternary(s.WidgetBoolValue(w), `true`, `false`)
}

func (w CheckboxWidget) ValidationValue(s Session) (string, bool) {
	return Ternary(s.WidgetBoolValue(w), `true`, `false`), true
}
