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

	SetTestAlertInfoState(true)
	defer SetTestAlertInfoState(false)

	s := "Sample panic message"
	TestPanicMessageLog.Reset()

	CheckArg(false, s)
	var str *string
	CheckNotNil(str, s)
	NotImplemented(s)
	NotSupported(s)
	Halt(s)
	ok := Error(s)
	CheckOk(ok, s)

	j.AssertMessage(TestPanicMessageLog.String())
}

// This test is very sensitive to line numbers; if this file changes, the hash might need
// updating.
func TestReportDelays(t *testing.T) {
	j := jt.New(t)

	SetTestAlertInfoState(true)
	defer SetTestAlertInfoState(false)

	//s := TestPanicSubstring
	TestPanicMessageLog.Reset()

	Alert("Normal message")
	Alert("-This shouldn't appear")

	TestAlertDuration = hour * 20
	Alert("!This should not appear, less than a day")
	TestAlertDuration = hour * 25
	Alert("!This should appear, more than a day")

	TestAlertDuration = day * 29
	Alert("?This should not appear, less than a month")
	TestAlertDuration = day * 32
	Alert("?This should appear, more than a month")

	for i := 0; i < 10; i++ {
		Alert("#4 This should appear four times, #4")
	}

	for i := 0; i < 10; i++ {
		Alert("#0 This shouldn't appear at all, #0")
	}

	f1("this is an alert without a skip")
	f1("<1 this is an alert with a skip of 1")

	j.AssertMessage(TestPanicMessageLog.String())
}

const minute = 60 * 1000
const hour = minute * 60
const day = hour * 24

func f1(key string) {
	Alert(key)
}
