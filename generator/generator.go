package generator

import (
	"crypto/ecdsa"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// TimeFunc returns current time.
// for testing, this function is overridable.
var TimeFunc = time.Now

// NewToken returns a new signed JSON Web Token.
func NewToken(privKey *ecdsa.PrivateKey, userID string) (string, error) {
	now := TimeFunc()
	claims := jwt.MapClaims{
		"iat":     now.Unix(),
		"exp":     now.Add(1 * time.Hour).Unix(),
		"user_id": userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privKey)
}
