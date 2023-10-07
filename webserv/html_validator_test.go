package webserv_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	. "github.com/jpsember/golang-base/webserv"
	"testing"
)

func TestValidatorValid(t *testing.T) {
	j := jt.New(t)
	validateHTML(j, true, `<a>"link"</a>`)
}

func TestMissingQuote(t *testing.T) {
	j := jt.New(t)
	validateHTML(j, false, `<a href="link without closing quote></a>`)
}

func validateHTML(j jt.JTest, expected bool, content string) error {
	js, err := SharedHTMLValidator().ValidateWithoutCache(content)
	j.Log("result:", INDENT, js)
	j.AssertEqual(expected, err == nil)
	return err
}
