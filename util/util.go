package util

import (
	"crypto/sha512"
	"encoding/hex"
	"unicode"

	"github.com/lestrrat-go/bufferpool"
)

// Hash password with salt and 30-times stretching using SHA512.
func Hash(password, salt string) string {
	hasher := sha512.New()
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)
	buf.WriteString(password)
	encoder := hex.NewEncoder(buf)
	for i := 0; i < 30; i++ {
		buf.WriteString(salt)
		buf.WriteTo(hasher)
		encoder.Write(hasher.Sum(nil))
		hasher.Reset()
	}

	return buf.String()
}

// SHA512Digest hashes using SHA512 and returns its hex digest.
func SHA512Digest(s string) string {
	h := sha512.Sum512([]byte(s))
	return hex.EncodeToString(h[:])
}

// IsDigit returns given string is all digit or not.
func IsDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
