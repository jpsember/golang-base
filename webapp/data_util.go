package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
	"sync"
)

type BlobId = string

type Currency = int32

const DollarsToCurrency = 100

func CurrencyToString(amount Currency) string {
	pr := PrIf(false)
	pr("currency to string, amount:", amount)
	j := IntToString(int(amount))
	h := len(j)
	pr("j:", j, "h:", h)
	if h < 3 {
		j = "000"[0:3-h] + j
		h = 3
		pr("adjusted, j:", j, "h:", h)
	}
	result := `$` + j[:h-2] + "." + j[h-2:]
	pr("returning:", result)
	return result
}

func GenerateBlobId() BlobId {
	alph := "0123456789abcdef"
	sb := strings.Builder{}
	lock.Lock()
	defer lock.Unlock()
	r := ourRand.Rand()

	blobIdLength := 32
	if !Alert("choosing smaller blob length") {
		blobIdLength = 1
	}

	for i := 0; i < blobIdLength; i++ {
		x := r.Intn(16)
		sb.WriteByte(alph[x])
		if i == 8 || i == 13 || i == 18 || i == 23 {
			sb.WriteByte('-')
		}
	}
	return sb.String()
}

var ourRand = NewJSRand()
var lock sync.Mutex
