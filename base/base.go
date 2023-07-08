package base

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Dashes = "------------------------------------------------------------------------------------------\n"

// Return location of current program as a string.
// Relies on debug.Stack() being available.
// Returns source filename (without its path) and the line number, e.g., "foo.go:78".
// A skipCount of zero returns the immediate caller's location.
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

	cursor := 2 + (2+skipCount)*2
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
	auxPanic(1, "Dying", message...)
}

func Halt(message ...any) {
	text := preparePanicMessage(1, "Halting", message...)
	if !testAlertState {
		Pr(text)
		os.Exit(1)
	} else {
		TestPanicMessageLog.WriteString(text + "\n")
	}
}

func NotSupported(message ...any) {
	auxPanic(1, "Not supported", message...)
}

func NotImplemented(message ...any) {
	auxPanic(1, "Not implemented", message...)
}

func IsNil(value any) bool {
	return value == nil
}

func CheckNotNil[T any](value T, message ...any) T {
	if IsNil(value) {
		auxPanic(1, "Argument is nil", message...)
	}
	return value
}

func CheckNonEmpty(s string, message ...any) string {
	if s == "" {
		BadArgWithSkip(1, JoinElementToList("string is empty;", message))
	}
	return s
}

func CheckArg(valid bool, message ...any) bool {
	if !valid {
		BadArgWithSkip(1, message...)
	}
	return valid
}

// A skip count of 0 reports the immediate caller's location
func BadArgWithSkip(skipCount int, message ...any) {
	auxPanic(skipCount+1, "Bad argument", message...)
}

func BadArg(message ...any) {
	BadArgWithSkip(1, message...)
}

func BadStateWithSkip(skipCount int, message ...any) {
	auxPanic(skipCount+1, "Bad state", message...)
}

func BadState(message ...any) {
	BadStateWithSkip(4, message...)
}

// Given a value and an error, make sure the error is nil, and return just the value
func AssertNoError[X any](arg1 X, err error) X {
	CheckOkWithSkip(1, err)
	return arg1
}

func CheckState(valid bool, message ...any) {
	if !valid {
		auxPanic(1, "Invalid state", message...)
	}
}

func preparePanicMessage(skipCount int, prefix string, message ...any) string {
	return CallerLocation(skipCount+1) + " *** " + prefix + "! " + ToString(message...)
}

func auxPanic(skipCount int, prefix string, message ...any) {
	msg := preparePanicMessage(skipCount+1, prefix, message...)
	if !testAlertState {
		panic(msg)
	} else {
		TestPanicMessageLog.WriteString(msg + "\n")
	}
}

// True if we're performing unit tests on Alerts, Assertions
var testAlertState bool
var TestPanicMessageLog = strings.Builder{}
var TestAlertDuration int64

// Panic if an error code is nonzero.
func CheckOk(err error, message ...any) {
	CheckOkWithSkip(1, err, message...)
}

// Panic if an error code is nonzero.
// Deprecated.  Use skip expression "<\d+"
func CheckOkWithSkip(skipCount int, err error, message ...any) {
	if err != nil {
		auxPanic(skipCount+1, "Error returned", JoinElementToList(err.Error(), message)...)
	}
}

