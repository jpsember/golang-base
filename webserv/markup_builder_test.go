package webserv_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"github.com/jpsember/golang-base/webserv"
	"testing"
)

var _ = Pr

func TestVoidTag(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	m := webserv.NewMarkupBuilder()
	m.TgOpen("div")
	m.TgClose()

	j.AssertMessage(m.String())
}

func TestTagWithAttributesAndContent(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()
	m.TgOpen("div")
	m.A(` class="foo"`)
	m.TgContent()
	m.A("Hello")
	m.TgClose()

	j.AssertMessage(m.String())
}

func TestTagWithAttributeInOpen(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()
	m.TgOpen(`div class="foo"`)
	m.TgContent()
	m.A("Hello")
	m.TgClose()

	j.AssertMessage(m.String())
}

func TestTagWithStyling(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()
	m.TgOpen(`div class="foo"`)
	m.Style(`width=4em;`)
	m.TgContent()
	m.A("Hello")
	m.TgClose()

	j.AssertMessage(m.String())
}

func TestTagWithoutContent1(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()

	m.TgOpen(`div`).TgClose()

	j.AssertMessage(m.String())
}

func TestTagWithoutContent2(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()

	m.TgOpen(`div class="foo"`).TgClose()

	j.AssertMessage(m.String())
}

func TestTag3(t *testing.T) {
	j := jt.New(t)

	m := webserv.NewMarkupBuilder()

	m.Comment("checkbox").TgOpen(`div class=`).A(QUOTED, `abc`).TgContent()
	{
		m.TgOpen(`input class="form-check-input" type="checkbox" id=`).A(QUOTED, "auxid").TgClose()
		{
			m.Comment("Label").TgOpen(`label class="form-check-label" for=`)
			m.A(QUOTED, "auxid").TgContent().A(ESCAPED, "fox & hound").TgClose()
		}
	}
	m.TgClose()

	j.AssertMessage(m.String())
}
