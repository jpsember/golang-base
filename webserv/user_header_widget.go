package webserv

type UserHeaderWidgetStruct struct {
	BaseWidgetObj
}

type UserHeaderWidget = *UserHeaderWidgetStruct

func NewUserHeaderWidget(id string) UserHeaderWidget {
	t := &UserHeaderWidgetStruct{}
	t.BaseId = id
	return t
}

func (w UserHeaderWidget) RenderTo(s Session, m MarkupBuilder) {
	app := SessionApp(s)

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
