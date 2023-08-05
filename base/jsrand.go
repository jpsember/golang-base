package base

import (
	"math/rand"
)

type jsRandStruct struct {
	random *rand.Rand
	seed   int64
	built  bool
}

type JSRand = *jsRandStruct

func NewJSRand() JSRand {
	return &jsRandStruct{}
}

func (r JSRand) SetSeed(seed int) JSRand {
	r.seed = int64(seed)
	r.built = false
	return r
}

func (r JSRand) Rand() *rand.Rand {
	if !r.built {
		if r.seed == 0 {
			r.seed = CurrentTimeMs()
		}
		r.random = rand.New(rand.NewSource(r.seed))
		r.built = true
	}
	return r.random
}
