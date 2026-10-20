package security

import (
	"regexp"
	"strings"
)

// redaction describes a single secret pattern to strip from responses.
// Patterns are applied in declaration order. Specific patterns must come
// before generic ones (SECRET_KV last) so that, for example, a GitHub
// token inside "token=ghp_..." is labeled as GH_TOKEN rather than as a
// generic key=value credential.
type redaction struct {
	name string
	re   *regexp.Regexp
}

var redactions = []redaction{
	{"AWS_KEY", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"GH_TOKEN", regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`)},
	{"JWT", regexp.MustCompile(`eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`)},
	{"PEM", regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+PRIVATE KEY-----`)},
	{"SECRET_KV", regexp.MustCompile(`(?i)(password|secret|api[_-]?key|token)\s*[:=]\s*\S{8,}`)},
}

var promptInjection = regexp.MustCompile(
	`(?i)(ignore\s+(previous|above)\s+instructions|disregard\s+all\s+prior|you\s+are\s+now\s+)`,
)

// SanitizeOutput redacts known secret patterns and wraps responses that
// contain likely prompt-injection payloads in untrusted markers.
//
// The returned findings slice names the categories that matched; it never
// contains the matched data itself, so it is safe to log. The transformed
// output is what the client sees.
//
// A generic redaction (SECRET_KV) is skipped on any match that already
// contains a "<REDACTED:" marker, so that a more specific pattern applied
// earlier is never overwritten by a later, broader one.
func SanitizeOutput(in string) (string, []string) {
	var findings []string
	out := in
	for _, r := range redactions {
		if !r.re.MatchString(out) {
			continue
		}
		if r.name == "SECRET_KV" {
			next := r.re.ReplaceAllStringFunc(out, func(m string) string {
				if strings.Contains(m, "<REDACTED:") {
					return m // already covered by a more specific pattern
				}
				return "<REDACTED:" + r.name + ">"
			})
			if next != out && !containsString(findings, r.name) {
				findings = append(findings, r.name)
			}
			out = next
			continue
		}
		if !containsString(findings, r.name) {
			findings = append(findings, r.name)
		}
		out = r.re.ReplaceAllString(out, "<REDACTED:"+r.name+">")
	}
	if promptInjection.MatchString(out) {
		findings = append(findings, "PROMPT_INJECTION")
		out = "<<<UNTRUSTED>>>\n" + out + "\n<<<END>>>"
	}
	return out, findings
}

// HasRedaction reports whether a string contains a redaction marker.
// Used by tests to assert that secrets were stripped.
func HasRedaction(s string) bool {
	return strings.Contains(s, "<REDACTED:")
}

func containsString(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
