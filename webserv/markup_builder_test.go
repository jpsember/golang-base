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
	m.StyleOn().A(`width=4em;`)
	m.StyleOff()
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
