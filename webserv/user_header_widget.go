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
	Todo("!Use new embedded widgets technique")
	app := SessionApp(s)
	user := app.UserForSession(s)
	signedIn := user.Id() != 0

	m.TgOpen(`div id=`).A(QUOTED, w.BaseId).TgContent()
	//m.OpenTag(`div id="`, w.BaseId, `"`)

	// Adding a background image; I read this post: https://mdbootstrap.com/docs/standard/navigation/headers/
	img := w.BgndImageMarkup
	if img != "" {
		m.TgOpen(`div class="bg-image" `).A(img).TgContent()
		//m.OpenTag(`div class="bg-image" `, img)
	}

	{
		m.TgOpen(`div class="text-end"`).TgContent()
		{
			if DebugUIFlag {
				pg := s.DebugPage
				nm := `??pagename??`
				if pg != nil {
					nm = pg.Name()
				}
				m.TgOpen(`span class="text-success"`).Style(`font-size:0.6em`).TgContent()
				m.A(`Page:`, nm)
				m.TgClose()
				m.TgOpen(`span class="m-2"`).TgContent().TgClose()
			}

			if signedIn {
				m.TgOpen(`span`).Style(`font-size:0.6em`).TgContent()
				m.A(`Welcome, `).Escape(user.Name())
				m.TgClose()
			}

			actionId := Ternary(signedIn, BUTTON_ID_SIGN_OUT, BUTTON_ID_SIGN_IN)

			m.TgOpen(`button class="m-2 btn btn-outline-primary btn-sm"`).A(
				` onclick="jsButton('`, s.baseIdPrefix+actionId, `')"`).Style(`font-size:0.6em`).TgContent()

			if signedIn {
				m.A(`Sign Out`)
			} else {
				m.A(`Sign In`)
			}
			m.TgClose()
		}
		m.TgClose()
	}
	if img != "" {
		m.TgClose()
		//m.CloseTag()
	}
	m.TgClose()
}
