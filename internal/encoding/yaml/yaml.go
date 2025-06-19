// SPDX-FileCopyrightText: 2025 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/goschtalt/goschtalt/internal/encoding"
)

/*
Output in one of a few formats:

If the value is short enough, it is rendered as a single line.
[short_value] is the short value if possible, otherwise it is the value.
[value] is the value if it is too long to fit on a single line.

The general format is:

---
# header comment
key: [short_value] 	# single line comment
  # multi-line comment
  # multi-line comment
  [block characters if needed]
  [value]

Examples:

---
# header comment
key: value         	# single line comment

---
# header comment
key:
  # multi-line comment
  # multi-line comment
  value

---
# header comment
key: "value"       	# single line comment

---
# header comment
key:
  # multi-line comment
  # multi-line comment
  "value"

---
# header comment
key:				# single line comment
  "valuesplit\
  overmultiple\
  lines"

---
# header comment
key:
  # multi-line comment
  # multi-line comment
  "valuesplit\
  overmultiple\
  lines"

These next two forms are not used because they are hard to correctly
determine since multiple spaces cause confusion/challenges.  These are
generally not used for configuration from what I can tell.  If that
changes, we can add them in the future.

---
# header comment
key:				# single line comment
  <
  value split
  over multiple
  lines

---
# header comment
key:
  # multi-line comment
  # multi-line comment
  <
  value split
  over multiple
  lines

*/

// Renderer defines rendering options for YAML output.
type Renderer struct {
	// MaxLineLength is the maximum line length for wrapping.
	// If <=0, no wrapping is done.
	MaxLineLength int

	// TrailingCommentColumn is the column for comments to start at if positive,
	// or number of spaces to indent comments if negative.
	// If zero, comments are not rendered.
	TrailingCommentColumn int

	// SpacesPerIndent is the number of spaces to indent each line.
	// If zero the default is 2.
	SpacesPerIndent int
}

