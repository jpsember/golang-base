package webserv

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
}

type InputWidget = *InputWidgetObj

func NewInputWidget(id string, size int) InputWidget {
	w := InputWidgetObj{
		BaseWidgetObj{
			Id: id,
		},
	}
	return &w
}

func (w InputWidget) RenderTo(m MarkupBuilder) {
	desc := `InputWidget ` + w.IdSummary()
	m.A(`<input type="text" value="`)
	m.A(desc)
	m.A(`" onchange="onChange('`)
	m.A(w.Id)
	m.A(`')">`)
	m.Cr()
	//m.OpenHtml(`input type="text" value="`+desc+`"`, desc)
}
