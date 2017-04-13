package main

import (
	"io"
	"crypto/rand"
)


//import "io"

type CryptState struct{
	key       []byte
	encryptIV []byte
	decryptIV []byte
}



func (cryptState *CryptState) GenerateKey() error {

	key := make([]byte, 100)// todo : need specified byte size
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return err
	}
	cryptState.key = key

	cryptState.encryptIV = make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, cryptState.encryptIV)
	if err != nil {
		return err
	}

	cryptState.decryptIV = make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, cryptState.decryptIV)
	if err != nil {
		return err
	}

	return nil
}



func SupportedModes() []string {
	return []string{
		"OCB2-AES128",
		"XSalsa20-Poly1305",
	}
}
