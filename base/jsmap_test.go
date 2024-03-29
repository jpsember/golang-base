package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"strings"
	"testing"
)

func TestJSMapPrettyPrint(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	j.SetVerbose()

	var jsmap = JSMapFromStringM(text1)
	var s = jsmap.String()

	j.GenerateMessage(s)
	j.AssertGenerated()
}

var text1 = `
  {"name" : "John", 
   "age":30, 
    "hobbies" : [
		"swimming", "coding",
	],
	"Ñio" : 42,
	"newlines": "alpha\nbravo\ncharlie",
  }
`

func TestPrintJSMapToString(t *testing.T) {
	j := jt.New(t)

	var p JSONParser
	p.WithText(text1)
	var jsmap = CheckOkWith(p.ParseMap())

	var s = ToString(jsmap)

	j.GenerateMessage(s)
	j.AssertGenerated()
}

func TestBadInput1(t *testing.T) {
	j := jt.New(t)
	var badtext = `{"nm":"J","ag":30, "hs": ["sw","co"], "si":"al be ch" }`

	var results = NewJSMap()
	for i := 0; i < 100; i++ {
		var s = badtext
		if i == 0 {
			results.PutNumbered(replaceQuotes(s))
		}
		s = corrupt(j, s)
		var p JSONParser
		p.WithText(s)
		p.ParseMap()
		if p.Error != nil {
			var q = NewJSMap()
			q.Put("", replaceQuotes(s))
			q.Put("err", replaceQuotes(p.Error.Error()))
			results.PutNumbered(q)
		}
	}
	j.AssertMessage(results)
}

func replaceQuotes(value string) string {
	return strings.ReplaceAll(value, "\"", "'")
}

var newBytes = []byte("abc%\":,{}")

func TestBadInput2(t *testing.T) {
	j := jt.New(t)
	b := strings.Builder{}
	var k = 1000
	for i := 0; i < k; i++ {
		b.WriteString(`{"":`)
	}
	b.WriteString(`"hi"`)
	for i := 0; i < k; i++ {
		b.WriteString(`}`)
	}
	var p JSONParser
	p.WithText(b.String())
	p.ParseMap()

	j.AssertMessage(p.Error)
}

func TestBadInput3(t *testing.T) {
	j := jt.New(t)
	b := strings.Builder{}
	var k = 1000
	for i := 0; i < k; i++ {
		b.WriteString(`[`)
	}
	b.WriteString(`"hi"`)
	for i := 0; i < k; i++ {
		b.WriteString(`]`)
	}
	var p JSONParser
	p.WithText(b.String())
	p.ParseList()

	j.AssertMessage(p.Error)
}

func corrupt(j jt.JTest, s string) string {
	var b = []byte(s)
	i := j.Rand().Intn(len(b))
	k := j.Rand().Intn(len(newBytes))
	var c = CopyOfBytes(b)
	c[i] = newBytes[k]
	return string(c)
}

func TestEscapes(t *testing.T) {
	j := jt.New(t)
	s := `{"":"^(\\w|-|\\.|\\x20|'|\\(|\\)|,)+$"}`
	var p JSONParser
	p.WithText(s)
	var jsmap = CheckOkWith(p.ParseMap())
	j.AssertMessage(jsmap)
}

func TestGenerateDir(t *testing.T) {
	j := jt.Newz(t)

	const tree1 = `
{"a.txt" : "",
 "b.txt" : "",
 "c"     : {"d.txt":"", "e.txt":"", "f" : {"g.txt" : ""}},
}
`

	var jsmap = JSMapFromStringM(tree1)
	j.GenerateSubdirs("source", jsmap)

	j.AssertGenerated()
}

func TestParseUnicode(t *testing.T) {
	if true {
		return
	}
	j := jt.Newz(t)

	str := "a’b"
	m := NewJSMap()
	m.Put("", str)

	s := m.String()
	m2 := JSMapFromStringM(s)
	j.AssertMessage(m2)
}

func TestEncodeBase64(t *testing.T) {
	js := jt.New(t)

	step := 1
	for j := 0; j < 2000; j += step {
		step += 2

		bytes := []byte{}
		for k := 0; k < j; k++ {
			bytes = append(bytes, byte(k))
		}

		str := EncodeBase64Maybe(bytes)
		bytes2 := DecodeBase64Maybe(str)

		js.AssertEqual(bytes, bytes2)
	}
}
