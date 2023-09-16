package webserv

type HeaderWidgetObj struct {
	BaseWidgetObj
}

type HeaderWidget = *HeaderWidgetObj

func NewHeaderWidget(id string) HeaderWidget {
	t := &HeaderWidgetObj{}
	t.BaseId = id
	return t
}

func (w HeaderWidget) RenderTo(s Session, m MarkupBuilder) {

	var app ServerApp
	app = s.app.(ServerApp)

	m.A(`<div id='`, w.BaseId, `'>`)
	m.DoIndent()
	{
		user := app.UserForSession(s)
		if user.Id() != 0 {
			m.A("Welcome, ", user.Name())
		} else {
			m.A("Please sign in")
		}
	}
	m.DoOutdent()
	m.A(`</div>`)
}
