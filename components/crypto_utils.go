package components

import (
	"crypto/sha512"
	"fmt"
)

type CryptoUtils struct{}

var CryptoUtilsInstance CryptoUtils

func (cu *CryptoUtils) BytesToSha512Hex(data []byte) string {
	hash := sha512.Sum512(data)

	return fmt.Sprintf("%x", hash)
}
