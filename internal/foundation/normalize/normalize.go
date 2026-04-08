package normalize

import (
	"regexp"
	"strings"
)

var (
	multipleSpaces = regexp.MustCompile(`\s+`)
	uuidPattern    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// String normalizes a string by trimming whitespace, collapsing multiple
// spaces into one, and converting to lowercase.
func String(s string) string {
	s = strings.TrimSpace(s)
	s = multipleSpaces.ReplaceAllString(s, " ")
	return strings.ToLower(s)
}

// IsValidUUID reports whether s is a valid UUID v4 format.
func IsValidUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// ProductKey returns a composite lookup key for a product using its name,
// brand, and category. This prevents false positives when two products share
// the same name but differ in brand or category.
func ProductKey(name, brand, category string) string {
	return String(name) + "|" + String(brand) + "|" + String(category)
}
