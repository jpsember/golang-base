package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

var _ = Pr

func TestJoin(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	//j.SetVerbose()

	var result = NewJSMap()
	var samples = []string{
		"a", "b", //
		"a/b", "c/z", //
	}
	for i := 0; i < len(samples); i += 2 {
		p := samples[i]
		parent, _ := NewPath(p)
		child := samples[i+1]
		r, _ := parent.Join(child)
		result.Put(p+" + "+child, string(r))
	}

	j.AssertMessage(result)

}
