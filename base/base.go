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

func isNil(value any) bool {
	return value == nil
}

func CheckNotNil[T any](value T, message ...any) T {
	if isNil(value) {
		auxAbort(1, "Argument is nil", message...)
	}
	return value
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

			// If we're exiting as the result of a panic, omit some rows:
			// The one starting with "panic({"
			if strings.HasPrefix(prefix, "Panic") {
				// Omit every row through the one with "panic(...", as well as any immediately following
				// that have the prefix "panic.go"
				newFirstRow := 0
				foundPanic := false
				for i, v := range st.Rows {
					if foundPanic {
						if !strings.HasPrefix(v, "panic.go") {
							break
						}
						newFirstRow++
					} else if strings.HasPrefix(v, "panic(") {
						foundPanic = true
						newFirstRow = i + 1
					}
				}
				if !foundPanic {
					Pr("...did not find a stack trace row beginning with 'panic('...")
				}
				h := len(st.Rows)
				st.Rows = st.Rows[MinInt(newFirstRow, h):h]
			}
			fmt.Println(st)
			nestedAbortFlag = false
		}
		os.Exit(1)
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
func ClearAlertHistory() {
	Alert("<1 clearing alert history")
	clearAlertHistoryFlag = true
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
	return int(CheckOkWith(result, err, "Failed to parse int from:", str))
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
		if clearPriorityAlertMapFlag {
			priorityAlertMap = NewJSMap()
			clearPriorityAlertMapFlag = false
		} else {
			priorityAlertMap = JSMapFromFileIfExistsM(priorityAlertPersistPath)
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
	return time.Now().Unix()
}

func SleepMs(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))
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
	Rows           []string
	MaxRowsPrinted int
	Elements       []stackTraceElement
}

//     {package name} {name + arguments of caller function}
//     {file where function was called}:{line number within file}  +{relative position of the function within the stack frame}

type stackTraceElement struct {
	Package            string
	CallerFunction     string
	CallerArguments    string
	CalleeFile         string
	CalleeLineNumber   int
	StackFramePosition string
}

func (e stackTraceElement) String() string {
	return NewJSMap().Put("1 Package", e.Package). //
							Put("2 CallerFunction", e.CallerFunction).         //
							Put("3 CallerArguments", e.CallerArguments).       //
							Put("4 CalleeFile", e.CalleeFile).                 //
							Put("5 LineNum", e.CalleeLineNumber).              //
							Put("6 StackFramePosition", e.StackFramePosition). //
							String()
}

type StackTrace = *StackTraceStruct

func GenerateStackTrace(skipFactor int) StackTrace {
	return NewStackTrace(string(debug.Stack()), 2+skipFactor)
}

func NewStackTrace(content string, skipFactor int) StackTrace {
	//Pr("NewStackTrace with content:", INDENT, Quoted(content))
	t := &StackTraceStruct{}
	t.SkipFactor = skipFactor
	t.parse(content)
	return t
}

func (st StackTrace) String() string {
	rows := st.Rows
	if st.MaxRowsPrinted > 0 {
		rows = rows[0:MinInt(st.MaxRowsPrinted, len(rows))]
	}
	return strings.Join(rows, "\n")
}

func quickVerify(flag bool) {
	CheckState(flag)
	if !flag {
		panic("flag has failed")
	}
}

func strRemainder(str string, startIndex int) string {
	q := len(str)
	//Pr("str remainder:", Quoted(str), "len:", q, "startIndex:", startIndex)
	quickVerify(startIndex <= q)
	CheckArg(startIndex >= 0 && startIndex <= q)
	return str[startIndex:q]
}

//func IndexFrom(str string, startIndex int, substr string) int {
//	quickVerify(startIndex >= 0 && startIndex <= len(str))
//	rem := strRemainder(str, startIndex)
//	k := strings.Index(rem, substr)
//	if k >= 0 {
//		k += startIndex
//	}
//	return k
//}

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
	//quickVerify(k >= 0)
	return k
}

