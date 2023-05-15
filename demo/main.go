package main

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/data"
)

func main() {
	var x = NewEnumInfo("inactive running stopped")
	Pr(x)
}
