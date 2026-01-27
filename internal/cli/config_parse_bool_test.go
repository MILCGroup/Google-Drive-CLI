package cli

import "testing"

func TestParseBool(t *testing.T) {
	truthy := []string{"true", "1", "yes", "on", " TRUE ", "Yes"}
	for _, v := range truthy {
		if !parseBool(v) {
			t.Fatalf("expected true for %q", v)
		}
	}

	falsey := []string{"false", "0", "no", "off", "", "maybe"}
	for _, v := range falsey {
		if parseBool(v) {
			t.Fatalf("expected false for %q", v)
		}
	}
}
