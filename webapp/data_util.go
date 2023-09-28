package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"net/mail"
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
	pr := PrIf("", false)
	pr("currency to string, amount:", amount)
	j := IntToString(amount)
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

func GenerateBlobName() BlobId {
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

type ValidateFlag int

func (x ValidateFlag) String() string {
	return BinaryN(int(x))
}

const (
	VALIDATE_EMPTYOK       ValidateFlag = 1 << iota // A blank value is ok
	VALIDATE_ONLY_NONEMPTY                          // Check only that the value isn't blank
)

func (f ValidateFlag) Has(bits ValidateFlag) bool {
	return (f & bits) == bits
}

const USER_NAME_MAX_LENGTH = 20

var ErrorEmptyUserName = Error("Please enter your name")
var ErrorUserNameTooLong = Error("Your name is too long")
var ErrorUserNameIllegalCharacters = Error("Name has illegal characters")
var ErrorEmptyAnimalName = Error("Please enter a name")

var ErrorEmptyUserPassword = Error("Please enter a password")
var ErrorEmptyUserEmail = Error("Please enter an email address")
var ErrorUserEmailInvalid = Error("The email address is invalid")

var ErrorUserPasswordLength = Error("Password must be between 8 and 20 characters")
var ErrorUserPasswordIllegalCharacters = Error("Password must not contain spaces")
var ErrorUserPasswordsDontMatch = Error("The two passwords don't match")
var ErrorUserEmailLength = Error("Email address is too long")

const USER_PASSWORD_MIN_LENGTH = 8
const USER_PASSWORD_MAX_LENGTH = 20
const USER_EMAIL_MAX_LENGTH = 40

const ANIMAL_NAME_MAX_LENGTH = 16
const ANIMAL_NAME_MIN_LENGTH = 2

var ErrorAnimalNameTooShort = Error("The name is too short")
var ErrorAnimalNameTooLong = Error("The name is too long")
var ErrorAnimalNameIllegalCharacters = Error("Your name has illegal characters")

var UserNameValidatorRegExp = Regexp(`^[a-zA-Z0-9_]+(?: [a-zA-Z0-9_]+)*$`)
var UserPasswordValidatorRegExp = Regexp(`^[^ ]+$`)
var AnimalNameValidatorRegExp = Regexp(`^[a-zA-Z]+(?: [a-zA-Z]+)*$`)

func ValidateAnimalName(name string, flag ValidateFlag) (string, error) {
	pr := PrIf("ValidateAnimalName", false)
	pr("name:", QUO, name, "flag:", flag)
	name = strings.TrimSpace(name)
	Todo("?Replace two or more spaces by a single space")
	validatedName := name
	var err error

	if name == "" {
		if !flag.Has(VALIDATE_EMPTYOK) {
			err = ErrorEmptyAnimalName
		}
	} else if len(name) > ANIMAL_NAME_MAX_LENGTH {
		err = ErrorAnimalNameTooLong
	} else if len(name) < ANIMAL_NAME_MIN_LENGTH {
		err = ErrorAnimalNameTooShort
	} else if !AnimalNameValidatorRegExp.MatchString(name) {
		err = ErrorAnimalNameIllegalCharacters
	}
	pr("returning", QUO, validatedName, "error:", err)
	return validatedName, err
}

func ValidateUserName(userName string, flag ValidateFlag) (string, error) {
	pr := PrIf("ValidateUserName", false)
	userName = strings.TrimSpace(userName)
	Todo("?Replace two or more spaces by a single space")
	validatedName := userName
	var err error

	if userName == "" {
		if !flag.Has(VALIDATE_EMPTYOK) {
			err = ErrorEmptyUserName
		}
	} else if len(userName) > USER_NAME_MAX_LENGTH {
		err = ErrorUserNameTooLong
	} else if !UserNameValidatorRegExp.MatchString(userName) {
		err = ErrorUserNameIllegalCharacters
	}
	pr("userName:", QUO, userName, "flags:", flag, "result:", QUO, validatedName, "err:", err)
	return validatedName, err
}

func ValidateUserPassword(password string, flag ValidateFlag) (string, error) {
	pr := PrIf(">ValidateUserPassword", false)
	pr("pwd:", QUO, password, flag)

	text := password
	text = strings.TrimSpace(text)
	var err error

	x := len(text)
	if x == 0 {
		if !flag.Has(VALIDATE_EMPTYOK) {
			err = ErrorEmptyUserPassword
		}
	} else if !flag.Has(VALIDATE_ONLY_NONEMPTY) {
		if x < USER_PASSWORD_MIN_LENGTH || x > USER_PASSWORD_MAX_LENGTH {
			err = ErrorUserPasswordLength
		} else if !UserPasswordValidatorRegExp.MatchString(text) {
			err = ErrorUserPasswordIllegalCharacters
		}
	}

	pr("before replaceWithTestInput:", err, text)
	err, text = replaceWithTestInput(err, text, "a", "bigpassword123")
	pr("after replaceWithTestInput:", err, text)
	return text, err
}

func ValidateEmailAddress(emailAddress string, flag ValidateFlag) (string, error) {
	pr := PrIf(">ValidateEmailAddress", false)
	pr("email:", QUO, emailAddress, flag)

	text := emailAddress
	text = strings.TrimSpace(text)
	var err error

	x := len(text)
	if x == 0 {
		if !flag.Has(VALIDATE_EMPTYOK) {
			err = ErrorEmptyUserEmail
		}
	} else if !flag.Has(VALIDATE_ONLY_NONEMPTY) {
		if x > USER_EMAIL_MAX_LENGTH {
			err = ErrorUserEmailLength
		} else {
			_, err = mail.ParseAddress(text)
			if err != nil {
				err = ErrorUserEmailInvalid
			}
		}
	}
	err, text = replaceWithTestInput(err, text, "a", "joe_user@anycompany.xyx")
	pr("returning:", QUO, text, err)
	return text, err
}

// If AllowTestInputs, and value is empty or equals shortcutForTest, return (nil, fullValueForTest).
func replaceWithTestInput(err error, value string, shortcutForTest string, fullValueForTest string) (error, string) {
	if AllowTestInputs {
		if value == shortcutForTest || value == "" {
			value = fullValueForTest
			Alert("?<2replaceWithTestInput;    replacing: " + shortcutForTest + " with: " + value)
			err = nil
		}
	}
	return err, value
}

func RandomEmailAddress(r JSRand) string {
	r = NullToRand(r)
	return RandomWord(r) + "@" + RandomWord(r) + ".net"
}

func BlobSummary(blob Blob) JSMap {
	b := blob.Build().ToBuilder()
	b.SetData(nil)
	r := b.ToJson().AsJSMap()
	r.Put("data", HexDumpWithASCII(ByteSlice(blob.Data(), 0, 16)))
	return r
}

var AnimalPicSizeNormal = IPointWith(600, 800)
