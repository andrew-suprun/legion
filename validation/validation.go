package validation

import (
	"regexp"
)

var EmailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.([^@\s]+){2,}$`)

func IsValidEmail(str string) bool {
	return EmailPattern.MatchString(str)
}
