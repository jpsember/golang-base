package jsdata

import (
	"strings"

	//. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

type EnumInfo struct {
	EnumNames []string
	EnumIds   map[string]uint32
}

func NewEnumInfo(enumNames string) *EnumInfo {
	var m = new(EnumInfo)
	m.EnumNames = strings.Split(enumNames, " ")
	m.EnumIds = make(map[string]uint32)
	for id, name := range m.EnumNames {
		var value = uint32(id)
		m.EnumIds[name] = value
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

func (info *EnumInfo) ValueOf(s string) (uint32, error) {
	if v, found := info.EnumIds[s], found {
		return v, nil
	}
	return 0, fmt.Errorf("can't find enum with label %q", s)
}
