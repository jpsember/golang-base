package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type CellWeightList = *Array[int]

func NewCellWeightList() CellWeightList {
	return NewArray[int]()
}

