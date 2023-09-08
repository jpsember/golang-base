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

var AllowTestInputs = Alert("!Allowing test inputs (user name, password, etc)")

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

type ValidateFlag int

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
var ErrorUserNameIllegalCharacters = Error("Your name has illegal characters")

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

var UserNameValidatorRegExp = Regexp(`^[a-zA-Z0-9_]+(?: [a-zA-Z0-9_]+)*$`)
var UserPasswordValidatorRegExp = Regexp(`^[^ ]+$`)

func ValidateUserName(userName string, flag ValidateFlag) (string, error) {
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
	err, validatedName = replaceWithTestInput(err, validatedName, "a", "joeuser42")
	return validatedName, err
}

func ValidateUserPassword(password string, flag ValidateFlag) (string, error) {
	pr := PrIf(false)
	pr("ValidateUserPassword:", Quoted(password), flag)

	text := password
	text = strings.TrimSpace(text)
	validatedName := text
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

	pr("before replaceWithTestInput:", err, validatedName)
	err, validatedName = replaceWithTestInput(err, validatedName, "a", "bigpassword123")
	pr("after replaceWithTestInput:", err, validatedName)
	return validatedName, err
}

func ValidateEmailAddress(emailAddress string, flag ValidateFlag) (string, error) {
	text := emailAddress
	text = strings.TrimSpace(text)
	validatedEmail := text
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
			_, err2 := mail.ParseAddress(text)
			if err2 != nil {
				err = ErrorUserEmailInvalid
			}
		}
	}

	err, validatedEmail = replaceWithTestInput(err, validatedEmail, "a", "joe_user@anycompany.xyx")
	return validatedEmail, err
}

// If AllowTestInputs, and value is empty or equals shortcutForTest, return (nil, fullValueForTest).
func replaceWithTestInput(err error, value string, shortcutForTest string, fullValueForTest string) (error, string) {
	if AllowTestInputs {
		if value == shortcutForTest || value == "" {
			value = fullValueForTest
			Alert("?<2replaceWithTestInput; replacing: " + shortcutForTest + " with: " + value)
			err = nil
		}
	}
	return err, value
}

var contentTypeNames = []string{
	"image/jpeg", //
	"image/png",  //
}

var contentTypeSignatures = []byte{
	3, 0xff, 0xd8, 0xff, // jpeg
	8, 137, 80, 78, 71, 13, 10, 26, 10, // png
}

func InferContentTypeFromBlob(blob Blob) string {
	result := ""
	data := blob.Data()

	sig := contentTypeSignatures

	i := 0
	for fn := 0; i < len(sig); fn++ {
		recSize := int(sig[i])
		j := i + 1
		i += recSize + 1
		match := true
		for k := 0; k < recSize; k++ {
			if sig[j+k] != data[k] {
				match = false
				break
			}
		}
		if match {
			result = contentTypeNames[fn]
			break
		}
	}

	if result == "" {
		Alert("#50<1Failed to determine content-type from bytes:", CR,
			HexDumpWithASCII(ByteSlice(blob.Data(), 0, 16)))
		result = "application/octet-stream"
	}
	return result
}
