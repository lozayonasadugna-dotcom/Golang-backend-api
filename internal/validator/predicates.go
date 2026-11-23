package validator

import (
	"regexp"
	"slices"
)

var PhoneRx = regexp.MustCompile(`^0[79][0-9]{8}$`)

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func In(value string, values ...string) bool {
	return slices.Contains(values, value)
}

func Unique(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
