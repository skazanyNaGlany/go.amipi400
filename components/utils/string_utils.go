package utils

import "unicode"

type StringUtils struct{}

var StringUtilsInstance StringUtils

func (st *StringUtils) IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
