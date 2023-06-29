package base_test

import (
	"github.com/jpsember/golang-base/jt"
	"testing"
)

import (
	. "github.com/jpsember/golang-base/base"
)

// Tests the proper reporting of error locations (i.e., all the 'skipCount' expressions).
// This test is very sensitive to line numbers; if this file changes, the hash might need
// updating.
func TestPanics(t *testing.T) {
	j := jt.New(t)

	s := TestPanicSubstring
	TestPanicMessageLog.Reset()

	CheckArg(false, s)
	CheckNotNil(nil, s)
	NotImplemented(s)
	NotSupported(s)
	Halt(s)
	ok := Error(s)
	CheckOk(ok, s)

	j.AssertMessage(TestPanicMessageLog.String())
}
