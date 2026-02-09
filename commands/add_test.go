package commands

import (
	"fmt"
	"testing"

	"github.com/guregu/null/v6"
)

func TestParseLineRange(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantStart null.Int
		wantEnd   null.Int
		wantErr   bool
	}{
		{"empty", "", null.Int{}, null.Int{}, false},
		{"single line", "42", null.IntFrom(42), null.IntFrom(42), false},
		{"range", "10,35", null.IntFrom(10), null.IntFrom(35), false},
		{"same start and end", "1,1", null.IntFrom(1), null.IntFrom(1), false},
		{"large range", "100,200", null.IntFrom(100), null.IntFrom(200), false},
		{"non-numeric", "abc", null.Int{}, null.Int{}, true},
		{"non-numeric start", "abc,42", null.Int{}, null.Int{}, true},
		{"non-numeric end", "42,abc", null.Int{}, null.Int{}, true},
		{"decimal", "10.5,20", null.Int{}, null.Int{}, true},
		{"start exceeds end", "35,10", null.Int{}, null.Int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseLineRange(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLineRange(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !nullIntEqual(start, tt.wantStart) || !nullIntEqual(end, tt.wantEnd) {
				t.Errorf("parseLineRange(%q) = (%s, %s), want (%s, %s)",
					tt.raw, fmtNullInt(start), fmtNullInt(end), fmtNullInt(tt.wantStart), fmtNullInt(tt.wantEnd))
			}
		})
	}
}

func nullIntEqual(a, b null.Int) bool {
	if !a.Valid && !b.Valid {
		return true
	}
	if a.Valid != b.Valid {
		return false
	}
	return a.Int64 == b.Int64
}

func fmtNullInt(n null.Int) string {
	if !n.Valid {
		return "null"
	}
	return fmt.Sprintf("%d", n.Int64)
}
