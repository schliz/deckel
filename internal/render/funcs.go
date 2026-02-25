package render

import (
	"fmt"
	"html/template"
	"math"
	"time"
)

// FuncMap returns template helper functions for formatting money, dates, etc.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatCents": formatCents,
		"formatTime":  formatTime,
		"formatDate":  formatDate,
		"abs":         absInt64,
		"neg":         negInt64,
		"seq":         seq,
	}
}

// formatCents converts an int64 amount in cents to German locale currency format.
// e.g., 150 → "1,50 EUR", 1234 → "12,34 EUR", -250 → "-2,50 EUR"
func formatCents(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	euros := cents / 100
	remainder := cents % 100
	prefix := ""
	if negative {
		prefix = "-"
	}
	return fmt.Sprintf("%s%d,%02d EUR", prefix, euros, remainder)
}

// formatTime formats a time.Time as "DD.MM.YYYY HH:MM" (German locale).
func formatTime(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

// formatDate formats a time.Time as "DD.MM.YYYY" (German locale).
func formatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

// absInt64 returns the absolute value of an int64.
func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// negInt64 negates an int64 value.
func negInt64(n int64) int64 {
	return -n
}

// seq returns a slice of ints [0, 1, ..., n-1].
func seq(n int) []int {
	if n <= 0 {
		return nil
	}
	if n > math.MaxInt32 {
		n = math.MaxInt32
	}
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}
