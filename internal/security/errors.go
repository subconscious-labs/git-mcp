// Package security hosts the middleware that guards every tool invocation:
// argument sanitization, path isolation, safe process execution and output
// sanitization. All layers are independent and composable.
package security

import "fmt"

// ViolationKind classifies a security rejection for logging and tests.
type ViolationKind string

const (
	// ViolationFlagInjection marks an argument that could be interpreted
	// as a flag by the underlying git command (e.g. starting with '-').
	ViolationFlagInjection ViolationKind = "flag_injection"

	// ViolationPathTraversal marks a repo_path that escapes the allowed
	// root, whether via "..", an absolute path, or a symlink.
	ViolationPathTraversal ViolationKind = "path_traversal"

	// ViolationSymlinkEscape marks a path whose resolution exits the root
	// through a symbolic link.
	ViolationSymlinkEscape ViolationKind = "symlink_escape"

	// ViolationInvalidRevision marks a value that git rev-parse rejects
	// as a valid commit-ish.
	ViolationInvalidRevision ViolationKind = "invalid_revision"

	// ViolationSensitiveData marks output that was redacted before being
	// returned to the client (secrets, tokens, keys).
	ViolationSensitiveData ViolationKind = "sensitive_data"
)

// SecurityError is returned whenever a middleware layer rejects a request
// or transforms a response. Callers must not unwrap it to recover data.
type SecurityError struct {
	Kind   ViolationKind
	Field  string
	Detail string
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("security violation [%s] field=%q: %s", e.Kind, e.Field, e.Detail)
}

// NewViolation constructs a SecurityError. It is the only intended
// constructor so that every rejection carries a Kind.
func NewViolation(kind ViolationKind, field, detail string) error {
	return &SecurityError{Kind: kind, Field: field, Detail: detail}
}
