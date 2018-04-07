package util_test

import (
	"testing"

	"github.com/nasa9084/ident/util"
)

const (
	password = "password"
	salt     = "userid"
	expected = "6be030e254dcc2d15a4fc016751aced63eab237dc8bfbf5640d4d028fb347b877b112b19a624f3aebbfd1404d0dcdd5c96b529c315e9ee7d0428f15e08437b19"
)

func TestHash(t *testing.T) {
	hash := util.Hash(password, salt)
	if hash != expected {
		t.Errorf("%s != %s", hash, expected)
		return
	}
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.Hash(password, salt)
	}
}

func TestIsDigit(t *testing.T) {
	candidates := []struct {
		in       string
		expected bool
	}{
		{"12345", true},
		{"", true},
		{"12345a", false},
		{"123+", false},
		{"a", false},
	}
	for _, c := range candidates {
		out := util.IsDigit(c.in)
		if out != c.expected {
			t.Errorf("%t != %t (input: %s)", out, c.expected, c.in)
		}
	}
}
