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

var Experiment = false && Alert("experiment is in effect")

var exitOnPanic = false

func ExitOnPanic() {
	exitOnPanic = true
	Alert("?<1Exiting program on panics")
}

var Dashes = "------------------------------------------------------------------------------------------\n"

// Return location of current program as a string.
// Relies on debug.Stack() being available.
// Returns source filename (without its path) and the line number, e.g., "foo.go:78".
// A skipCount of zero returns the immediate caller's location.
func CallerLocation(skipCount int) string {
	st := GenerateStackTrace(skipCount + 1)
	if len(st.Elements) != 0 {
		return st.Elements[0].StringBrief()
	}
	return "<no location available!>"
}

func Caller() string {
	return CallerLocation(2)
}

func Callers(skipStart, count int) string {
	x := GenerateStackTrace(1 + skipStart).Elements
	x = ClampedSlice(x, 0, 0+count)
	if len(x) == 0 {
		return "<no location available!>"
	}
	sb := strings.Builder{}
	for _, y := range x {
		sb.WriteString(y.StringBrief())
		sb.WriteByte('\n')
	}
	return sb.String()
}

func Panic(message ...any) {
	auxAbort(1, "Panic", message...)
}

func Die(message ...any) {
	auxAbort(1, "Dying", message...)
}

func Halt(message ...any) {
	auxAbort(1, "Halting", message...)
}

func NotSupported(message ...any) {
	auxAbort(1, "Not supported", message...)
}

func NotImplemented(message ...any) {
	auxAbort(1, "Not implemented", message...)
}

var Issue97 = Alert("issue97 in effect")

func isNil(value any) bool {
	return value == nil
}

func CheckNonEmpty(s string, message ...any) string {
	if s == "" {
		auxAbortWithArgument(1, "String is empty", s, message...)
	}
	return s
}

func CheckArg(valid bool, message ...any) bool {
	if !valid {
		auxAbort(1, "Bad argument", message...)
	}
	return valid
}

func BadArg(message ...any) {
	auxAbort(1, "Bad argument", message...)
}

func BadState(message ...any) {
	auxAbort(1, "Bad state", message...)
}

// Given an error, panic if it is not nil
func CheckOk(err error, message ...any) {
	auxCheckOk(1, err, message...)
}

// Given a value and an error, make sure the error is nil, and return the value
func CheckOkWith[X any](arg1 X, err error, message ...any) X {
	auxCheckOk(1, err, message...)
	return arg1
}

func auxCheckOk(skipCount int, err error, message ...any) {
	if err != nil {
		messageStr := ToString(message...)
		messageInfo := extractAlertInfo(messageStr)
		auxAbortWithArgument(1+skipCount+messageInfo.skipCount, "Unexpected error", err.Error(), messageInfo.key)
	}
}

func auxAbortWithArgument(skipCount int, prefix string, argument string, message ...any) {
	messageStr := ToString(message...)
	messageInfo := extractAlertInfo(messageStr)
	auxAbort(1+skipCount+messageInfo.skipCount, prefix, Quoted(argument)+" "+messageInfo.key)
}

func CheckState(valid bool, message ...any) {
	if !valid {
		auxAbort(1, "Invalid state", message...)
	}
}

var nestedAbortFlag bool

func auxAbort(skipCount int, prefix string, message ...any) {
	// Both the prefix and the message can contain skip information, so
	// parse and sum them

	prefixInfo := extractAlertInfo(prefix)
	messageStr := ToString(message...)
	messageInfo := extractAlertInfo(messageStr)

	netSkipCount := prefixInfo.skipCount + messageInfo.skipCount + skipCount + 1
	msg := "*** " + prefixInfo.key + "! " + messageInfo.key

	if !testAlertState {
		// Print the message to stdout in case it doesn't later get printed in this convenient way
		fmt.Println(msg)
		if nestedAbortFlag {
			fmt.Println("Nested exception:", INDENT, string(debug.Stack()))
		} else {
			nestedAbortFlag = true
			st := GenerateStackTrace(netSkipCount)
			if strings.HasPrefix(prefix, "Halting") {
				st.MaxRowsPrinted = 1
			}
			fmt.Println(st)
			nestedAbortFlag = false
		}

		if exitOnPanic {
			os.Exit(1)
		}
		Pr("about to panic:", msg)
		panic(msg)
	} else {
		TestAbortMessageLog.WriteString(msg + "\n")
	}
}

