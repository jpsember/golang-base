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
	m.HtmlComment("Have onChange send the id of the widget with the text back to the server")
	m.A(`<input type="text" id=`).Quoted(w.Id).A(` value=`).Quoted(desc).A(` onchange=`).Quoted(`jsVal('` + w.Id + `')`).A(`>`).Cr()
}
