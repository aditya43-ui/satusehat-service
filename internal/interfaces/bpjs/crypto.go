package bpjs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// DecryptPayload mengembalikan payload BPJS API yang sudah didekripsi dan disalurkan ke mesin dekompresi otomatis.
func DecryptPayload(cipherText string, consID, secretKey, timestamp string) (string, error) {
	if cipherText == "" {
		return "", errors.New("ciphertext cannot be empty")
	}

	// 1. Generate Key dari gabungan ConsID + SecretKey + Timestamp
	keyString := consID + secretKey + timestamp
	hash := sha256.Sum256([]byte(keyString))
	key := hash[:]

	// IV adalah 16 byte pertama dari key
	iv := key[:16]

	// 2. Base64 Decode CipherText
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", errors.New("failed to base64 decode cipher text")
	}

	// 3. AES-256-CBC Decryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherBytes)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	// Unpad PKCS7
	cipherBytes = unpadPKCS7(cipherBytes)

	// 4. Automatis Dekompresi menggunakan Advanced Handler (LZString/Gzip)
	return DecompressPayload(cipherBytes)
}

func unpadPKCS7(b []byte) []byte {
	length := len(b)
	if length == 0 {
		return b
	}
	unpadding := int(b[length-1])
	return b[:(length - unpadding)]
}