// True if we're performing unit tests on Alerts, Assertions
var testAlertState bool
var TestAbortMessageLog = strings.Builder{}
var TestAlertDuration int64

func CheckNil(result any, message ...any) {
	if result != nil {
		auxAbortWithArgument(1, "Result is not nil", ToString(result), message...)
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

func TypeOf(arg any) string {
	if arg == nil {
		return "<nil>"
	}
	return fmt.Sprint(reflect.TypeOf(arg))
}

// Print an alert if an alert with its key hasn't already been printed.
// The key is printed, along with the additional message components.
// If the key has a prefix '!', it is a "low priority" alert - if the key already
// appears in a json map stored on the desktop, it does not print it.
func auxAlert(skipCount int, key string, prompt string, additionalMessage ...any) {
	// Acquire the lock while we test and increment the current session report count for this alert
	debugLock.Lock()
	processClearAlertHistoryFlag()
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
		// Do this before locking, as it might attempt to use locks
		FindProjectDirM() //<-- but rework this... we want it to fall back on using current directory if there is no project dir

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
	locn := CallerLocation(skipCount + info.skipCount + 1)
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
		TestAbortMessageLog.WriteString(text + "\n")
	}
}

func Todo(key string, message ...any) bool {
	auxAlert(1, key, "TODO", message...)
	return true
}

// Deprecated.  So references show up in editor for easy deletion.
func ClearAlertHistory(flag bool) {
	if flag {
		Alert("<1 clearing alert history")
		clearAlertHistoryFlag = flag
	} else {
		Alert("<1 not clearing alert history")
	}
}

func processClearAlertHistoryFlag() {
	if clearAlertHistoryFlag {
		clearPriorityAlertMapFlag = true
		debugLocMap = make(map[string]int)
		priorityAlertMap = nil
		clearAlertHistoryFlag = false
	}
}

var clearAlertHistoryFlag bool
var clearPriorityAlertMapFlag bool

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
	if array == nil {
		BadArg("nil array")
	}
	result := make([]byte, len(array))
	copy(result, array)
	return result
}

func ParseInt(str string) (int, error) {
	result, err := ParseInt64(str)
	return int(result), err
}

func ParseIntM(str string) int {
	return int(ParseInt64M(str))
}

func ParseInt64(str string) (int64, error) {
	result, err := strconv.ParseInt(str, 10, 64)
	return result, err
}

func ParseInt64M(str string) int64 {
	result, err := ParseInt64(str)
	return CheckOkWith(result, err, "Failed to parse int64 from:", str)
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
	skipCount     int
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
			info.skipCount += sf
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
		priorityAlertMap = NewJSMap()
		if clearPriorityAlertMapFlag {
		} else {
			restored, err := JSMapFromFileIfExists(priorityAlertPersistPath)
			if err != nil {
				Pr("Problem parsing:", priorityAlertPersistPath, ", error:", err)
				priorityAlertMap = NewJSMap()
				// Discard old file
				priorityAlertPersistPath.DeleteFile()
			} else {
				priorityAlertMap = restored
			}
		}
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
		priorityAlertPersistPath.WriteStringM(priorityAlertMap.CompactString())
	}
	return true
}

func CurrentTimeMs() int64 {
	return time.Now().UnixMilli()
}

func SleepMs(ms int) {
	time.Sleep(MsToDuration(ms))
}

func MsToDuration(ms int) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

// Convert an array of a particular type to an array of any.
func ToAny[T any](vals []T) []any {
	s := make([]any, len(vals))
	for i, v := range vals {
		s[i] = v
	}
	return s
}

// Build a map[K]V from a sequence of arguments key0,val0,key1,val1,....
func BuildMap[K comparable, V any](keyValPairs ...any) map[K]V {
	m := make(map[K]V)
	CheckArg(len(keyValPairs)%2 == 0, "<1expected 2n elements")
	for i := 0; i < len(keyValPairs); i += 2 {
		obj1 := keyValPairs[i]
		obj2 := keyValPairs[i+1]
		key, ok1 := obj1.(K)
		CheckArg(ok1, "<1failed to cast key:", obj1)
		val, ok2 := obj2.(V)
		CheckArg(ok2, "<1failed to cast value:", obj2)
		if _, ok := m[key]; ok {
			BadArg("<1Duplicate key:", key)
		}
		m[key] = val
	}
	return m
}

