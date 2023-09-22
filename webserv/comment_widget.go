package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A widget that serves only to insert comments within a page.

type CommentWidgetObj struct {
	BaseWidgetObj
	markup string
}

type CommentWidget = *CommentWidgetObj

func NewCommentWidget(id string, args ...any) CommentWidget {
	t := &CommentWidgetObj{}
	t.InitBase(id)

	content := ToString(args...)
	t.markup = "<!-- " + NewHtmlString(content).Escaped() + " -->"
	return t
}

func (w CommentWidget) RenderTo(s Session, m MarkupBuilder) {
	m.Cr()
	m.handlePendingCr()
	m.WriteString(w.markup)
	m.Cr()
}