func CheckNil(result any, message ...any) {
	if result != nil {
		auxPanic(1, "Result is not nil", JoinElementToList(ToString(result)+"; \n", message)...)
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

// Print an alert if an alert with its key hasn't already been printed.
// The key is printed, along with the additional message components.
// If the key has a prefix '!', it is a "low priority" alert - if the key already
// appears in a json map stored on the desktop, it does not print it.
func auxAlert(skipCount int, key string, prompt string, additionalMessage ...any) {
	// Acquire the lock while we test and increment the current session report count for this alert
	debugLock.Lock()
	info := extractAlertInfo(key)
	cachedInfo := debugLocMap[info.key] + 1
	debugLocMap[info.key] = cachedInfo
	debugLock.Unlock()

	// If we are never to print this alert, exit now
	if info.maxPerSession == 0 {
		return
	}

	// If there's a multi-session priority value, process it
	//
	if info.delayMs > 0 {
		debugLock.Lock()
		flag := processAlertForMultipleSessions(info)
		debugLock.Unlock()
		if !flag {
			return
		}
	} else {
		// If we've exceeded the max per session count, exit now
		if cachedInfo > info.maxPerSession {
			return
		}
	}

	var output strings.Builder
	locn := CallerLocation(skipCount + info.skipFactor + 1)
	output.WriteString(locn)
	output.WriteString(" ***")
	output.WriteString(" ")
	output.WriteString(prompt)
	output.WriteString(": ")
	if len(additionalMessage) != 0 {
		output.WriteString(info.key + " " + ToString(additionalMessage...))
	} else {
		output.WriteString(info.key)
	}

	text := output.String()
	if !testAlertState {
		fmt.Println(text)
	} else {
		TestPanicMessageLog.WriteString(text + "\n")
	}
}

func Todo(key string, message ...any) bool {
	auxAlert(1, key, "TODO", message...)
	return true
}

// Print an Alert if an Alert with its key hasn't already been printed.
// The key is printed, along with the additional message components
func Alert(key string, additionalMessage ...any) bool {
	auxAlert(1, key, "WARNING", additionalMessage...)
	return true
}

var debugLocMap = make(map[string]int)
var debugLock sync.RWMutex

func Quoted(x string) string {
	return "\"" + x + "\""
}

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

const dotsString = ".........................................................................................................."
const dotsStringLength = len(dotsString)

// Get string of zero or more periods; if count < 0, returns empty string.
func Dots(count int) string {
	if count < 0 {
		count = 0
	}
	if count <= dotsStringLength {
		return dotsString[0:count]
	}
	return dotsString + Dots(count-dotsStringLength)
}

func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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
	var s = ToString(message...)
	return errors.New(s)
}

func HasKey[K comparable, V any](m map[K]V, key K) bool {
	var _, result = m[key]
	return result
}

// ---------------------------------------------------------------------------------------
// Generated data type interface
// ---------------------------------------------------------------------------------------

type DataClass interface {
	fmt.Stringer
	ToJson() JSEntity
	Parse(source JSEntity) DataClass
}

var regexpCache = &sync.Map{}

func Regexp(expr string) *regexp.Regexp {
	value, ok := regexpCache.Load(expr)
	if ok {
		return value.(*regexp.Regexp)
	}
	pat, err := regexp.Compile(expr)
	CheckOk(err, "trouble compiling regexp:", Quoted(expr))
	regexpCache.Store(expr, pat)
	return pat
}

func JoinElementToList(obj any, list2 []any) []any {
	return JoinLists([]any{obj}, list2)
}

func JoinLists(list1 []any, list2 []any) []any {
	result := make([]any, 0, len(list1)+len(list2))
	result = append(result, list1...)
	result = append(result, list2...)
	return result
}

// Move this to some other package later
func CopyOfBytes(array []byte) []byte {
	CheckNotNil(array)
	result := make([]byte, len(array))
	copy(result, array)
	return result
}

func ParseInt(str string) (int64, error) {
	result, err := strconv.ParseInt(str, 10, 64)
	return result, err
}

func ParseIntM(str string) int {
	result, err := ParseInt(str)
	CheckOk(err, "Failed to parse int from:", Quoted(str))
	return int(result)
}

func IntToString(value int) string {
	return strconv.Itoa(value)
}

// ------------------------------------------------------------------------------------
// Alerts with priorities
// ------------------------------------------------------------------------------------

func SetTestAlertInfoState(state bool) {
	if state {
		testAlertState = true
		priorityAlertMap = NewJSMap()
	} else {
		testAlertState = false
		priorityAlertMap = nil
	}
	TestAlertDuration = 0
}

var priorityAlertPersistPath Path
var priorityAlertMap JSMap

type alertInfo struct {
	key           string // The string used to access the report count for this alert
	delayMs       int64
	maxPerSession int
	skipFactor    int
}

// Parse an alert key into an alertInfo structure.
// Can contain zero or more prefixes of the form:
//
// -		       		Never print
// !					Print about once per day
// ?		       		Print about once per month
// #[0-9]+              Print n times, every time program is run
// <[0-9]+              Skip first n entries in stack trace
func extractAlertInfo(key string) alertInfo {

	const minute = 60 * 1000
	const hour = minute * 60

	info := alertInfo{
		maxPerSession: 1,
	}
	cursor := 0
	lkey := len(key)
	for cursor < lkey {
		ch := key[cursor]
		cursor++
		if ch == '-' {
			info.maxPerSession = 0
			break
		}
		if ch == '!' {
			info.delayMs = hour * 24
		} else if ch == '?' {
			info.delayMs = hour * 24 * 31
		} else if ch == '#' {
			cursor, info.maxPerSession = extractInt(key, cursor)
		} else if ch == '<' {
			var sf int
			cursor, sf = extractInt(key, cursor)
			info.skipFactor += sf
		} else if ch == ' ' {
			// ignore leading spaces
		} else {
			cursor--
			break
		}
	}
	info.key = key[cursor:]
	return info
}

func extractInt(s string, cursor int) (newCursor int, value int) {
	sLen := len(s)
	newCursor = cursor
	value = 0
	for newCursor < sLen {
		ch := s[newCursor]
		if ch < '0' || ch > '9' {
			break
		}
		value = value*10 + (int)(ch-'0')
		newCursor++
	}
	return
}

func processAlertForMultipleSessions(info alertInfo) bool {
	if priorityAlertMap == nil {

		// Look for a project directory, a git repository, or the current directory, in that order, for a file named .go_flags.json

		d, _ := FindProjectDir()
		if d.Empty() {
			d, _ = AscendToDirectoryContainingFile("", ".git")
			if d.Empty() {
				d = CurrentDirectory()
			}
		}
		priorityAlertPersistPath = d.JoinM(".go_flags.json")
		priorityAlertMap = JSMapFromFileIfExistsM(priorityAlertPersistPath)
		const expectedVersion = 2
		if priorityAlertMap.OptInt("version", 0) != expectedVersion {
			priorityAlertMap.Clear().Put("version", expectedVersion)
		}
	}

	m := priorityAlertMap.OptMapOrEmpty(info.key)
	currTime := CurrentTimeMs()
	elapsed := TestAlertDuration
	if elapsed == 0 {
		lastReport := m.OptLong("r", 0)
		elapsed = currTime - lastReport
	}
	CheckArg(elapsed >= 0)
	if elapsed < info.delayMs {
		return false
	}
	m.Put("r", currTime)
	priorityAlertMap.Put(info.key, m)
	if !testAlertState {
		priorityAlertPersistPath.WriteStringM(priorityAlertMap.String())
	}
	return true
}

func CurrentTimeMs() int64 {
	return int64(time.Now().Unix())
}

func SleepMs(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))
}
