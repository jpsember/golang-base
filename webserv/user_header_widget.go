package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type UserHeaderWidgetStruct struct {
	BaseWidgetObj
	BgndImageMarkup string
	listener        ButtonWidgetListener
}

type UserHeaderWidget = *UserHeaderWidgetStruct

func NewUserHeaderWidget(id string, listener ButtonWidgetListener) UserHeaderWidget {
	t := &UserHeaderWidgetStruct{}
	t.InitBase(id)
	t.listener = listener
	t.LowListen = t.buttonListenWrapper
	return t
}

func (w UserHeaderWidget) buttonListenWrapper(sess Session, widget Widget, value string) (any, error) {
	w.listener(sess, widget, value)
	return nil, nil
}

const (
	USER_HEADER_ACTION_SIGN_OUT = "sign_out"
	USER_HEADER_ACTION_SIGN_IN  = "sign_in"
)

func (w UserHeaderWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf("UserHeaderWidget", true)
	pr("RenderTo, widget id", w.Id(), "BaseId:", w.Id())
	Todo("!Use new embedded widgets technique")
	app := SessionApp(s)
	user := app.UserForSession(s)
	signedIn := user.Id() != 0

	m.TgOpen(`div id=`).A(QUOTED, w.Id()).TgContent()

	// Adding a background image; I read this post: https://mdbootstrap.com/docs/standard/navigation/headers/
	img := w.BgndImageMarkup
	if img != "" {
		m.TgOpen(`div class="bg-image" `).A(img).TgContent()
	}

	{
		m.TgOpen(`div class="text-end"`).TgContent()
		{
			if DebugUIFlag {
				pg := s.debugPage
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

			actionId := Ternary(signedIn, USER_HEADER_ACTION_SIGN_OUT, USER_HEADER_ACTION_SIGN_IN)

			m.TgOpen(`button class="m-2 btn btn-outline-primary btn-sm"`).A(
				` onclick="jsButton('`, w.Id(), `.`, actionId, `')"`).Style(`font-size:0.6em`).TgContent()

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
