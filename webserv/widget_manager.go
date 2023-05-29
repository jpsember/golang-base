package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

var _ = Pr

type WidgetObj struct {
	Id string
}

type Widget = *WidgetObj

type WidgetManagerObj struct {
	pendingColumnWeights []int
}

type WidgetManager = *WidgetManagerObj

var digitsExpr = Regexp(`^\d+$`)

/**
 * <pre>
 *
 * Set the number of columns, and which ones can grow, for the next view in
 * the hierarchy. The columns expression is a string of column expressions,
 * which may be one of:
 *
 *     "."   a column with weight zero
 *     "x"   a column with weight 100
 *     "\d+" column with integer weight
 *
 * Spaces are ignored, except to separate integer weights from each other.
 * </pre>
 */
func (m WidgetManager) Columns(columnsExpr string) WidgetManager {
	CheckState(m.pendingColumnWeights == nil, "previous column weights were never used")

	columnSizes := NewArray[int]()
	for _, word := range strings.Split(columnsExpr, " ") {
		size := 0
		if digitsExpr.MatchString(word) {
			size = ParseIntM(word)
			columnSizes.Add(size)
		} else {
			w := []byte(word)
			for _, c := range w {
				if c == '.' {
					size = 0
				} else if c == 'x' {
					size = 100
				} else {
					BadArg("Can't parse columns expression:", Quoted(columnsExpr))
				}
				columnSizes.Add(size)
			}
		}
	}
	m.pendingColumnWeights = columnSizes.Array()
	return m
}
