package types

import (
	"crypto/sha256"
	"fmt"
)

type Hash [sha256.Size]byte

func (h Hash) ToString() string {
	// return string(h[:])
	outString := ""
	for _, b := range h {
		outString = outString + fmt.Sprint(b)
	}
	return outString
}
