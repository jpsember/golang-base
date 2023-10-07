package webserv_test

import (
	"github.com/jpsember/golang-base/jt"
	. "github.com/jpsember/golang-base/webserv"
	"testing"
)

func TestValidatorValid(t *testing.T) {
	j := jt.New(t)
	validateHTML(j, `<a>"link"</a>`)
}

func TestMissingQuote(t *testing.T) {
	j := jt.New(t)
	validateHTML(j, `<a href="link without closing quote></a>`)
}

func validateHTML(j jt.JTest, content string) {
	js, err := SharedHTMLValidator().Validate(content)
	if err != nil {
		js.Put("", err.Error())
	}
	j.AssertMessage(js.String())
}
