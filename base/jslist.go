package base

type JSListStruct struct {
	wrappedList []JSEntity
}
type JSList = *JSListStruct

func JSListWith[T any](values []T) JSList {
	var out = NewJSList()
	for _, x := range values {
		out.Add(x)
	}
	return out
}

// ---------------------------------------------------------------------------------------
// JSEntity interface
// ---------------------------------------------------------------------------------------

func (js JSList) AsInteger() int64 {
	panic("not supported")
}

func (js JSList) AsFloat() float64 {
	panic("not supported")
}

func (js JSList) AsString() string {
	panic("not supported")
}

func (js JSList) AsBool() bool {
	panic("not supported")
}

func (js JSList) AsJSMap() JSMap {
	panic("not supported")
}

func (js JSList) AsJSList() JSList {
	return js
}

// Implements the fmt.Stringer interface.  By default, we perform
// a pretty print of the JSListStruct.  This simplifies a lot of things.
func (js JSList) String() string {
	return PrintJSEntity(js, true)
}

// Convert JSListStruct to string, without pretty printing.
func (js JSList) CompactString() string {
	return PrintJSEntity(js, false)
}

func (js JSList) PrintTo(context *JSONPrinter) {
	var s = context.StringBuilder
	s.WriteByte('[')
	var index = 0
	for _, val := range js.wrappedList {
		if index != 0 {
			s.WriteByte(',')
		}
		index++
		val.PrintTo(context)
	}
	s.WriteByte(']')
}

// ---------------------------------------------------------------------------------------

// Factory constructor.  Do *not* construct via JSListStruct().
func NewJSList() JSList {
	var m = new(JSListStruct)
	m.wrappedList = make([]JSEntity, 0, 10)
	return m
}

func (p *JSONParser) ParseList() (JSList, error) {
	p.adjustNest(1)
	var result []JSEntity
	p.ReadExpectedByte('[')
	commaExpected := false
	for !p.hasProblem() {
		if p.readIf(']') {
			break
		}
		if commaExpected {
			p.ReadExpectedByte(',')
			if p.readIf(']') {
				break
			}
			commaExpected = false
		}
		elem := p.readValue()
		if p.hasProblem() {
			break
		}
		result = append(result, elem)
		commaExpected = true
	}
	p.skipWhitespace()
	p.adjustNest(-1)
	var jsList JSList
	if p.Error == nil {
		jsList = new(JSListStruct)
		jsList.wrappedList = result
	}
	return jsList, p.Error
}

func (js JSList) Add(value any) JSList {
	if value == nil {
		BadArg("value is nil")
	}
	js.wrappedList = append(js.wrappedList, ToJSEntity(value))
	return js
}

func (js JSList) Get(index int) JSEntity {
	return js.wrappedList[index]
}

func (js JSList) AsMaps() []JSMap {
	var x []JSMap
	for _, y := range js.wrappedList {
		x = append(x, y.AsJSMap())
	}
	return x
}

func (js JSList) Length() int {
	return len(js.wrappedList)
}

func (js JSList) AsByteArray() []byte {
	w := js.wrappedList
	res := make([]byte, len(w))
	for i, v := range w {
		res[i] = byte(v.AsInteger())
	}
	return res
}
