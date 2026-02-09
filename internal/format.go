package internal

import (
	"fmt"
	"strconv"

	"github.com/guregu/null/v6"
)

// Pluralize returns singular if n == 1, plural otherwise.
func Pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// ShortSHA returns the first 7 characters of a SHA hash.
func ShortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// ShortID returns the first 8 characters of a UUID string.
func ShortID(id fmt.Stringer) string {
	s := id.String()
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// FormatLineRange formats null.Int start/end as "N" or "N-M".
// Returns "" if start is null.
func FormatLineRange(startLine, endLine null.Int) string {
	if !startLine.Valid {
		return ""
	}
	s := strconv.FormatInt(startLine.Int64, 10)
	if endLine.Valid && endLine.Int64 != startLine.Int64 {
		s += "-" + strconv.FormatInt(endLine.Int64, 10)
	}
	return s
}
