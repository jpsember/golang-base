package webserv

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
	w.GrowTo(1 + index)
	w.weights[index] = weight
}

func (w CellWeightList) GrowTo(size int) {
	for len(w.weights) < size {
		w.weights = append(w.weights, 0)
	}
}

func (w CellWeightList) Get(index int) int {
	if len(w.weights) <= index {
		return 0
	}
	return w.weights[index]
}
