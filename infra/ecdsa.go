package infra

import (
	"crypto/ecdsa"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
)

func LoadPrivateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return jwt.ParseECPrivateKeyFromPEM(b)
}
