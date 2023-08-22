package base

import (
	"math/rand"
	"strings"
	_ "strings"
)

func RandomText(rand *rand.Rand, maxLength int, withLinefeeds bool) string {

	sample := "orhxxidfusuytelrcfdlordburswfxzjfjllppdsywgswkvukrammvxvsjzqwplxcpkoekiznlgsgjfonlugreiqvtvpjgrqotzu"

	sb := strings.Builder{}
	length := MinInt(maxLength, 2+rand.Intn(maxLength))
	for sb.Len() < length {
		wordSize := rand.Intn(8) + 2
		if withLinefeeds && rand.Intn(4) == 0 {
			sb.WriteByte('\n')
		} else {
			sb.WriteByte(' ')
		}
		c := rand.Intn(len(sample) - wordSize)
		sb.WriteString(sample[c : c+wordSize])
	}
	return strings.TrimSpace(sb.String())
}

func ParseEnumFromMap(enumInfo *EnumInfo, m JSMap, key string, defaultValue int) int {
	var result = defaultValue
	var val = m.OptString(key, "")
	if val != "" {
		if id, found := enumInfo.EnumIds[val]; found {
			result = int(id)
		} else {
			BadArg("No such value for enum:", val)
		}
	}
	return result
}
