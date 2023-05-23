// Put the tests in a separate (but related) package, so we avoid cyclic imports of json
package json_test

import (
	"strings"
	"testing" // We still need to import the standard testing package

	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"github.com/jpsember/golang-base/jt"
)

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

func TestJSMapPrettyPrint(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	j.SetVerbose()

	var jsmap = JSMapFromString(text1)
	var s = jsmap.String()

	j.GenerateMessage(s)
	j.AssertGenerated()
}

func TestPrintJSMapToString(t *testing.T) {
	j := jt.New(t)

	var p JSONParser
	p.WithText(text1)
	var jsmap = p.ParseMap()

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

func corrupt(j *jt.J, s string) string {
	var b = []byte(s)
	i := j.Rand().Intn(len(b))
	k := j.Rand().Intn(len(newBytes))
	var c = CopyOfBytes(b)
	c[i] = newBytes[k]
	return string(c)
}
