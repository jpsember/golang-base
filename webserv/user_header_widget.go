package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

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
	user := app.UserForSession(s)
	signedIn := user.Id() != 0

	Todo("Include debug page name optionally")

	// These are the widgets we'd like, something like this:
	_ = `               
                 <div id=".3" class="text-end">
                  <span style="font-size:0.6em">
                    Welcome, manager1
                  </span>

                  <button class="m-2 btn btn-outline-primary btn-sm" style="font-size:0.6em">Sign Out</button>

                </div>
`

	fontSizeExpr := ` style="font-size:0.6em"`
	m.OpenTag(`div id=`, w.BaseId, `'`)
	m.DoIndent()
	{
		m.OpenTag(`div class="text-end"`)
		{
			if DebugUIFlag {
				pg := s.DebugPage
				nm := `??pagename??`
				if pg != nil {
					nm = pg.Name()
				}
				m.OpenTag(`span class="text-success"`, fontSizeExpr)
				m.A(`Page:`, nm)
				m.CloseTag()
				m.OpenTag(`span class="m-2"`)
				m.CloseTag()
			}

			if signedIn {
				m.OpenTag(`span`, fontSizeExpr)
				m.A(`Welcome, `).Escape(user.Name())
				m.CloseTag()
			}

			Todo("Add button listener")
			m.OpenTag(`button class="m-2 btn btn-outline-primary btn-sm"`, fontSizeExpr)
			if signedIn {
				m.A(`Sign Out`)
			} else {
				m.A(`Sign In`)
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.DoOutdent()
	m.CloseTag()
}
