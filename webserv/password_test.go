package webserv_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	. "github.com/jpsember/golang-base/webserv"
	"testing"
)

func TestHashPassword(t *testing.T) {
	j := jt.New(t)

	pwd := "extemperaneous"
	hash, salt := HashPassword(pwd)
	j.Log("hash:", JSListWith(hash))
	j.Log("salt:", salt)

	hash2 := HashPasswordWithSalt(pwd, salt)
	j.AssertEqual(hash, hash2)
}

func TestVerifyPasswordWithSalt(t *testing.T) {
	j := jt.New(t)

	hashExpected := []byte{144, 158, 214, 211, 161, 15, 172, 211, 79, 172, 228,
		72, 221, 213, 234, 107, 245, 233, 171, 71, 125, 132, 135, 221, 69, 73, 189, 66, 1, 69, 33, 232}

	salt := 1_234_567
	result := VerifyPassword(salt, hashExpected, "extemperaneous")
	j.AssertTrue(result)
}

func TestVariousPasswords(t *testing.T) {
	j := jt.New(t)

	r := NewJSRand().SetSeed(1965)

	mp := NewJSMap()

	for mp.Size() < 50 {
		word := RandomText(r, 24, false)
		if len(word) < 9 || len(word) > 23 {
			continue
		}
		hash, salt := HashPassword(word)
		m2 := NewJSMap().Put("hash", JSListWith(hash)).Put("salt", salt)
		mp.Put(word, m2)

		CheckState(VerifyPassword(salt, hash, word))
		CheckState(!VerifyPassword(salt, hash, word+"z"))
		CheckState(!VerifyPassword(salt, hash, word[1:]))
	}
	j.Log(mp)
}
