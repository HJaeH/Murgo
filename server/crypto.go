package server

import (
	"crypto/rand"
	"io"
)

type CryptState struct {
	key       []byte
	encryptIV []byte
	decryptIV []byte
}

func (c *CryptState) GenerateKey() error {

	key := make([]byte, 100)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return err
	}
	c.key = key

	c.encryptIV = make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, c.encryptIV)
	if err != nil {
		return err
	}

	c.decryptIV = make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, c.decryptIV)
	if err != nil {
		return err
	}
	return nil
}

func (c *CryptState) Key() []byte {
	return c.key
}

func (c *CryptState) EncryptIV() []byte {
	return c.encryptIV
}

func (c *CryptState) DecryptIV() []byte {
	return c.decryptIV
}

func SupportedModes() []string {
	return []string{
		"OCB2-AES128",
		"XSalsa20-Poly1305",
	}
}
