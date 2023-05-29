package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This should probably just be replaced by an array of integers

type CellWeightList struct {
	weights []int
}

func NewCellWeightList() CellWeightList {
	return CellWeightList{
		weights: []int{},
	}
}

func (w CellWeightList) Set(index int, weight int) {
	Pr("Set index:", index, "weight:", weight, "len:", len(w.weights))
	w.GrowTo(1 + index)
	Pr("weights:", len(w.weights))
	w.weights[index] = weight
}

func (w CellWeightList) GrowTo(size int) {
	Pr("grow to:", size, "current:", len(w.weights))
	for len(w.weights) < size {
		w.weights = append(w.weights, 0)
		Pr("grown to:", len(w.weights))
	}
	Pr("w weights:", w.weights, len(w.weights))
}

func (w CellWeightList) Get(index int) int {
	if len(w.weights) <= index {
		return 0
	}
	return w.weights[index]
}
