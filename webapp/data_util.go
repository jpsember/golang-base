package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

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
