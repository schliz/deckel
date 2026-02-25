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
		"add":         addInt,
		"sub":         subInt,
		"map":         makeMap,
		"withinMinutes": withinMinutes,
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

// addInt returns a + b.
func addInt(a, b int) int { return a + b }

// subInt returns a - b.
func subInt(a, b int) int { return a - b }

// makeMap creates a map from alternating key-value pairs for use in templates.
func makeMap(pairs ...any) map[string]any {
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); ok {
			m[key] = pairs[i+1]
		}
	}
	return m
}

// withinMinutes returns true if t is within the last n minutes from now.
func withinMinutes(t time.Time, minutes int) bool {
	return time.Since(t) <= time.Duration(minutes)*time.Minute
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
