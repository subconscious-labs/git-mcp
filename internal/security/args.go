package security

import "strings"

// ValidatePositional rejects values that git could interpret as flags
// or that contain control characters. It is applied to every user-supplied
// positional argument (branch names, paths, revisions, refspecs).
//
// This is the primary defense against CVE-2025-68144-style flag injection:
// a value like "--output=/tmp/pwn" would otherwise be treated by git as
// a flag rather than as a positional argument.
func ValidatePositional(field, value string) error {
	if value == "" {
		return NewViolation(ViolationFlagInjection, field, "empty value")
	}
	if strings.HasPrefix(value, "-") {
		return NewViolation(ViolationFlagInjection, field,
			"value starts with '-' and could be interpreted as a git flag")
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return NewViolation(ViolationFlagInjection, field,
				"control character in value")
		}
	}
	return nil
}
