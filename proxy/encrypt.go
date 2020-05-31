package proxy

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// decrypt
func Decrypt(ciphertext []byte, privatekey *rsa.PrivateKey) ([]byte, error) {
	// decodedtext, err := base64.StdEncoding.DecodeString(ciphertext)
	// if err != nil {
	// 	return "", fmt.Errorf("base64 decode failed, error=%s\n", err.Error())
	// }

	sha256hash := sha256.New()
	decryptedtext, err := rsa.DecryptOAEP(sha256hash, rand.Reader, privatekey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA DecryptOAEP failed, error=%s\n", err.Error())
	}

	return decryptedtext, err
}

// encrypt
func Encrypt(plaintext []byte, publickey *rsa.PublicKey) ([]byte, error) {
	label := []byte("")
	sha256hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(sha256hash, rand.Reader, publickey, plaintext, label)
	if err != nil {
		return nil, fmt.Errorf("RSA EncryptOAEP failed, error=%s\n", err.Error())
	}
	return ciphertext, err
}

// decrypt
func DecryptPkcs(ciphertext []byte, privatekey *rsa.PrivateKey) ([]byte, error) {
	decryptedtext, err := rsa.DecryptPKCS1v15(nil, privatekey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("RSA DecryptPkcs failed, error=%s\n", err.Error())
	}

	return decryptedtext, err
}

// encrypt
func EncryptPkcs(plaintext []byte, publickey *rsa.PublicKey) ([]byte, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(nil, publickey, plaintext)
	if err != nil {
		return nil, fmt.Errorf("RSA EncryptPkcs failed, error=%s\n", err.Error())
	}
	return ciphertext, err
}

// Load private key from base64
func LoadPrivateKeyBase64(base64key string) (*rsa.PrivateKey, error) {
	keybytes, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed, error=%s\n", err.Error())
	}

	privatekey, err := x509.ParsePKCS1PrivateKey(keybytes)
	if err != nil {
		return nil, errors.New("parse private key error!")
	}

	return privatekey, nil
}

func LoadPublicKeyBase64(base64key string) (*rsa.PublicKey, error) {
	keybytes, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed, error=%s\n", err.Error())
	}

	pubkeyinterface, err := x509.ParsePKIXPublicKey(keybytes)
	if err != nil {
		return nil, err
	}

	publickey := pubkeyinterface.(*rsa.PublicKey)
	return publickey, nil
}

// Load private key from private key file
func LoadPrivateKeyFile(keyfile string) (*rsa.PrivateKey, error) {
	keybuffer, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(keybuffer))
	if block == nil {
		return nil, errors.New("private key error!")
	}

	privatekey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse private key error!")
	}

	return privatekey, nil
}

func LoadPublicKeyFile(keyfile string) (*rsa.PublicKey, error) {
	keybuffer, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keybuffer)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubkeyinterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publickey := pubkeyinterface.(*rsa.PublicKey)
	return publickey, nil
}

func GenerateKey(bit int) (*rsa.PrivateKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, bit)
	if err != nil {
		return nil, err
	}

	return privatekey, nil
}

func DumpPrivateKeyFile(privatekey *rsa.PrivateKey, filename string) error {
	var keybytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keybytes,
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

func DumpPrivateKeyBytes(privatekey *rsa.PrivateKey) (err error, b []byte) {
	var keybytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keybytes,
	}
	reader := bytes.NewBuffer(nil)
	err = pem.Encode(reader, block)
	if err != nil {
		return err, nil
	}
	return err, reader.Bytes()
}

func DumpPublicKeyFile(publickey *rsa.PublicKey, filename string) error {
	keybytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keybytes,
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

func DumpPublicKeyBytes(publickey *rsa.PublicKey) (err error, b []byte) {
	keybytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		return err, nil
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keybytes,
	}

	reader := bytes.NewBuffer(nil)
	err = pem.Encode(reader, block)
	if err != nil {
		return err, nil
	}
	return err, reader.Bytes()
}
