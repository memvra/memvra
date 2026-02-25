package cli

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		err   bool
	}{
		{"24h", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"1d", 24 * time.Hour, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.err && err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct{ input, want string }{
		{"decision", "Decision"},
		{"", ""},
		{"TODO", "TODO"},
	}
	for _, tt := range tests {
		if got := capitalize(tt.input); got != tt.want {
			t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPluralS(t *testing.T) {
	if pluralS(1) != "" {
		t.Error("expected empty for 1")
	}
	if pluralS(2) != "s" {
		t.Error("expected 's' for 2")
	}
	if pluralS(0) != "s" {
		t.Error("expected 's' for 0")
	}
}
