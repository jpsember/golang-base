package data

import (
	"testing"

	. "github.com/jpsember/golang-base/base"
)

func TestEnumInfo(t *testing.T) {

	var info = NewEnumInfo("starting active dead")

	Pr("hello")
	t.Log(info.EnumIds)
	// if got != want {
	// 	t.Errorf("got %q, wanted %q", got, want)
	// }
}
