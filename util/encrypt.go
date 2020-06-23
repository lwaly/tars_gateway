package util

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

func DecryptPkcsByBytes(ciphertext, privatekey []byte) (decryptedtext []byte, err error) {
	var tempPrivatekey *rsa.PrivateKey
	if tempPrivatekey, err = LoadPrivateKeyBytes(privatekey); nil != err {
		return
	}

	if 0 != len(ciphertext)%((tempPrivatekey.N.BitLen()+7)/8) {
		return nil, fmt.Errorf("error input.%d %d\n", len(ciphertext), (tempPrivatekey.N.BitLen()+7)/8)
	}
	for i := 0; i < len(ciphertext); i += (tempPrivatekey.N.BitLen() + 7) / 8 {
		decryptedtextTemp, err := rsa.DecryptPKCS1v15(rand.Reader, tempPrivatekey, ciphertext)
		if err != nil {
			return nil, fmt.Errorf("RSA DecryptPkcs failed, error=%s\n", err.Error())
		}
		decryptedtext = append(decryptedtext, decryptedtextTemp...)
	}

	return decryptedtext, err
}

func DecryptByBytes(ciphertext, privatekey []byte) (decryptedtext []byte, err error) {
	var tempPrivatekey *rsa.PrivateKey
	if tempPrivatekey, err = LoadPrivateKeyBytes(privatekey); nil != err {
		return
	}
	if 0 != len(ciphertext)%((tempPrivatekey.N.BitLen()+7)/8) {
		return nil, fmt.Errorf("error input.%d %d\n", len(ciphertext), (tempPrivatekey.N.BitLen()+7)/8)
	}
	for i := 0; i < len(ciphertext); i += (tempPrivatekey.N.BitLen() + 7) / 8 {
		sha256hash := sha256.New()
		decryptedtextTemp, err := rsa.DecryptOAEP(sha256hash, rand.Reader, tempPrivatekey, ciphertext[i:i+(tempPrivatekey.N.BitLen()+7)/8], nil)
		if err != nil {
			return nil, fmt.Errorf("RSA DecryptOAEP failed, error=%s\n", err.Error())
		}
		decryptedtext = append(decryptedtext, decryptedtextTemp...)
	}

	return decryptedtext, err
}

// decrypt
func Decrypt(ciphertext []byte, privatekey *rsa.PrivateKey) (decryptedtext []byte, err error) {
	sha256hash := sha256.New()
	if nil == privatekey || 0 != len(ciphertext)%((privatekey.N.BitLen()+7)/8) {
		return nil, fmt.Errorf("error input.%d %d\n", len(ciphertext), (privatekey.N.BitLen()+7)/8)
	}
	for i := 0; i < len(ciphertext); i += ((privatekey.N.BitLen() + 7) / 8) {
		decryptedtextTemp, err := rsa.DecryptOAEP(sha256hash, rand.Reader, privatekey, ciphertext[i:i+((privatekey.N.BitLen()+7)/8)], nil)
		if err != nil {
			return nil, fmt.Errorf("RSA DecryptOAEP failed, error=%s\n", err.Error())
		}
		decryptedtext = append(decryptedtext, decryptedtextTemp...)
	}

	return decryptedtext, nil
}

// decrypt
func DecryptPkcs(ciphertext []byte, privatekey *rsa.PrivateKey) (decryptedtext []byte, err error) {
	if nil == privatekey || 0 != len(ciphertext)%((privatekey.N.BitLen()+7)/8) {
		return nil, fmt.Errorf("error input.%d %d\n", len(ciphertext), (privatekey.N.BitLen()+7)/8)
	}
	for i := 0; i < len(ciphertext); i += (privatekey.N.BitLen() + 7) / 8 {
		decryptedtextTemp, err := rsa.DecryptPKCS1v15(rand.Reader, privatekey, ciphertext[i:i+(privatekey.N.BitLen()+7)/8])
		if err != nil {
			return nil, fmt.Errorf("RSA DecryptPkcs failed, error=%s\n", err.Error())
		}
		decryptedtext = append(decryptedtext, decryptedtextTemp...)
	}

	return decryptedtext, err
}

// encrypt
func Encrypt(plaintext []byte, publickey *rsa.PublicKey) (ciphertext []byte, err error) {
	l := 0
	length := len(plaintext)
	sha256hash := sha256.New()
	for i := 0; i < length; {
		if length-i < (publickey.N.BitLen()+7)/8-2*sha256hash.Size()-2 {
			l = length - i
		} else {
			l = (publickey.N.BitLen()+7)/8 - 2*sha256hash.Size() - 2
		}
		ciphertextTemp, err := rsa.EncryptOAEP(sha256hash, rand.Reader, publickey, plaintext[i:i+l], nil)
		if err != nil {
			return nil, fmt.Errorf("RSA EncryptOAEP failed, error=%s\n", err.Error())
		}
		ciphertext = append(ciphertext, ciphertextTemp...)
		i += l
	}
	return ciphertext, err
}

// encrypt
func EncryptPkcs(plaintext []byte, publickey *rsa.PublicKey) (ciphertext []byte, err error) {
	l := 0
	length := len(plaintext)
	for i := 0; i < length; {
		if length-i < (publickey.N.BitLen()+7)/8-11 {
			l = length - i
		} else {
			l = (publickey.N.BitLen()+7)/8 - 11
		}
		ciphertextTemp, err := rsa.EncryptPKCS1v15(rand.Reader, publickey, plaintext[i:i+l])
		if err != nil {
			return nil, fmt.Errorf("RSA EncryptPkcs failed, error=%s\n", err.Error())
		}
		ciphertext = append(ciphertext, ciphertextTemp...)
		i += l
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

func LoadPrivateKeyBytes(b []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("Private key error")
	}

	pubkeyinterface, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubkeyinterface, nil
}

func LoadPublicKeyBytes(b []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(b)
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
