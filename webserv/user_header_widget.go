package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type UserHeaderWidgetStruct struct {
	BaseWidgetObj
	BgndImageMarkup string
}

type UserHeaderWidget = *UserHeaderWidgetStruct

func NewUserHeaderWidget(id string) UserHeaderWidget {
	t := &UserHeaderWidgetStruct{}
	t.InitBase(id)
	return t
}

const (
	HEADER_WIDGET_BUTTON_PREFIX = "uhdr."
	BUTTON_ID_SIGN_OUT          = HEADER_WIDGET_BUTTON_PREFIX + "sign_out"
	BUTTON_ID_SIGN_IN           = HEADER_WIDGET_BUTTON_PREFIX + "sign_in"
)

func (w UserHeaderWidget) RenderTo(s Session, m MarkupBuilder) {
	app := SessionApp(s)
	user := app.UserForSession(s)
	signedIn := user.Id() != 0

	fontSizeExpr := ` style="font-size:0.6em"`
	m.OpenTag(`div id=`, w.BaseId, `'`)

	// Adding a background image; I read this post: https://mdbootstrap.com/docs/standard/navigation/headers/
	img := w.BgndImageMarkup
	if img != "" {
		m.OpenTag(`div class="bg-image" `, img)
	}

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

			actionId := Ternary(signedIn, BUTTON_ID_SIGN_OUT, BUTTON_ID_SIGN_IN)

			m.OpenTag(`button class="m-2 btn btn-outline-primary btn-sm"`, fontSizeExpr,
				` onclick='jsButton("`, actionId, `")'`)

			if signedIn {
				m.A(`Sign Out`)
			} else {
				m.A(`Sign In`)
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	if img != "" {
		m.CloseTag()
	}
	m.CloseTag()
}
