package base

import (
	"math/rand"
	"time"
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
			// This global counter is used to avoid generating the same seed if multiple Rand()'s are generated in a very short period
			extraRandTicker++
			r.seed = time.Now().UnixNano() + extraRandTicker
		}
		r.random = rand.New(rand.NewSource(r.seed))
		r.built = true
	}
	return r.random
}

var extraRandTicker int64
