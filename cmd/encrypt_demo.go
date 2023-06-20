package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
)

// adapted from https://medium.com/insiderengineering/aes-encryption-and-decryption-in-golang-php-and-both-with-full-codes-ceb598a34f41

type EncryptOper struct {
	BaseObject
}

func (oper *EncryptOper) UserCommand() string {
	return "encrypt"
}

func (oper *EncryptOper) Perform(app *App) {
	oper.SetVerbose(true)
	pr := oper.Log
	pr("this is EncryptOper.perform")
	exp()
}

func (oper *EncryptOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Performs AES encryption/decryption.")
}

func (oper *EncryptOper) ProcessArgs(c *CmdLineArgs) {
	for c.HasNextArg() {
		var arg = c.NextArg()
		switch arg {
		default:
			c.SetError("extraneous argument:", arg)
		}
	}
}

func main() {
	var oper = &EncryptOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("encrypt_demo")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests")
	app.AddTestArgs("--verbose")
	app.Start()
}

func stringToBytes(str string) []byte {
	return []byte(str)
}

func exp() {
	plainText := "Hello, World!"

	Pr("Message  :", plainText)
	bytes := stringToBytes(plainText)
	Pr("Input    :", bytes)

	encrypted, err := GetAESEncryptedBytes(bytes)
	CheckOk(err)

	Pr("Encrypted:", encrypted)

	decrypted, err := GetAESDecryptedBytes(encrypted)
	CheckOk(err)

	Pr("Decrypted:", decrypted)
	Pr("Message  :", string(decrypted))
}

// GetAESDecrypted decrypts given text in AES 256 CBC
func GetAESDecryptedBytes(ciphertext []byte) ([]byte, error) {
	key := "my32digitkey12345678901234567890"
	iv := "my16digitIvKey12"

	block, err := aes.NewCipher([]byte(key))

	if err != nil {
		return nil, err
	}

	CheckArg(len(ciphertext)%aes.BlockSize == 0)

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, ciphertext)
	ciphertext = PKCS5UnPadding(ciphertext)

	return ciphertext, nil
}

// PKCS5UnPadding  pads a certain blob of data with necessary data to be used in AES block cipher
func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])

	return src[:(length - unpadding)]
}

func GetAESEncryptedBytes(sourceBytes []byte) ([]byte, error) {
	key := "my32digitkey12345678901234567890"
	iv := "my16digitIvKey12"

	var plainTextBlock []byte
	length := len(sourceBytes)

	if length%16 != 0 {
		extendBlock := 16 - (length % 16)
		plainTextBlock = make([]byte, length+extendBlock)
		copy(plainTextBlock[length:], bytes.Repeat([]byte{uint8(extendBlock)}, extendBlock))
	} else {
		plainTextBlock = make([]byte, length)
	}

	copy(plainTextBlock, sourceBytes)
	block, err := aes.NewCipher([]byte(key))
	CheckOk(err)

	ciphertext := make([]byte, len(plainTextBlock))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainTextBlock)

	Pr("encrypted:", ciphertext)
	return ciphertext, nil
}
