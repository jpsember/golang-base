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
