package crypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/usesend0/send0/internal/constant"
)

func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privatekey, &privatekey.PublicKey, nil
}

func EncodedToPrivateKey(str string) (*rsa.PrivateKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return BytesToPrivateKey(keyBytes)
}

func PrivateKeyToBytes(privateKey *rsa.PrivateKey) ([]byte, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytes,
	}), nil
}

func PublicKeyToBytes(key *rsa.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}), nil
}

func PublicKeyToEncoded(key *rsa.PublicKey) (string, error) {
	bytes, err := PublicKeyToBytes(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func BytesToPrivateKey(privateKeyBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyBytes)
	b := block.Bytes
	key, err := x509.ParsePKCS8PrivateKey(b)
	if err != nil {
		return nil, err
	}
	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to parse private key")
	}

	return rsaPrivateKey, nil
}

func BytesToPublicKey(publicKeyBytes []byte) (*rsa.PublicKey, error) {
	if publicKeyBytes == nil {
		return nil, errors.New("public key is nil")
	}
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode public key")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPublicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, err
	}

	return rsaPublicKey, nil
}

func SignHash(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, data)
}

func GenerateHash(data []byte) ([]byte, error) {
	hash := crypto.SHA256.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func GenerateSecret() (string, error) {
	return GenerateRandomString(constant.SecretKeyLength, constant.SecretChars)
}

func GenerateRandomString(length int, chars []rune) (string, error) {
	max := len(chars) - 1
	str := make([]rune, length)
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "", nil
	}
	for i, rb := range randBytes {
		str[i] = chars[int(rb)%max]
	}

	return string(str), nil
}

func TrimPrefixAndSuffix(keyBytes []byte) string {
	key := bytes.NewBuffer(keyBytes).String()
	parts := strings.Split(key, "\n")

	// remove first and last element and join the rest without any space
	return strings.Join(parts[1:len(parts)-2], "")
}
