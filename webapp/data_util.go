package webapp

import (
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"strings"
	"sync"
)

type BlobId string

const blobIdLength = 10

func (b BlobId) String() string {
	return string(b)
}

func StringToBlobId(s string) BlobId {
	if len(s) != blobIdLength {
		BadArg("<1Not a legal blob id:", Quoted(s))
	}
	return BlobId(s)
}

type Currency = int

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

	for i := 0; i < blobIdLength; i++ {
		x := r.Intn(16)
		sb.WriteByte(alph[x])
	}
	return StringToBlobId(sb.String())
}

var ourRand = NewJSRand()
var lock sync.Mutex

func PerformBlobExperiment(db Database) {
	data := []byte{
		2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41,
	}
	for i := 0; i < 10; i++ {
		Pr("inserting:", i)
		bl, err := db.InsertBlob(data)
		Pr("result:", INDENT, bl, CR, err)

		Pr("verifying:")
		blob, err := db.ReadBlob(StringToBlobId(bl.Id()))
		CheckState(bytes.Equal(data, blob.Data()))
	}
}

const USER_NAME_MAX_LENGTH = 20

// var ErrorNone = Error("")
var ErrorEmptyUserName = Error("Please enter your name")
var ErrorUserNameTooLong = Error("Your name is too long")
var ErrorUserNameIllegalCharacters = Error("Your name has illegal characters")

var ErrorEmptyUserPassword = Error("Please enter a password")
var ErrorUserPasswordLength = Error("Password must be between 8 and 20 characters")
var ErrorUserPasswordIllegalCharacters = Error("Password must not contain spaces")
var ErrorUserPasswordsDontMatch = Error("The two passwords don't match")

const USER_PASSWORD_MIN_LENGTH = 8
const USER_PASSWORD_MAX_LENGTH = 20

var UserNameValidatorRegExp = Regexp(`^[a-zA-Z0-9_]+(?: [a-zA-Z0-9_]+)*$`)
var UserPasswordValidatorRegExp = Regexp(`^[^ ]+$`)

func ValidateUserName(userName string, emptyOk bool) (string, error) {
	userName = strings.TrimSpace(userName)
	Todo("?Replace two or more spaces by a single space")
	validatedName := userName
	var err error

	if userName == "" {
		if !emptyOk {
			err = ErrorEmptyUserName
		}
	} else if len(userName) > USER_NAME_MAX_LENGTH {
		err = ErrorUserNameTooLong
	} else if !UserNameValidatorRegExp.MatchString(userName) {
		err = ErrorUserNameIllegalCharacters
	}
	return validatedName, err
}

func ValidateUserPassword(password string, emptyOk bool) (string, error) {
	password = strings.TrimSpace(password)
	validatedName := password
	var err error

	x := len(password)
	if x == 0 {
		if !emptyOk {
			err = ErrorEmptyUserPassword
		}
	} else if x < USER_PASSWORD_MIN_LENGTH || x > USER_PASSWORD_MAX_LENGTH {
		err = ErrorUserPasswordLength
	} else if !UserPasswordValidatorRegExp.MatchString(password) {
		err = ErrorUserPasswordIllegalCharacters
	}
	return validatedName, err
}
