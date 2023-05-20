package base

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

var Dashes = "------------------------------------------------------------------------------------------\n"

// Return location of current program as a string.
// Relies on debug.Stack() being available.
// Returns source filename (without its path) and the line number, e.g., "foo.go:78".
func CallerLocation(skipCount int) string {
	var db = false

	w := string(debug.Stack())
	if db {
		Pr("debug.Stack:\n", Dashes, w, Dashes)
	}
	lines := strings.Split(w, "\n")
	if db {
		Pr("lines:\n", Dashes)
		for _, x := range lines {
			Pr("[" + x + "]")
		}
	}

	cursor := 2 + skipCount*2
	if cursor >= len(lines) {
		return "<cannot parse debug_loc (0)>"
	}
	line := strings.TrimSpace(lines[cursor])
	if db {
		Pr("line:\n", line)
	}

	// Trim the +0xHHH from the end (if there is one)
	cutoff := strings.LastIndex(line, " +0x")
	if cutoff < 0 {
		cutoff = len(line)
	}

	// Trim any path components up to last '/'
	j := 0
	i := strings.LastIndex(line, "/")
	if i < cutoff {
		j = i + 1
	}
	return line[j:cutoff]
}

func Die(message ...any) {
	panic("*** Dying (" + CallerLocation(3) + ") " + ToString(message...))
}

func Halt(message ...any) {
	var text = "*** Halting (" + CallerLocation(3) + ")"
	if len(message) != 0 {
		text += ": " + ToString(message...)
	}
	Pr(text)
	os.Exit(1)
}

func NotSupported(message ...any) {
	panic("*** Not supported (" + CallerLocation(3) + ") " + ToString(message...))
}

func NotImplemented(message ...any) {
	panic("*** Not implemented (" + CallerLocation(3) + ") " + ToString(message...))
}

func CheckNotNil(value any, message ...any) any {
	if value == nil {
		str := "*** Argument is nil (" + CallerLocation(3) + ") "
		if len(message) != 0 {
			str = str + "; \n" + ToString(message...)
		}
		panic(str)
	}
	return value
}

func CheckArg(valid bool, message ...any) bool {
	if !valid {
		BadArgWithSkip(3, message...)
	}
	return valid
}

func BadArgWithSkip(skipCount int, message ...any) {
	panic("*** Bad argument!  (" + CallerLocation(skipCount+1) + ") " + ToString(message...))
}

func BadArg(message ...any) {
	BadArgWithSkip(4, message)
}

func BadStateWithSkip(skipCount int, message ...any) {
	panic("*** Bad state!  (" + CallerLocation(skipCount+1) + ") " + ToString(message...))
}

func BadState(message ...any) {
	BadStateWithSkip(4, message)
}

func CheckState(valid bool, message ...any) {
	if !valid {
		panic("*** Invalid state! (" + CallerLocation(3) + ") " + ToString(message...))
	}
}

// Panic if an error code is nonzero.
func CheckOk(err error, message ...any) {
	CheckOkWithSkip(2, err, message)
}

// Panic if an error code is nonzero.
func CheckOkWithSkip(skipCount int, err error, message ...any) {
	if err != nil {
		panic("*** Error returned: (" + CallerLocation(3+skipCount) + ") " + err.Error() + "; " + ToString(message...))
	}
}

func CheckNil(result any, message ...any) {
	if result != nil {
		str := "*** Result is not nil! (" + CallerLocation(3) + ") " + ToString(result)
		if len(message) != 0 {
			str = str + "; \n" + ToString(message...)
		}
		panic(str)
	}
}

func Empty(text string) bool {
	return len(text) == 0
}

func NonEmpty(text string) bool {
	return !Empty(text)
}

// Get information about a variable; its value, and its type
func Info(arg any) string {
	if arg == nil {
		return "<nil>"
	}
	// Avoid calling BasePrinter for this, since it might cause endless recursion
	return fmt.Sprint("Value[", arg, "],Type[", reflect.TypeOf(arg), "]")
}

// Print an Alert if an Alert with its key hasn't already been printed.
// The key is printed, along with the additional message components
func auxAlert(key string, prompt string, additionalMessage ...any) {
	if !debugLocMap[key] {
		var output strings.Builder
		output.WriteString("***")
		output.WriteString(" ")
		output.WriteString(prompt)
		output.WriteString(": ")
		if len(additionalMessage) != 0 {
			output.WriteString(key + " " + ToString(additionalMessage...))
		} else {
			output.WriteString(key)
		}
		locn := CallerLocation(4)
		output.WriteString(" (")
		output.WriteString(locn)
		output.WriteString(")")
		fmt.Println(output.String())
		debugLocMap[key] = true
	}
}

func Todo(key string, message ...any) bool {
	auxAlert(key, "TODO", message...)
	return true
}

// Print an Alert if an Alert with its key hasn't already been printed.
// The key is printed, along with the additional message components
func Alert(key string, additionalMessage ...any) bool {
	auxAlert(key, "WARNING", additionalMessage...)
	return true
}

var debugLocMap = make(map[string]bool)

func Quoted(x string) string {
	return "\"" + x + "\""
}

var DASHES = "\n----------------------------------------------------------------------------------\n"

var SPACES = "                                                                " +
	"                                                                " +
	"                                                                " +
	"                                                                "

// Get string of zero or more spaces; if count < 0, returns empty string.
func Spaces(count int) string {
	if count < 0 {
		count = 0
	}
	if count <= len(SPACES) {
		return SPACES[0:count]
	}
	return SPACES + Spaces(count-len(SPACES))
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AbsInt32(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}
func AbsLong(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// Construct an error with printed arguments
func Error(message ...any) error {
	var s = ToString(message)
	return errors.New(s)
}

func HasKey[K comparable, V any](m map[K]V, key K) bool {
	var _, result = m[key]
	return result
}

// ---------------------------------------------------------------------------------------
// Generated data type interface
// ---------------------------------------------------------------------------------------

// We can include it here, because it doesn't reference any external dependencies (e.g. JSEntity)

type DataClass interface {
	fmt.Stringer
	ToJson() any // This should return a JSEntity, to be defined elsewhere
	Parse(source any) DataClass
}
