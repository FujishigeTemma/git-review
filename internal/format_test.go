package internal

import (
	"testing"

	"github.com/guregu/null/v6"
)

func TestShortSHA(t *testing.T) {
	tests := []struct {
		name string
		sha  string
		want string
	}{
		{"full SHA", "abcdef1234567890", "abcdef1"},
		{"short SHA", "abc", "abc"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortSHA(tt.sha)
			if got != tt.want {
				t.Errorf("ShortSHA(%q) = %q, want %q", tt.sha, got, tt.want)
			}
		})
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"full UUID", "0194b5a0-1234-7890-abcd-ef1234567890", "0194b5a0"},
		{"short", "abc", "abc"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortID(stringerStr(tt.id))
			if got != tt.want {
				t.Errorf("ShortID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

// stringerStr is a test helper that implements fmt.Stringer for a plain string.
type stringerStr string

func (s stringerStr) String() string { return string(s) }

func TestPluralize(t *testing.T) {
	tests := []struct {
		n        int
		singular string
		plural   string
		want     string
	}{
		{0, "commit", "commits", "commits"},
		{1, "commit", "commits", "commit"},
		{2, "commit", "commits", "commits"},
	}
	for _, tt := range tests {
		got := Pluralize(tt.n, tt.singular, tt.plural)
		if got != tt.want {
			t.Errorf("Pluralize(%d, %q, %q) = %q, want %q", tt.n, tt.singular, tt.plural, got, tt.want)
		}
	}
}

func TestFormatLineRange(t *testing.T) {
	tests := []struct {
		name  string
		start null.Int
		end   null.Int
		want  string
	}{
		{
			name:  "both null",
			start: null.Int{},
			end:   null.Int{},
			want:  "",
		},
		{
			name:  "start valid end null",
			start: null.IntFrom(10),
			end:   null.Int{},
			want:  "10",
		},
		{
			name:  "both valid same",
			start: null.IntFrom(5),
			end:   null.IntFrom(5),
			want:  "5",
		},
		{
			name:  "both valid different",
			start: null.IntFrom(5),
			end:   null.IntFrom(12),
			want:  "5-12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLineRange(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("FormatLineRange() = %q, want %q", got, tt.want)
			}
		})
	}
}
