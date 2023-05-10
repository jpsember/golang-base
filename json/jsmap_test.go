// Put the tests in a separate (but related) package, so we avoid cyclic imports of json
package json_test

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"github.com/jpsember/golang-base/jt"
	"testing" // We still need to import the standard testing package
)

// This gets rid of the 'unused import' compile error, and
// as a bonus, lets us type 'pr' without capitalization.
// I *think* it doesn't modify the code at all (i.e., there's
// no difference between calling Pr(...) and pr(...).
var pr = Pr

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
	// Get our tester that wraps the standard one
	j := jt.New(t)
	j.SetVerbose()

	var p JSONParser
	p.WithText(text1)
	var jsmap = p.ParseMap()

	Todo("can we create a utility method for this?")
	var printer = NewJSONPrinter(true)
	jsmap.PrintTo(printer)
	var s = printer.GetPrintResult()

	j.GenerateMessage(s)
	j.AssertGenerated()
}

func TestPrintJSMapToString(t *testing.T) {
	// Get our tester that wraps the standard one
	j := jt.New(t)
	j.SetVerbose()

	var p JSONParser
	p.WithText(text1)
	var jsmap = p.ParseMap()

	var s = ToString(jsmap)
	fmt.Println("tostring jsmap:", s)
	Halt()

	j.GenerateMessage(s)
	j.AssertGenerated()
}
