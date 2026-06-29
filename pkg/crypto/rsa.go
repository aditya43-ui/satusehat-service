package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"strings"
)

// GenerateRSAKeyPair menghasilkan private dan public key RSA dalam format string PEM.
// Umumnya Satu Sehat mewajibkan RSA dengan panjang minimal 2048 bit.
func GenerateRSAKeyPair(bits int) (privateKeyPEM string, publicKeyPEM string, err error) {
	// 1. Generate RSA Private Key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("gagal mengenerate private key: %w", err)
	}

	// 2. Encode Private Key ke format PEM (PKCS#1)
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}
	privateKeyPEM = string(pem.EncodeToMemory(privBlock))

	// 3. Extract dan Encode Public Key ke format PEM (SPKI)
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("gagal marshal public key: %w", err)
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	// Output standard PEM dari Go (pem.EncodeToMemory) sudah identik dengan openssl_pkey_get_details di PHP.
	publicKeyPEM = string(pem.EncodeToMemory(pubBlock))

	return privateKeyPEM, publicKeyPEM, nil
}

// Public Key resmi Kemenkes (Production / Staging) untuk mengenkripsi kunci AES.
const KemkesPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxLwvebfOrPLIODIxAwFp
4Qhksdtn7bEby5OhkQNLTdClGAbTe2tOO5Tiib9pcdruKxTodo481iGXTHR5033I
A5X55PegFeoY95NH5Noj6UUhyTFfRuwnhtGJgv9buTeBa4pLgHakfebqzKXr0Lce
/Ff1MnmQAdJTlvpOdVWJggsb26fD3cXyxQsbgtQYntmek2qvex/gPM9Nqa5qYrXx
8KuGuqHIFQa5t7UUH8WcxlLVRHWOtEQ3+Y6TQr8sIpSVszfhpjh9+Cag1EgaMzk+
HhAxMtXZgpyHffGHmPJ9eXbBO008tUzrE88fcuJ5pMF0LATO6ayXTKgZVU0WO/4e
iQIDAQAB
-----END PUBLIC KEY-----`

const (
	beginEncryptedMsg = "-----BEGIN ENCRYPTED MESSAGE-----\r\n"
	endEncryptedMsg   = "-----END ENCRYPTED MESSAGE-----"
)

// EncryptSatuSehatPayload mengenkripsi JSON payload sesuai standar SatuSehat (AES-256-GCM & RSA-OAEP).
func EncryptSatuSehatPayload(message []byte) (string, error) {
	block, _ := pem.Decode([]byte(KemkesPublicKeyPEM))
	if block == nil {
		return "", errors.New("gagal mem-parsing Kemenkes Public Key")
	}
	pubKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	kemkesPubKey := pubKeyAny.(*rsa.PublicKey)

	// 1. Generate 32 bytes AES Symmetric Key
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return "", err
	}

	// 2. Encrypt AES Key menggunakan Kemenkes Public Key (RSA-OAEP)
	wrappedAesKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, kemkesPubKey, aesKey, nil)
	if err != nil {
		return "", err
	}

	// 3. Encrypt payload menggunakan AES-256-GCM
	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12) // 12 bytes IV / Nonce
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal menggabungkan Ciphertext + Authentication Tag (16 bytes)
	ciphertextWithTag := aesgcm.Seal(nil, nonce, message, nil)

	// 4. Concat: WrappedKey(256) + IV(12) + CipherText + Tag(16)
	encryptedData := append(nonce, ciphertextWithTag...)
	payload := append(wrappedAesKey, encryptedData...)

	// 5. Base64 Encode & Chunk Split (76 characters)
	b64 := base64.StdEncoding.EncodeToString(payload)
	var chunked strings.Builder
	for i := 0; i < len(b64); i += 76 {
		end := i + 76
		if end > len(b64) {
			end = len(b64)
		}
		chunked.WriteString(b64[i:end])
		chunked.WriteString("\r\n")
	}

	return beginEncryptedMsg + chunked.String() + endEncryptedMsg, nil
}

// DecryptSatuSehatPayload mendekripsi response Kemenkes yang telah dienkripsi.
func DecryptSatuSehatPayload(encryptedStr string, privPEM string) ([]byte, error) {
	startIdx := strings.Index(encryptedStr, "-----BEGIN ENCRYPTED MESSAGE-----")
	endIdx := strings.Index(encryptedStr, "-----END ENCRYPTED MESSAGE-----")
	if startIdx == -1 || endIdx == -1 {
		return nil, errors.New("respons tidak memiliki tag ENCRYPTED MESSAGE")
	}

	b64Data := encryptedStr[startIdx+len("-----BEGIN ENCRYPTED MESSAGE-----") : endIdx]
	b64Data = strings.ReplaceAll(strings.ReplaceAll(b64Data, "\r", ""), "\n", "")

	binaryData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Data))
	if err != nil || len(binaryData) < 256+12 {
		return nil, errors.New("gagal me-decode base64 payload / panjang payload tidak valid")
	}

	wrappedKey := binaryData[:256]
	iv := binaryData[256 : 256+12]
	ciphertextWithTag := binaryData[256+12:]

	block, _ := pem.Decode([]byte(privPEM))
	privKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, wrappedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal me-unwrap kunci AES: %w", err)
	}

	blockCipher, _ := aes.NewCipher(aesKey)
	aesgcm, _ := cipher.NewGCM(blockCipher)

	return aesgcm.Open(nil, iv, ciphertextWithTag, nil)
}
