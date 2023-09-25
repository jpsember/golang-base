package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

const WebServDebug = true

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

var contentTypeNames = []string{
	"image/jpeg", //
	"image/png",  //
}

var contentTypeSignatures = []byte{
	3, 0xff, 0xd8, 0xff, // jpeg
	8, 137, 80, 78, 71, 13, 10, 26, 10, // png
}

func InferContentTypeFromBlob(data []byte) string {
	result := ""

	sig := contentTypeSignatures

	i := 0
	for fn := 0; i < len(sig); fn++ {
		recSize := int(sig[i])
		j := i + 1
		i += recSize + 1
		match := true
		for k := 0; k < recSize; k++ {
			if sig[j+k] != data[k] {
				match = false
				break
			}
		}
		if match {
			result = contentTypeNames[fn]
			break
		}
	}

	if result == "" {
		Alert("#50<1Failed to determine content-type from bytes:", CR,
			HexDumpWithASCII(ByteSlice(data, 0, 16)))
		result = "application/octet-stream"
	}
	return result
}

const DOT_DELIMITER = '.'

func AssertNoDots(expr string) string {
	if FirstDot(expr) >= 0 {
		BadArg("Expression must not have dots:", QUO, expr)
	}
	return expr
}

func FirstDot(expr string) int {
	return strings.IndexByte(expr, DOT_DELIMITER)
}

// Join nonempty strings with '.' delimiter; if no nonempty strings, return "".
func DotJoin(args ...string) string {
	s := strings.Builder{}
	for _, arg := range args {
		if arg != "" {
			if s.Len() != 0 {
				s.WriteByte(DOT_DELIMITER)
			}
			s.WriteString(arg)
		}
	}
	return s.String()
}

// Extract the first argument from a dotted expression (xxxx.yyyy.zzzz) and return xxxx, yyyy.zzzz; or "","" if none exist
func ExtractFirstDotArg(expr string) (string, string) {
	pr := PrIf("ExtractFirstDotArg", false)
	pr("ExtractFirstDotArg")
	var arg, remainder string
	for {
		pr("expression:", QUO, expr)
		dotPos := FirstDot(expr)
		pr("dotPos:", dotPos)
		if dotPos < 0 {
			arg, remainder = expr, ""
			break
		}
		// If there's a leading dot, remove it and repeat (so we don't return an empty argument unless there are no more)
		if dotPos == 0 {
			pr("leading dot")
			expr = expr[1:]
		} else {
			arg, remainder = expr[0:dotPos], expr[dotPos+1:]
			break
		}
	}
	pr("returning:", QUO, arg, QUO, remainder)
	return arg, remainder
}
