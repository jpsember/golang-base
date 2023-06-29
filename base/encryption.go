package base

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
)

const AlgorithmNonceSize int = 12
const AlgorithmKeySize int = 16
const PBKDF2SaltSize int = 16
const PBKDF2Iterations int = 32767

func EncryptBytes(bytes []byte, password string) ([]byte, error) {
	// Generate a 128-bit salt using a CSPRNG.
	salt := make([]byte, PBKDF2SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return []byte{}, err
	}

	// Derive a key using PBKDF2.
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AlgorithmKeySize, sha256.New)

	// encryptHelper and prepend salt.
	ciphertextAndNonce, err := encryptHelper(bytes, key)
	if err != nil {
		return []byte{}, err
	}

	ciphertextAndNonceAndSalt := make([]byte, 0)
	ciphertextAndNonceAndSalt = append(ciphertextAndNonceAndSalt, salt...)
	ciphertextAndNonceAndSalt = append(ciphertextAndNonceAndSalt, ciphertextAndNonce...)

	return ciphertextAndNonceAndSalt, nil
}

func encryptHelper(plaintext, key []byte) ([]byte, error) {
	// Generate a 96-bit nonce using a CSPRNG.
	nonce := make([]byte, AlgorithmNonceSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// Create the cipher and block.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// encryptHelper and prepend nonce.
	ciphertext := cipher.Seal(nil, nonce, plaintext, nil)
	ciphertextAndNonce := make([]byte, 0)

	ciphertextAndNonce = append(ciphertextAndNonce, nonce...)
	ciphertextAndNonce = append(ciphertextAndNonce, ciphertext...)

	return ciphertextAndNonce, nil
}

func DecryptBytes(ciphertextAndNonceAndSalt []byte, password string) ([]byte, error) {

	// Create slices pointing to the salt and ciphertextAndNonce.
	salt := ciphertextAndNonceAndSalt[:PBKDF2SaltSize]
	ciphertextAndNonce := ciphertextAndNonceAndSalt[PBKDF2SaltSize:]

	// Derive the key using PBKDF2.
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AlgorithmKeySize, sha256.New)

	// decryptHelper and return result.
	plaintext, err := decryptHelper(ciphertextAndNonce, key)
	if err != nil {
		return []byte{}, err
	}

	return plaintext, nil
}

func decryptHelper(ciphertextAndNonce, key []byte) ([]byte, error) {
	// Create slices pointing to the ciphertext and nonce.
	nonce := ciphertextAndNonce[:AlgorithmNonceSize]
	ciphertext := ciphertextAndNonce[AlgorithmNonceSize:]

	// Create the cipher and block.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// decryptHelper and return result.
	plaintext, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
