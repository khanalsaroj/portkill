package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	Version = "v9.9.9"
	got := String()
	for _, want := range []string{"portkill", "v9.9.9"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}

func TestShort(t *testing.T) {
	Version = "v1.2.3"
	if Short() != "v1.2.3" {
		t.Errorf("Short() = %q, want v1.2.3", Short())
	}
}
