package xrand

import (
	"fmt"
	"testing"
)

func TestRandomString(t *testing.T) {
	var r = NewRandomStringGenerator(10, Alphanumeric+Symbols)
	var s = r.New()
	fmt.Printf("random: %s\n", s)
}