func BuildStringStringMap(keyValPairs ...string) map[string]string {
	return BuildMap[string, string](ToAny(keyValPairs)...)
}

// Get value for key, returning i) default value if key doesn't exist, ii) whether it existed
func OptMapValue[K comparable, V any](m map[K]V, key K, defaultValue V) (result V, ok bool) {
	val, ok := m[key]
	if !ok {
		val = defaultValue
	}
	return val, ok
}

// Get value for key from map; fail if missing
func MapValue[K comparable, V any](m map[K]V, key K) V {
	val, ok := m[key]
	if !ok {
		BadArg("<1Key not found within map:", key)
	}
	return val
}

func Ternary[V any](flag bool, ifTrue V, ifFalse V) V {
	if flag {
		return ifTrue
	}
	return ifFalse
}

func MyMod(value int, divisor int) int {
	if divisor <= 0 {
		BadArg("<1divisor <= 0:", divisor, "value:", value)
	}

	k := value % divisor
	if value < 0 && k != 0 {
		k += divisor
	}
	return k
}

// ------------------------------------------------------------------------------------
// Strack traces
// ------------------------------------------------------------------------------------

type StackTraceStruct struct {
	Preamble       string
	Content        string
	SkipFactor     int
	MaxRowsPrinted int
	Elements       []stackTraceElement
}

type stackTraceElementStruct struct {
	Package            string
	CallerFunction     string
	CallerArguments    string
	CalleeFile         string
	CalleeLineNumber   int
	StackFramePosition string
	raw0, raw1         string
	formattedLong      string
	formattedBrief     string
}
type stackTraceElement = *stackTraceElementStruct

var repoDirOrEmpty string

func init() {
	if x, ok := FindRepoDir(); ok == nil {
		repoDirOrEmpty = x.String()
	}
}

func (e stackTraceElement) prepareStrings() {
	if e.formattedLong == "" {

		{
			val := e.raw0

			// If there is a package, it will end at the first . following the last /
			i := strings.LastIndex(val, "/")
			if i >= 0 {
				i = strFirstFrom(val, i, ".")
				e.Package = val[0:i]
			}
			j := strLastIndex(val, "(")
			e.CallerFunction = val[i+1 : j]
			e.CallerArguments = val[j:]
		}
		{
			val := strings.TrimSpace(e.raw1)
			i := strFirst(val, ":")
			callerPath := val[0:i]

			callerPath = strings.TrimPrefix(callerPath, repoDirOrEmpty)
			e.CalleeFile = NewPathM(callerPath).Base()

			// If there is a stack frame position, it will be preceded by +0x
			rem := val[i+1:]
			j := strings.Index(rem, "+0x")
			if j >= 0 {
				e.StackFramePosition = val[j:]
				rem = rem[0:j]
			}
			e.CalleeLineNumber = ParseIntM(strings.TrimSpace(rem))
		}

		// Convert the stack trace element to a display version
		s1 := e.CalleeFile + ":" + IntToString(e.CalleeLineNumber)
		e.formattedBrief = s1
		e.formattedLong = s1 + Spaces(24-len(s1)) + " " + e.CallerFunction
	}
}

func (e stackTraceElement) StringDetailed() string {
	e.prepareStrings()
	return e.formattedLong
}

func (e stackTraceElement) StringBrief() string {
	e.prepareStrings()
	return e.formattedBrief
}

type StackTrace = *StackTraceStruct

func GenerateStackTrace(skipFactor int) StackTrace {
	return NewStackTrace(string(debug.Stack()), 2+skipFactor)
}

func NewStackTrace(content string, skipFactor int) StackTrace {
	t := &StackTraceStruct{}
	t.SkipFactor = skipFactor
	t.parse(content)
	return t
}

