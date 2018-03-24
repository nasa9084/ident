package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	privKey := genPrivKey()
	genPubKey(privKey)
}

func genPrivKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		panic(err)
	}
	privKeyFile, err := os.OpenFile("key/id_ecdsa", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(privKeyFile, &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	if err != nil {
		panic(err)
	}
	return privKey
}

func genPubKey(privKey *ecdsa.PrivateKey) {
	pubKey := privKey.Public()
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		panic(err)
	}
	pubKeyFile, err := os.OpenFile("key/id_ecdsa.pub", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(pubKeyFile, &pem.Block{
		Type:  "ECDSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	if err != nil {
		panic(err)
	}
}
