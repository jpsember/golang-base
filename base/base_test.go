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

	s := "lorem ipsum"
	TestAbortMessageLog.Reset()

	CheckArg(false, s)
	NotImplemented(s)
	NotSupported(s)
	Halt(s)
	Die(s)
	CheckNonEmpty("", s)
	ok := Error("Sed", "ut", "perspiciatis", "unde", "omnis")
	CheckOkWith("sample error", ok, s)
	nestedAssertions("<1 Nested assertions")
	nestedAssertions2()
	expression := NewPathM("alpha/bravo")
	CheckNil(expression, s)

	j.AssertMessage(TestAbortMessageLog.String())
}

// This test is very sensitive to line numbers; if this file changes, the hash might need
// updating.
func TestAlerts(t *testing.T) {
	j := jt.New(t)

	SetTestAlertInfoState(true)
	defer SetTestAlertInfoState(false)

	TestAbortMessageLog.Reset()

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

	j.AssertMessage(TestAbortMessageLog.String())
}

const minute = 60 * 1000
const hour = minute * 60
const day = hour * 24

func f1(key string) {
	Alert(key)
}

func nestedAssertions(s string) {
	CheckArg(false, s)
	NotImplemented(s)
	NotSupported(s)
	Halt(s)
	ok := Error("This", "is", "an", "error", "message")
	CheckOkWith("sample result", ok, s)
}

func nestedAssertions2() {
	const s = "<2 Nested assertions"
	nestedAssertions(s)
}
