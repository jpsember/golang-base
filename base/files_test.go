package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

var _ = Pr

func TestPaths(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	//j.SetVerbose()

	var result = NewJSMap()
	var samples = []string{".", "", "..", "/", "alpha/bravo", //
		"alpha bravo", "alpha\bravo", "alpha/bravo/", //
		"a/b/..//c", "./alpha", "../alpha", ".../s"}

	for _, n := range samples {
		p, err := NewPath(n)
		if err != nil {
			result.Put(n, err.Error())
		} else {
			result.Put(n, p.String())
		}
	}

	j.AssertMessage(result)

}

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
