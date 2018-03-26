package util

import (
	"crypto/sha512"
	"encoding/hex"

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