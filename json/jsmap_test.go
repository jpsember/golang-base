// Put the tests in a separate (but related) package, so we avoid cyclic imports of json
package json_test

import (
	"strconv"
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
	var badtext = `{"name" : "John", "age":30, "hobbies" : ["swimming", "coding"], "sign": "alpha bravo charlie" }`

	var results = NewJSMap()
	for i := 0; i < 100; i++ {

		var s = badtext
		s = corrupt(j, s)
		j.Log("s:", s)
		var p JSONParser
		p.WithText(s)
		p.ParseMap()
		if p.Error != nil {
			var q = NewJSMap()
			q.Put("", s)
			q.Put("err", p.Error.Error())
			results.Put(strconv.Itoa(i), q)
		}
	}
	j.AssertMessage(results)
}

var newBytes = []byte("truefalse\":,{}")

func corrupt(j *jt.J, s string) string {
	var b = []byte(s)
	i := j.Rand().Intn(len(b))
	k := j.Rand().Intn(len(newBytes))
	var c = CopyOfBytes(b)

	//Pr("s:", s)
	//Pr("i:", i)
	//Pr("k:", k)
	//Pr("len b:", len(b))
	//Pr("len newbytes:", len(newBytes))
	c[i] = newBytes[k]
	return string(c)
}
