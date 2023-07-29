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

const bgColors = "#fc7f03#fcce03#58bf58#4aa3b5#cfa8ed#fa7fc1#b2f7a6#b2f7a6#90adad#3588cc#b06dfc"
const colorExprLen = 7
const numColors = len(bgColors) / colorExprLen

func DebugColor(index int) string {
	j := (index & 0xffff) % numColors
	c := j * colorExprLen
	return bgColors[c : c+colorExprLen]
}
