package data

import (
	"encoding/base64"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/json"
	"math/rand"
	"strings"
	_ "strings"
)

func RandomText(rand *rand.Rand, maxLength int, withLinefeeds bool) string {

	sample := "orhxxidfusuytelrcfdlordburswfxzjfjllppdsywgswkvukrammvxvsjzqwplxcpkoekiznlgsgjfonlugreiqvtvpjgrqotzu"

	sb := strings.Builder{}
	length := MinInt(maxLength, rand.Intn(maxLength+2))
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

const DATA_TYPE_DELIMITER = "`"
const DATA_TYPE_SUFFIX_BYTE = DATA_TYPE_DELIMITER + "b"

//const DATA_TYPE_SUFFIX_SHORT = DATA_TYPE_DELIMITER + "s"
//const DATA_TYPE_SUFFIX_INT = DATA_TYPE_DELIMITER + "i"
//const DATA_TYPE_SUFFIX_LONG = DATA_TYPE_DELIMITER + "l"
//const DATA_TYPE_SUFFIX_FLOAT = DATA_TYPE_DELIMITER + "f"
//const DATA_TYPE_SUFFIX_DOUBLE = DATA_TYPE_DELIMITER + "d"

func removeDataTypeSuffix(s string, optionalSuffix string) string {
	if len(s) >= 2 {
		suffixStart := len(s) - 2
		if s[suffixStart] == '`' {
			existingSuffixChar := s[suffixStart+1]
			if existingSuffixChar != optionalSuffix[1] {
				BadArg("string has suffix", existingSuffixChar, "expected", optionalSuffix[1])
			}
			s = s[0:suffixStart]
		}
	}
	return s
}

// Encode a byte array as a Base64 string, with our data type suffix added
func EncodeBase64(byteArray []byte) string {
	return base64.URLEncoding.EncodeToString(byteArray) + DATA_TYPE_SUFFIX_BYTE
}

// Encode a byte array as a Base64 string if it is fairly long
func EncodeBase64Maybe(byteArray []byte) json.JSEntity {
	if len(byteArray) > 8 {
		return json.JString(EncodeBase64(byteArray))
	}
	return json.JSListWith(byteArray)
}

func ParseBase64(s string) []byte {
	s = removeDataTypeSuffix(s, DATA_TYPE_SUFFIX_BYTE)
	result, err := base64.StdEncoding.DecodeString(s)
	CheckOk(err)
	return result
}

/**
 * Parse an array of bytes from a value that is either a JSList, or a base64
 * string. This is so we are prepared to read it whether or not it has been
 * stored in a space-saving base64 form.
 */
func DecodeBase64Maybe(ent json.JSEntity) []byte {
	if arr, ok := ent.(json.JString); ok {
		return ParseBase64(arr.AsString())
	}
	if arr, ok := ent.(json.JSList); ok {
		return arr.AsByteArray()
	}
	BadArg("unexpected type:", Info(ent))
	return nil
}
