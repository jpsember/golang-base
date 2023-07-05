package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

func TestPaths(t *testing.T) {
	j := jt.New(t)

	j.SetVerbose()

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

func TestEmptyAndRoot(t *testing.T) {
	j := jt.New(t)
	r := NewPathM("/")
	p2 := NewPathM("alpha")
	j.AssertTrue(r.IsRoot())
	j.AssertFalse(p2.IsRoot())
}

func TestJoinRoot(t *testing.T) {
	j := jt.New(t)
	p := NewPathM("/")
	p3 := NewPathM("foo")
	join(p, p3)
	join(NewPathM("alpha"), p3)
	j.AssertMessage(sb.String())
}

func TestCurrentDir(t *testing.T) {
	j := jt.New(t)
	p := NewPathM(".")
	j.AssertMessage(p)
}

func TestCurrentDir2(t *testing.T) {
	j := jt.New(t)
	p := NewPathM(".")
	j.AssertMessage(p.JoinM("alpha/charlie"))
}

func TestAscendAbs(t *testing.T) {
	j := jt.New(t)
	p := NewPathM("/alpha/bravo/charlie")
	for {
		if p.Empty() {
			break
		}
		sb.Pr(p.String()).Cr()
		p = p.Parent()
	}
	j.AssertMessage(sb.String())
}

func TestAscendRelative(t *testing.T) {
	j := jt.New(t)
	p := NewPathM("alpha/bravo/charlie")
	for {
		if p.Empty() {
			break
		}
		sb.Pr(p.String()).Cr()
		p = p.Parent()
	}
	j.AssertMessage(sb.String())
}

func join(a Path, b Path) {
	sb.Pr("join", Quoted(a.String()), "to", Quoted(b.String()), "is:", Quoted(a.JoinPathM(b).String())).Cr()
}

var sb = BasePrinter{}
