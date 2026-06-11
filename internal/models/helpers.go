package models

import (
	"time"
)

// ParseTime parses a SQLite datetime string to time.Time.
// SQLite's datetime() function returns "2006-01-02 15:04:05" format.
func ParseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

// ParseTimePtr parses a SQLite datetime string to *time.Time.
func ParseTimePtr(s string) *time.Time {
	t := ParseTime(s)
	if t.IsZero() {
		return nil
	}
	return &t
}

// FormatTime formats time.Time to SQLite datetime string.
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
