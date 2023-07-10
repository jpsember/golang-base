package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

type Q struct {
	message string
}

func q(msg ...any) *Q {
	var x = Q{message: ToString(msg)}
	return &x
}

func (q *Q) String() string { return q.message }

func TestArray(t *testing.T) {
	j := jt.New(t)
	j.SetVerbose()

	var a = NewArray[*Q]()
	a.Add(q("q1"))
	a.Add(q("q2"))
	a.Add(q("q3"))
	CheckArg(a.Size() == 3)
}

func TestAddAndRemoveLots(t *testing.T) {
	j := jt.New(t)
	j.Nothing()

	var a = NewArray[*Q]()
	for i := 0; i < 100; i++ {
		a.Add(q("q #", i))
		CheckState(a.Size() == i+1)
	}

	CheckArg(a.Size() == 100)
}

func TestSort(t *testing.T) {
	j := jt.New(t)
	var a = NewArray[string]()
	a.Add("milk")
	a.Add("eggs")
	a.Add("raisins")
	a.Add("flour")
	a.Sort()

	j.AssertMessage(a)
}
