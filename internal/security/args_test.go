package security

import "testing"

func TestValidatePositional_Rejects(t *testing.T) {
	cases := []struct{ name, field, value string }{
		{"leading dash", "branch", "--force"},
		{"upload-pack injection", "branch", "--upload-pack=/tmp/pwn"},
		{"output injection", "target", "--output=/tmp/pwn"},
		{"empty", "branch", ""},
		{"null byte", "branch", "main\x00evil"},
		{"newline", "branch", "main\nevil"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := ValidatePositional(c.field, c.value); err == nil {
				t.Fatalf("expected rejection for %q", c.value)
			}
		})
	}
}

func TestValidatePositional_Accepts(t *testing.T) {
	for _, v := range []string{
		"main",
		"feature/foo",
		"v1.0.0",
		"a/b-c_d",
		"HEAD:refs/for/master/topic", // AGit refspec
		"+main:refs/heads/foo",       // force refspec (allowed by design)
	} {
		if err := ValidatePositional("branch", v); err != nil {
			t.Errorf("false positive: %q (%v)", v, err)
		}
	}
}

// TestCVE2025_68144_Regression pins the exact payload from the advisory.
func TestCVE2025_68144_Regression(t *testing.T) {
	if err := ValidatePositional("target", "--output=/tmp/pwn"); err == nil {
		t.Fatal("CVE regression: flag injection not blocked")
	}
}