func strFirstFrom(str string, from int, substr string) int {
	remainder := str[from:len(str)]
	return strFirst(remainder, substr) + from
}
func (st StackTrace) parse(content string) {

	content = strings.TrimSpace(content)

	// If the paths we parse lie within the current git repo, we'll show only their filenames.
	//
	var repoDirPrefix string
	if repoDir, err := FindRepoDir(); err == nil {
		repoDirPrefix = repoDir.String()
	}

	st.Content = content

	skipped := 0
	rows := NewArray[string]()

	lines := strings.Split(content, "\n")

	// The first line in the stack trace seems to always be something like 'goroutine 1 [running]:'
	//
	// The remaining 2n lines have this format:
	//
	//     {package name} {name + arguments of caller function}
	//     {file where function was called}:{line number within file}  +{relative position of the function within the stack frame}
	//
	//
	quickVerify(len(lines)%2 == 1)
	st.Preamble = lines[0]
	quickVerify(strings.HasPrefix(st.Preamble, "goroutine"))
	for cursor := 1; cursor < len(lines); cursor += 2 {
		elem := stackTraceElement{}

		{
			val := lines[cursor+0]

			Pr("processing first line in pair:", INDENT, val)

			i := strings.LastIndex(val, "/")
			if i >= 0 {
				i = strFirstFrom(val, i, ".")
				elem.Package = val[0:i]
			}
			j := strLastIndex(val, "(")
			elem.CallerFunction = val[i+1 : j]
			elem.CallerArguments = strRemainder(val, j)
		}
		{
			val := strings.TrimSpace(lines[cursor+1])
			Pr("processing second line in pair:", INDENT, val)
			i := strFirst(val, ":")
			callerPath := val[0:i]
			callerPath = strings.TrimPrefix(callerPath, repoDirPrefix)
			elem.CalleeFile = NewPathM(callerPath).Base()

			rem := strRemainder(val, i+1)
			j := strings.Index(rem, "+0x")
			if j >= 0 {
				elem.StackFramePosition = strRemainder(val, j)
				rem = rem[0:j]
			}
			//j := strFirstFrom(val+seek, i, seek)
			elem.CalleeLineNumber = ParseIntM(strings.TrimSpace(rem))
			//elem.StackFramePosition = strRemainder(val, j)
		}
		//j := strLastIndex(val, "(")
		//
		//packageAndMethod := val[0:j]
		//Pr("packageAndMethod:", packageAndMethod)
		//
		//h := strLastIndex(packageAndMethod, "/")
		//
		//packageName := packageAndMethod[0:h]
		//Pr("packageName:", packageName)
		//
		//callerNameAndArgs := strRemainder(packageAndMethod, h+1)
		//Pr("callerNameAndArgs:", callerNameAndArgs)
		//
		//h2 := strLastIndex(callerNameAndArgs, "(")
		//args := strRemainder(callerNameAndArgs, h2)
		//callerName := callerNameAndArgs[0:h2]
		//
		////}
		////h := strings.LastIndex(packageAndMethod,"/)
		////arguments := val[j:valLen]
		//
		//Pr("parsing:", val)
		//Pr("...pkg/method:", packageAndMethod)
		//Pr("...pkg:", packageName)
		//Pr("...callerNameAndArgs:", callerNameAndArgs)
		//Pr("callerName:", callerName)
		//Pr("args:", args)
		//
		////val = strings.TrimSpace(lines[cursor+1])
		////
		////q := strings.LastIndex(val[0:j], ".")
		////quickVerify(q > 0)
		////functionName := val[q+1 : j]
		////Pr("functionName:", functionName)
		//
		//val2 := strings.TrimSpace(lines[cursor+1])
		//cols := strings.Fields(val2)
		//quickVerify(len(cols) == 2)
		//callerPathWithLineNum := cols[0]
		//j2 := strings.Index(callerPathWithLineNum, ":")
		//quickVerify(j2 > 0)
		//callerPath := callerPathWithLineNum[0:j2]
		//callerPath = strings.TrimPrefix(callerPath, repoDirPrefix)
		//callerLineNum := ParseIntM(strRemainder(callerPathWithLineNum, j2+1))
		////callerPath = strings.TrimPrefix(callerPath, repoDirPrefix)
		////xp := NewPathM(callerPath)
		////callerFilename := xp.Base()
		//
		//elem := stackTraceElement{
		//	Package:            packageName,
		//	CallerFunction:     callerName,
		//	CallerArguments:    args,
		//	CalleeFile:         NewPathM(callerPath),
		//	CalleeLineNumber:   callerLineNum,
		//	StackFramePosition: cols[1],
		//}

		Pr("parsed elem:", INDENT, elem)
		if skipped < st.SkipFactor {
			skipped++
		} else {
			rows.Add(ToString(elem))
		}
	}
	//Pr("what is prefix, exactly?")
	//prefix := ""
	//
	//for _, val := range strings.Split(content, "\n") {
	//	result := val
	//	for {
	//		// The first line in the stack trace seems to always be something like 'goroutine 1 [running]:', so
	//		// it doesn't follow the pattern of
	//		//
	//		//     {package name} {name + arguments of caller function}
	//		//     {file where function was called}:{line number within file}  +{relative position of the function within the stack frame}
	//		//
	//		// ...so, store it as a 'preamble'
	//		//
	//		if strings.HasPrefix(val, "goroutine ") {
	//			st.Preamble = val
	//			result = ""
	//			break
	//		}
	//
	//		if strings.HasPrefix(val, "\t") {
	//			val := strings.TrimSpace(val)
	//			cols := strings.Fields(val)
	//			if len(cols) != 2 {
	//				break
	//			}
	//			result = cols[0]
	//			result = strings.TrimPrefix(result, repoDirPrefix)
	//			Pr("trimmed:", Quoted(cols[0]), " prefix:", repoDirPrefix)
	//			Pr("result :", Quoted(result))
	//			xp := NewPathM(result)
	//			result = xp.Base()
	//			break
	//		}
	//
	//	}
	//	if result != "" {
	//		if skipped < st.SkipFactor {
	//			skipped++
	//		} else {
	//			rows.Add(result + ` ` + prefix)
	//		}
	//	}
	//}
	st.Rows = rows.Array()
}

// Wrap a main function so that we can modify the display of any panic stack traces that occur.
func WrapMain(mainFunc func()) {
	Todo("!We may want the option to not exit the program after displaying the panic's stack trace")
	defer func() {
		if r := recover(); r != nil {
			Panic()
		}
	}()
	mainFunc()
}

func CausePanic() int {
	sum := 0
	for i := -3; i < 3; i++ {
		sum += 10 / i
	}
	return sum
}
