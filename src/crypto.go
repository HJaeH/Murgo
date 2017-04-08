package main

//import "io"

type CryptState struct{
	key       []byte
	encryptIV []byte
	decryptIV []byte
}


/*

func (cryptState *CryptState) GenerateKey(mode string) error {

	key := make([]byte, cm.KeySize())
	_, err = io.ReadFull(rand.Reader, key)
	if err != nil {
		return err
	}

	cm.SetKey(key)
	cs.mode = cm
	cs.Key = key

	cs.EncryptIV = make([]byte, cm.NonceSize())
	_, err = io.ReadFull(rand.Reader, cs.EncryptIV)
	if err != nil {
		return err
	}

	cs.DecryptIV = make([]byte, cm.NonceSize())
	_, err = io.ReadFull(rand.Reader, cs.DecryptIV)
	if err != nil {
		return err
	}

	return nil
}


*/

func SupportedModes() []string {
	return []string{
		"OCB2-AES128",
		"XSalsa20-Poly1305",
	}
}
