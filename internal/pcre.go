package internal

import (
	"github.com/Valgard/go-pcre"
)

func MatchAll(pattern string, subject string, flags int, offset int) ([]pcre.Match, bool) {
	var matches []pcre.Match

	regexp := pcre.MustCompile(pattern, flags)
	defer regexp.FreeRegexp()

	m, err := regexp.FindAllOffset(subject, 0, offset)
	if err != nil {
		return matches, false
	}

	if len(m) == 0 {
		return m, false
	}

	for _, match := range m {
		// skip matches before offset
		if offset > match.Loc[0] {
			continue
		}
		matches = append(matches, match)
	}

	if len(matches) == 0 {
		return matches, false
	}

	return matches, true
}

func Match(pattern string, subject string, flags int, offset int) (pcre.Match, bool) {
	matches, ok := MatchAll(pattern, subject, flags, offset)
	if !ok {
		return pcre.Match{}, false
	}

	return matches[0], true
}

func ReplaceCallback(pattern string, subject string, flags int, callable func(matches pcre.Match) (string, error)) (string, error) {
	matches, ok := MatchAll(pattern, subject, flags, 0)
	if !ok {
		return subject, nil
	}

	var value string
	for _, match := range matches {
		var err error
		value, err = callable(match)
		if err != nil  {
			return "", err
		}
	}

	return value, nil
}
