package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// Escaper interface performs html escaping on its argument
type Escaper interface {
	Escaped() string
}

var HTMLRand = NewJSRand().SetSeed(1965)

const bgColors = "#fc7f03#fcce03#58bf58#4aa3b5#cfa8ed#fa7fc1#b2f7a6#b2f7a6#90adad#3588cc#b06dfc"
const colorExprLen = 7
const numColors = len(bgColors) / colorExprLen

func DebugColor(index int) string {
	j := (index & 0xffff) % numColors
	c := j * colorExprLen
	return bgColors[c : c+colorExprLen]
}

func DebugColorForString(str string) string {
	return DebugColor(SimpleHashCodeForBytes([]byte(str)))
}

func SimpleHashCodeForBytes(bytes []byte) int {
	sum := 0
	for _, x := range bytes {
		sum += int(x)
	}
	return MaxInt(sum&0xffff, 1)
}

var alignStrs = []string{
	"", "text-center", "text-left", "text-right",
}

func TextAlignStr(align WidgetAlign) string {
	// Wtf, keeps changing:  https://stackoverflow.com/questions/15446189
	return alignStrs[align]
}
