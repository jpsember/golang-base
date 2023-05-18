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

func (q *Q) String() string { return q.message }

var q1 = Q{message: "q1"}
var q2 = Q{message: "q2"}
var q3 = Q{message: "q3"}

func TestArray(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var a = NewArray[Q]()
	a.Add(q1)
	a.Add(q2)
	a.Add(q3)
	CheckArg(a.Size() == 3)
}

func TestAddAndRemoveLots(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var a = NewArray[Q]()
	for i := 0; i < 100; i++ {
		a.Add(q1)
		a.Add(q2)
	}

	CheckArg(a.Size() == 200)
}
