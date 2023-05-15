package jsdata

import (
	"strings"

	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// Implementation of fmt.stringer for DataClass
// TODO: If we have a more specific implementation, will that take priority?
func String(obj DataClass) string {
	var x = obj.ToJson()
	var js = x.(JSEntity)
	return PrintJSEntity(js, true)
}

type EnumInfo struct {
	EnumNames []string
	EnumIds   map[string]uint32
}

func NewEnumInfo(enumNames string) *EnumInfo {
	var m = new(EnumInfo)
	m.EnumNames = strings.Split(enumNames, " ")
	m.EnumIds = make(map[string]uint32)
	for id, name := range m.EnumNames {
		m.EnumIds[name] = uint32(id)
	}
	return m
}

func (info *EnumInfo) String() string {
	var m = NewJSMap()
	m.Put("", "EnumInfo")
	m.Put("names", JSListWithStrings(info.EnumNames))

	var m2 = NewJSMap()
	for k, v := range info.EnumIds {
		m2.Put(k, v)
	}
	m.Put("ids", m2)
	return m.String()
}
