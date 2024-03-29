package webserv

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	. "github.com/jpsember/golang-base/base"
	"time"
)

func HashPassword(password string) ([]byte, int) {
	active := false
	pr := PrIf("HashPassword", active)
	salt := int(time.Now().UnixMilli()) * 27644437
	pr("salt:", salt, salt%1000)

	hash := HashPasswordWithSalt(password, salt)
	if active {
		Pr("Hash:", INDENT, JSListWith(hash))
	}
	return hash, salt
}

func isPwdLegalLength(pwdBytes []byte) bool {
	x := len(pwdBytes)
	return x >= USER_PASSWORD_MIN_LENGTH && x <= USER_PASSWORD_MAX_LENGTH

}

const USER_PASSWORD_MIN_LENGTH = 8
const USER_PASSWORD_MAX_LENGTH = 20

const saltLength = 8
const chunkLength = 32

func HashPasswordWithSalt(password string, salt int) []byte {
	active := true
	pr := PrIf("HashPassword", false)

	pwdBytes := []byte(password)
	x := len(pwdBytes)
	CheckArg(isPwdLegalLength(pwdBytes), "password length", QUO, password)

	// Create a buffer to hold the salt and password
	buffer := make([]byte, chunkLength, chunkLength)

	// Store salt in the first 8 bytes
	binary.LittleEndian.PutUint64(buffer, uint64(salt))
	pr("added salt:", INDENT, HexDumpWithASCII(buffer))

	copy(buffer[saltLength:saltLength+x], pwdBytes)
	pr("added pwd:", INDENT, HexDumpWithASCII(buffer))

	CheckState(len(buffer) == chunkLength)

	h := sha256.New()
	h.Write(buffer)
	hashedResult := h.Sum(nil)
	if active {
		pr("SHA256 hash:", INDENT, JSListWith(hashedResult))
	}
	return hashedResult
}

func VerifyPassword(salt int, validHash []byte, password string) bool {
	active := false
	pr := PrIf("VerifyPassword", active)

	if !isPwdLegalLength([]byte(password)) {
		return false
	}
	calcHash := HashPasswordWithSalt(password, salt)

	if active {
		pr("password:", password)
		pr("hash:", INDENT, JSListWith(validHash))
		pr("calc:", INDENT, JSListWith(calcHash))
	}

	return bytes.Equal(validHash, calcHash)
}
