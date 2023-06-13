package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"html"
)

var _ = Pr

const (
	AlertSuccess = iota
	AlertInfo
	AlertWarning
	AlertDanger
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
	w.Id = id
	return &w
}

var classNames = []string{`success`, `info`, `warning`, `danger`}

func (w AlertWidget) RenderTo(m MarkupBuilder, state JSMap) {
	pr := PrIf(false)
	desc := `AlertWidget ` + w.IdSummary()
	pr("rendering AlertWidget, desc:", desc, "class:", w.Class)
	m.A(`<div class='alert alert-`)
	m.A(classNames[w.Class])
	m.A(`' role='alert' id='`)
	m.A(w.Id)
	m.A(`'>`)
	alertMsg := state.OptString(w.Id, "No alert message found!")
	m.A(html.EscapeString(alertMsg))
	m.A(`</div>`)
	m.Cr()
}