func (r *Renderer) headers(w io.Writer, item encoding.Encodeable) error {
	// Render header comments
	for _, hc := range item.Headers() {
		nl := "\n"
		if strings.HasSuffix(hc, "\n") {
			nl = ""
		}
		if _, err := fmt.Fprintf(w, "%s# %s%s", r.indent(item.Indent()), hc, nl); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) Encode(w io.Writer, item encoding.Encodeable) error {
	if _, err := fmt.Fprintln(w, "---"); err != nil {
		return err
	}
	return r.encode(w, item)
}

func (r *Renderer) encode(w io.Writer, item encoding.Encodeable) error {
	if err := r.headers(w, item); err != nil {
		return err
	}

	if err := r.node(w, item); err != nil {
		return err
	}

	if item.Children() != nil {
		children := item.Children()
		sort.Sort(children)
		for _, child := range children {
			if err := r.encode(w, child); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Renderer) node(w io.Writer, item encoding.Encodeable) error {
	inline := item.Inline()

	// Render the sequence/key part of the line
	line := fmt.Sprintf("%s-", r.indent(item.Indent()))
	if item.Key() != nil {
		line = fmt.Sprintf("%s%s:", r.indent(item.Indent()), *item.Key())
	}
	var v []string
	var block string
	if item.Value() != nil {
		v, block = formatValue(item.Value(), r.maxLineLength(len(line)-1))
	}

	if len(inline) <= 1 && len(v) == 1 {
		line += " " + v[0]
	}

	spaces := max(r.TrailingCommentColumn-len(line), 1)

	var blockWritten bool
	if item.Indent() >= 0 {
		// Write the line with the key and value and optionally inline comment
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}

		if len(inline) == 1 {
			if _, err := fmt.Fprintf(w, "%s# %s", strings.Repeat(" ", spaces), inline[0]); err != nil {
				return err
			}
		}
		if len(inline) == 0 && block != "" {
			if _, err := fmt.Fprintf(w, " %s", block); err != nil {
				return err
			}
			blockWritten = true
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}

	left := r.indent(item.Indent() + 1)
	if len(inline) > 1 {
		for _, c := range inline {
			if _, err := fmt.Fprintf(w, "%s# %s\n", left, c); err != nil {
				return err
			}
		}
	}

	if len(inline) > 1 || len(v) > 1 {
		if !blockWritten && block != "" {
			if _, err := fmt.Fprintf(w, "%s%s\n", left, block); err != nil {
				return err
			}
		}
		v, _ := formatValue(item.Value(), r.maxLineLength(len(left)))
		for _, line := range v {
			if _, err := fmt.Fprintf(w, "%s%s\n", left, line); err != nil {
				return err
			}
		}
	}

	return nil
}

// maxLineLength calculates the maximum line length for wrapping based on the
// MaxLineLength and the prefix length (e.g., key and indent).  If MaxLineLength
// is less than or equal to zero, no wrapping is done and the function returns 0.
// If the prefix takes up more than the MaxLineLength, it will be adjusted to
// ensure the maximum line length is generally sane.  This is done by multiplying
// the MaxLineLength by until it is greater than the prefix length.
func (r *Renderer) maxLineLength(prefixLen int) int {
	if r.MaxLineLength <= 0 {
		return 0 // No wrapping
	}

	maxLen := r.MaxLineLength - prefixLen
	for i := 2; maxLen < 1; i++ {
		maxLen = r.MaxLineLength*i - prefixLen
	}

	return maxLen
}

// indent returns a string of spaces for indentation based on the
// SpacesPerIndent setting. If SpacesPerIndent is zero, it defaults to 2 spaces.
// If the indent level is less than or equal to zero, it returns an empty string.
func (r *Renderer) indent(i int) string {
	if i <= 0 {
		return ""
	}

	spaces := r.SpacesPerIndent
	if spaces <= 0 {
		spaces = 2 // Default to 2 spaces if not set
	}
	return strings.Repeat(" ", i*spaces)
}

// isKeyword checks if a string is a potential YAML keyword.
func isKeyword(s string) bool {
	switch strings.ToLower(s) {
	case "null",
		"true", "false",
		"~",
		".inf", "-.inf",
		".nan",
		"yes", "no",
		"on", "off":
		return true
	}
	return false
}

// hasSpecialChars checks if a string contains special characters that would
// require it to be quoted in YAML. It allows newlines and tabs, to not be
// qualified as special characters but checks for control characters, certain
// punctuation, and specific sequences that are problematic in YAML.
func hasSpecialChars(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\t' {
			continue // Allow newlines and tabs
		}
		if unicode.IsControl(r) && r < 128 {
			return true
		}
		if strings.ContainsRune("#&*!|'\"%@`", r) {
			return true
		}
		if strings.Contains(s, ": ") {
			return true
		}
		if strings.Contains(s, string('\\')) {
			return true
		}
	}
	return false
}

// hasSpecialSpaces checks if a string has leading or trailing spaces that would
// require it to be quoted in YAML.  Newlines are excluded from this check,
func hasSpecialSpaces(s string) bool {
	var first, last bool

	if len(s) > 0 {
		c0 := rune(s[0])
		clast := rune(s[len(s)-1])
		first = (c0 != '\n') && unicode.IsSpace(c0)
		last = (clast != '\n') && unicode.IsSpace(clast)
	}
	return first || last
}

// chunkString splits a string into chunks of a specified length, adding a
// backslash at the end of each chunk except the last one. If n is less than or
// equal to zero, it returns the original string in a slice.  The string passed
// in needs to be a double-quoted string.
func chunkString(s string, n int) []string {
	s = "\"" + s + "\""
	if n <= 0 {
		return []string{s}
	}

	lines := make([]string, 0, len(s)/n+1)

	var line string
	for _, r := range s {
		line += string(r)
		if len(line)+1 < n {
			continue
		}
		line += "\\"

		lines = append(lines, line)
		line = ""
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
}

// formatValue formats a value for YAML output, handling special cases like
// keywords, newlines, spaces, and special characters. It returns a slice of
// strings representing the formatted value, and a string indicating the style
// of the value (e.g., block style, quoted style, etc.).
func formatValue(v *string, maxLineLength int) ([]string, string) {
	if v == nil || *v == "" {
		return []string{"''"}, ""
	}
	val := *v

	if isKeyword(val) {
		return []string{"'" + val + "'"}, ""
	}

	//specials := hasSpecialChars(val)
	newlines := strings.Contains(val, "\n")
	spaces := hasSpecialSpaces(val)

	if spaces {
		return chunkString(val, maxLineLength), ""
	}

	if newlines {
		lines := strings.Split(val, "\n")
		style := "|-"
		if strings.HasSuffix(val, "\n") {
			style = "|+"
			// Remove the last empty line
			lines = lines[:len(lines)-1]
		}

		return lines, style
	}

	if hasSpecialChars(val) {
		return chunkString(val, maxLineLength), ""
	}

	if maxLineLength > 0 && len(val) > maxLineLength {
		return chunkString(val, maxLineLength), ""
	}

	return []string{val}, ""
}