func (st StackTrace) String() string {
	elem := st.Elements
	if st.MaxRowsPrinted > 0 {
		elem = elem[0:MinInt(st.MaxRowsPrinted, len(elem))]
	}
	sb := strings.Builder{}
	for _, x := range elem {
		sb.WriteString(x.StringDetailed())
		sb.WriteByte('\n')
	}
	return sb.String()
}

func strFirst(str string, substr string) int {
	k := strings.Index(str, substr)
	CheckArg(k >= 0, "string doesn't contain substr:", str, substr)
	return k
}
func strLastIndex(str string, substr string) int {
	k := strings.LastIndex(str, substr)
	if k < 0 {
		CheckArg(k >= 0, "string doesn't contain substring:", Quoted(str), Quoted(substr))
	}
	return k
}

func strFirstFrom(str string, from int, substr string) int {
	remainder := str[from:]
	return strFirst(remainder, substr) + from
}

func (st StackTrace) parse(content string) {
	content = strings.TrimSpace(content)

	st.Content = content

	skipped := 0
	elements := []stackTraceElement{}

	lines := strings.Split(content, "\n")

	// The first line in the stack trace, which we will call the preamble, is something like 'goroutine 1 [running]:'
	//
	// The remaining 2n lines form pairs, where each has this format:
	//
	//     {package name} {name + arguments of caller function}
	//     {file where function was called}:{line number within file}  +{relative position of the function within the stack frame}
	//
	// Here is a typical stack trace pair:
	//
	// runtime/debug.Stack()
	//         /usr/local/go/src/runtime/debug/stack.go:24 +0x65
	//
	// package:  "runtime/debug"
	// method:   "Stack"
	// args:     "()"
	//
	// filename: "/usr/local/go/src/runtime/debug/stack.go"
	// line num: "24"
	// offset:   "+0x65"
	//
	// Here are some peculiar examples of the first element of the pair:
	//
	// github.com/jpsember/golang-base/base.(*StackTraceStruct).parse(0xc0000725a0, {0xc00013a900?, 0x1000?})
	//
	// package:  "github.com/jpsember/golang-base/base"
	// method:   "(*StackTraceStruct).parse"
	// args:     "(0xc0000725a0, {0xc00013a900?, 0x1000?})"
	//
	// panic({0x1004bc580, 0x1007e6e90})
	//
	// package:  ""
	// method:   "panic"
	// args:     "({0x1004bc580, 0x1007e6e90})"
	//
	st.Preamble = lines[0]

	for cursor := 1; cursor < len(lines); cursor += 2 {
		if skipped < st.SkipFactor {
			skipped++
		} else {
			elements = append(elements,
				&stackTraceElementStruct{
					raw0: lines[cursor+0],
					raw1: lines[cursor+1],
				})
		}
	}
	st.Elements = elements
}

func CausePanic() int {
	sum := 0
	for i := -3; i < 3; i++ {
		sum += 10 / i
	}
	return sum
}

func CatchPanic(handler func()) {
	if r := recover(); r != nil {
		Pr("catching panic:", r)
		Pr(GenerateStackTrace(2))
		if handler != nil {
			handler()
		}
	}
}

func ByteSlice(bytes []byte, start int, length int) []byte {
	ln := len(bytes)

	if start < 0 {
		start = ln + start
	}
	return ClampedSlice(bytes, start, start+length)
}

func BinaryN(value2 int) string {
	value := uint64(value2)
	var digits int
	if value >= 0x100000000000000 {
		digits = 64
	} else if value >= 0x1000000 {
		digits = 32
	} else if value >= 0x10000 {
		digits = 24
	} else if value >= 0x100 {
		digits = 16
	} else {
		digits = 8
	}

	sb := strings.Builder{}
	for i := 0; i < digits; i++ {
		var ch byte = '.'
		if value&(1<<(digits-1-i)) != 0 {
			ch = '1'
		}
		sb.WriteByte(ch)
	}
	sb.WriteString(" $")
	ndig := (digits + 3) / 4
	appendHex(&sb, uint64(value), ndig)
	sb.WriteString(" #")
	sb.WriteString(strconv.FormatUint(value, 10))
	return sb.String()
}

