package jimg_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

func TestReadJpg(t *testing.T) {
	j := jt.Newz(t)
	p := NewPathM("resources/balloons.jpg")
	bytes := p.ReadBytesM()
	i, err := jimg.DecodeImage(bytes)
	CheckOk(err)

	Pr("image:", INDENT, i.ToJson())
	j.AssertTrue(true)
}
