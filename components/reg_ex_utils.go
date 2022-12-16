package components

import "regexp"

type RegExUtils struct{}

func (ru RegExUtils) FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)
	results := map[string]string{}

	if match == nil {
		return results
	}

	for i, name := range match {
		if i == 0 {
			// skip 0 match, since it will be whole line
			continue
		}

		results[regex.SubexpNames()[i]] = name
	}

	return results
}
