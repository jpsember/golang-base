// Put the tests in a separate (but related) package, so we avoid cyclic imports of json
package json_test

import (
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
