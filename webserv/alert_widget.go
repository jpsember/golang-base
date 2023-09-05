package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

var _ = Pr

const (
	AlertSuccess = iota
	AlertInfo
	AlertWarning
	AlertDanger
	AlertTotal
)

type AlertClass int

type AlertWidgetObj struct {
	BaseWidgetObj
	Class AlertClass
}

type AlertWidget = *AlertWidgetObj

func NewAlertWidget(id string, alertClass AlertClass) AlertWidget {

	w := AlertWidgetObj{
		Class: alertClass,
	}
	w.BaseId = id
	return &w
}

var classNames = []string{`success`, `info`, `warning`, `danger`}

func (w AlertWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	pr := PrIf(false)
	desc := `AlertWidget ` + w.IdSummary()
	pr("rendering AlertWidget, desc:", desc, "class:", w.Class)
	alertMsg := state.OptString(w.BaseId, "No alert message found!")
	m.A(`<div class='alert alert-`, classNames[w.Class],
		`' role='alert' id='`, w.BaseId, `'>`, html.EscapeString(alertMsg), `</div>`).Cr()
}