func appendHex(sb *strings.Builder, value uint64, ndigits int) {
	for ch := 0; ch < ndigits; ch++ {
		shiftCount := (ndigits - 1 - ch) << 2
		val := int((value >> shiftCount) & 0xf)

		var c byte
		if val < 10 {
			c = ('0' + byte(val))
		} else {
			c = ('a' + byte(val-10))
		}
		sb.WriteByte(c)
	}
}

func ToHex(value uint64, ndigits int) string {
	s := strings.Builder{}
	appendHex(&s, value, ndigits)
	return s.String()
}

func HexDump(byteArray []byte) string {
	return hexDump(byteArray, false)
}

func HexDumpWithASCII(byteArray []byte) string {
	return hexDump(byteArray, true)
}

func hexDump(byteArray []byte, withASCII bool) string {
	sb := strings.Builder{}

	const groupSize = 1 << 2 // Must be power of 2

	const rowSize = 16
	const hideZeros = true
	const groups = true

	length := len(byteArray)
	i := 0
	for i < length {
		rSize := rowSize
		if rSize+i > length {
			rSize = length - i
		}
		address := i
		appendHex(&sb, uint64(address), 4)
		sb.WriteString(`: `)
		if groups {
			sb.WriteString("| ")
		}
		for j := 0; j < rowSize; j++ {
			if j < rSize {
				val := byteArray[i+j]
				if hideZeros && val == 0 {
					sb.WriteString("  ")
				} else {
					appendHex(&sb, uint64(val), 2)
				}
			} else {
				sb.WriteString("  ")
			}
			sb.WriteByte(' ')
			if groups {
				if (j & (groupSize - 1)) == groupSize-1 {
					sb.WriteString("| ")
				}
			}
		}
		if withASCII {
			sb.WriteByte(' ')

			for j := 0; j < rSize; j++ {
				v := byteArray[i+j]
				if v < 0x20 || v >= 0x80 {
					v = '.'
				}
				sb.WriteByte(v)
				if groups && ((j & (groupSize - 1)) == groupSize-1) {
					sb.WriteByte(' ')
				}
			}
		}
		sb.WriteByte('\n')
		i += rSize
	}
	return sb.String()

}

func StringFromOptError(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}

func UpdateErrorWithString(err error, message string) error {
	if err == nil && message != "" {
		err = Error(message)
	}
	return err
}

// If string has a prefix, return string with prefix removed and true; else, original string and false
func TrimIfPrefix(text string, prefix string) (string, bool) {
	if strings.HasPrefix(text, prefix) {
		return text[len(prefix):], true
	}
	return text, false
}

var IntegerOutOfRangeError = Error("integer is out of range")

// Attempt to parse value as positive integer
func ParseAsPositiveInt(text string) (int, error) {
	value1, err := ParseInt(text)
	value := int(value1)
	if err == nil && value <= 0 {
		err = IntegerOutOfRangeError
	}
	if err != nil {
		value = 0
	}
	return value, err
}

func ClampedSlice[K any](slice []K, start int, end int) []K {
	start = Clamp(start, 0, len(slice))
	end = Clamp(end, start, len(slice))
	return slice[start:end]
}

func ReportIfError(err error, msg ...any) bool {
	if err != nil {
		Alert("#50<1Error occurred, ignoring!  Error:", err, INDENT, "Message:", ToString(msg...))
		return true
	}
	return false
}

func Last[T any](slice []T) T {
	i := len(slice)
	return slice[i-1]
}

func PopLast[T any](slice []T) (T, []T) {
	i := len(slice)
	return slice[i-1], slice[:i-1]
}

func DeleteSliceElements[T any](slice []T, delStart int, delCount int) []T {
	return append(slice[:delStart], slice[delStart+delCount:]...)
}

func Truncated(arg any) string {
	switch v := arg.(type) {
	case nil:
		return "<nil>"
	case string:
		return trunc(v)
	case JSMap:
		return trunc(PrintJSEntity(v, false))
	case JSList:
		return trunc(PrintJSEntity(v, false))
	default:
		return trunc(ToString(v))
	}
}
func trunc(x string) string {
	const maxLen = 75
	if len(x) > maxLen {
		return x[0:maxLen] + "..."
	}
	return x
}

const (
	JMs   = 1
	JSec  = JMs * 1000
	JMin  = JSec * 60
	JHour = JMin * 60
)
