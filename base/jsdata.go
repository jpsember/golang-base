package base

import (
	"math/rand"
	"strings"
	_ "strings"
)

func RandomText(j JSRand, maxLength int, withLinefeeds bool) string {
	sb := strings.Builder{}
	length := MinInt(maxLength, 2+rand.Intn(maxLength))
	for sb.Len() < length {
		if withLinefeeds && rand.Intn(4) == 0 {
			sb.WriteByte('\n')
		} else {
			sb.WriteByte(' ')
		}
		sb.WriteString(RandomWord(j))
	}
	return strings.TrimSpace(sb.String())
}

func RandomWord(j JSRand) string {
	sample := "orhxxidfusuytelrcfdlordburswfxzjfjllppdsywgswkvukrammvxvsjzqwplxcpkoekiznlgsgjfonlugreiqvtvpjgrqotzu"
	wordSize := rand.Intn(8) + 2
	c := rand.Intn(len(sample) - wordSize)
	return sample[c : c+wordSize]
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

func ParseOrDefault(json JSEntity, defaultValue DataClass) (DataClass, error) {
	var err error
	result := attemptParse(json, defaultValue)
	if result == nil {
		err = DataClassParseError
	}
	return result, err
}

var DataClassParseError = Error("DataClass parse error")

func attemptParse(json JSEntity, parser DataClass) DataClass {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	return parser.Parse(json)
}
