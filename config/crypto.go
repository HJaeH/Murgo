package config

import (
	"crypto/rand"
	"io"
)

type CryptState struct {
	key       []byte
	encryptIV []byte
	decryptIV []byte

	LastGoodTime int64

	Good         uint32
	Late         uint32
	Lost         uint32
	Resync       uint32
	RemoteGood   uint32
	RemoteLate   uint32
	RemoteLost   uint32
	RemoteResync uint32
}

func (cryptState *CryptState) GenerateKey() error {

	key := make([]byte, 100) // todo : need a specified byte size
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

func (cryptState *CryptState) Key() []byte {
	return cryptState.key
}

func (cryptState *CryptState) EncryptIV() []byte {
	return cryptState.encryptIV
}

func (cryptState *CryptState) DecryptIV() []byte {
	return cryptState.decryptIV
}

func SupportedModes() []string {
	return []string{
		"OCB2-AES128",
		"XSalsa20-Poly1305",
	}
}
