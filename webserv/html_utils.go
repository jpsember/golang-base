package webserv

import "math/rand"

// Escaper interface performs html escaping on its argument
type Escaper interface {
	Escaped() string
}

func RandSeed(seed int) *rand.Rand {
	randSeed = seed
	randObj = nil
	return Rand()
}

func Rand() *rand.Rand {
	if randObj == nil {
		if randSeed == 0 {
			randSeed = 1965
		}
		randObj = rand.New(rand.NewSource(int64(randSeed)))
	}
	return randObj
}

var randSeed int
var randObj *rand.Rand
