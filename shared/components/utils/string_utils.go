package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/thoas/go-funk"
)

type StringUtils struct{}

var StringUtilsInstance StringUtils
var ISALNUM = regexp.MustCompile(`^[a-zA-Z0-9]*$`)

func (st *StringUtils) IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

func (st *StringUtils) StringUnify(strToUnify string, exclude ...string) string {
	unified := ""

	for _, iChar := range strToUnify {
		iCharStr := string(iChar)

		if funk.ContainsString(exclude, iCharStr) {
			unified += iCharStr
			continue
		}

		iCharStr = strings.ToLower(iCharStr)

		if st.IsAlNum(iCharStr) {
			unified += iCharStr
		} else {
			unified += " "
		}

	}

	unified = strings.TrimSpace(unified)
	unified2 := ""

	for _, iStr := range strings.Split(unified, " ") {
		iStr = strings.TrimSpace(iStr)

		if iStr == "" {
			continue
		}

		unified2 += iStr + " "
	}

	unified2 = strings.TrimSpace(unified2)

	return unified2
}

func (st *StringUtils) IsAlNum(s string) bool {
	return ISALNUM.MatchString(s)
}

func (st *StringUtils) StringToInt(s string, base int, bitSize int) (int, error) {
	i, err := strconv.ParseInt(s, base, bitSize)

	if err != nil {
		return 0, err
	}

	return int(i), nil
}
