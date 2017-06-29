package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var EncryptKey []byte = []byte("test-text-to-code-cookies-hakem!")

func encryptAES(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
func decryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext	too	short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func encrypt(s string) (string, error) {
	aes, err := encryptAES([]byte(s), EncryptKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(aes), nil
}

func decrypt(s string) (string, error) {
	cipher, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	value, err := decryptAES(cipher, EncryptKey)
	if err != nil {
		return "", err
	}
	return string(value), nil
}
