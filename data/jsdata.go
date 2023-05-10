package jsdata

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// Implementation of fmt.stringer for DataClass
// TODO: If we have a more specific implementation, will that take priority?
// TODO: Should this be moved to the datagen package?
func String(obj DataClass) string {
	var x = obj.ToJson()
	var js = x.(JSEntity)
	return PrintJSEntity(js, true)
}
