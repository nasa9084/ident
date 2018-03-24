package totp

import (
	"encoding/base32"
	"fmt"
	"strconv"
	"time"

	"github.com/nasa9084/go-hotp"
)

const keyURITemplate = "otpauth://totp/%s:%s?issuer=%s&secret=%s&digits=%d"

// Generator generates Time-based One-Time Password
type Generator struct {
	TimeStep  uint64 // X in RFC6238
	StartTime int64  // T0 in RFC6238, default 0 is OK
	Secret    string // shared secret for generate hotp
	Digit     int
}

// New returns new 6-digit TOTP generator.
func New(secret string) *Generator {
	return &Generator{
		Secret: secret,
		Digit:  6,
	}
}

// Generate OTP
func (g *Generator) Generate() int64 {
	if g.TimeStep == 0 {
		g.TimeStep = 30
	}
	now := time.Now().UTC().Unix()
	t := (now - g.StartTime) / int64(g.TimeStep)
	h := hotp.Generator{
		Secret:  g.Secret,
		Digit:   g.Digit,
		Counter: uint64(t),
	}
	return h.Generate()
}

// GenerateString generates string OTP.
func (g *Generator) GenerateString() string {
	return fmt.Sprintf("%0"+strconv.Itoa(g.Digit)+"d", g.Generate())
}

// URI returns TOTP key URI.
func (g *Generator) URI(issuer, account string) string {
	return fmt.Sprintf(keyURITemplate, issuer, account, issuer, base32.StdEncoding.EncodeToString([]byte(g.Secret)), g.Digit)
}
