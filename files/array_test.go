package files_test

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

var _ = Pr

type Q struct {
	message string
}

func q(msg ...any) *Q {
	var x = Q{message: ToString(msg)}
	return &x
}

func (q *Q) String() string { return q.message }

func TestArray(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var a = NewArray[*Q]()
	a.Add(q("q1"))
	a.Add(q("q2"))
	a.Add(q("q3"))
	CheckArg(a.Size() == 3)
}

func TestAddAndRemoveLots(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var a = NewArray[*Q]()
	for i := 0; i < 100; i++ {
		a.Add(q("q #", i))
		CheckState(a.Size() == i+1)
	}

	CheckArg(a.Size() == 100)
}
