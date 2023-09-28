package webserv

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	. "github.com/jpsember/golang-base/base"
	"strings"
	"time"
)

func HashPassword(password string) ([]byte, int) {
	active := true
	pr := PrIf("HashPassword", active)
	salt := int(time.Now().UnixMilli()) * 27644437
	pr("salt:", salt, salt%1000)

	hash := HashPasswordWithSalt(password, salt)
	if active {
		Pr("Hash:", INDENT, JSListWith(hash))
	}
	return hash, salt
}

func HashPasswordWithSalt(password string, salt int) []byte {
	active := true
	pr := PrIf("HashPassword", active)
	const saltLength = 8
	const chunkLength = 32
	const maxPwdLength = chunkLength - saltLength

	pwdBytes := []byte(password)

	x := len(pwdBytes)
	CheckArg(x >= 8 && x <= maxPwdLength, "password length", QUO, password)

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
		pr("SHA256 hash:", INDENT, BytesAsSourceArray(hashedResult))
	}
	return hashedResult
}

func VerifyPassword(salt int, validHash []byte, password string) bool {
	active := true
	pr := PrIf("VerifyPassword", active)
	pr("password:", password)
	pr("hash:", INDENT, HexDump(validHash))

	calcHash := HashPasswordWithSalt(password, salt)
	pr("calc:", INDENT, calcHash)

	pr("type of validHash:", Info(validHash))
	pr("type of calcdHash:", Info(calcHash))
	return bytes.Equal(validHash, calcHash)
}

func BytesAsSourceArray(bytes []byte) string {
	s := strings.Builder{}
	s.WriteByte('[')
	for i, x := range bytes {
		if i != 0 {
			s.WriteByte(',')
		}
		s.WriteString(IntToString(int(x)))
	}
	s.WriteByte(']')
	return s.String()
}
