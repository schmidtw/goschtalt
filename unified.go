// SPDX-FileCopyrightText: 2025 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/goschtalt/goschtalt/internal/encoding"
	"github.com/goschtalt/goschtalt/pkg/doc"
	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/k0kubun/pp"
)

type unified struct {
	doc    *doc.Object
	key    *string
	value  any
	preset any
	indent int

	array bool

	children map[string]unified
}

var _ encoding.Encodeable = &unified{}

func (u *unified) Indent() int {
	return u.indent
}

func (u *unified) Headers() []string {
	var rv []string
	var details strings.Builder
	var comma string

	if u.doc != nil {
		rv = strings.Split(u.doc.Doc, "\n")

		if u.doc.Deprecated {
			details.WriteString("!!! DEPRECATED !!!")
		}

		fmt.Fprintf(&details, "type: %s", u.doc.TypeString())
		comma = ", "
		fmt.Fprintf(&details, "\n")
	}

	if u.preset != nil {
		details.WriteString(comma)
		fmt.Fprintf(&details, "default: %s", toString(u.preset))
	}

	if details.Len() > 0 {
		rv = append(rv, details.String())
	}

	return rv
}

func (u *unified) Inline() []string {
	return nil
}

func (u *unified) Key() *string {
	return u.key
}

func (u *unified) Children() encoding.Encodeables {
	if len(u.children) == 0 {
		return nil
	}
	rv := make(encoding.Encodeables, 0, len(u.children))
	for key := range u.children {
		child := u.children[key]
		rv = append(rv, &child)
	}
	sort.Sort(rv)
	return rv
}

func (u *unified) Value() *string {
	if u.value == nil {
		return nil
	}

	s := toString(u.value)
	return &s
}

func toString(value any) string {
	s := "%v"
	switch v := value.(type) {
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case int, int8, int16, int32, int64:
		s = "%d"
	case uint, uint8, uint16, uint32, uint64:
		s = "%ud"
	case float32, float64:
		s = "%f"
	default:
	}
	return fmt.Sprintf(s, value)
}

func (u unified) childrenKeys() []string {
	keys := make([]string, 0, len(u.children))
	for key := range u.children {
		keys = append(keys, key)
	}

	// Perform a natural sort: numeric if possible, else string
	sort.Slice(keys, func(i, j int) bool {
		return naturalLess(keys[i], keys[j])
	})

	return keys
}

// naturalLess compares two strings, attempting to parse them as
// numbers (int/float) for comparison; if parsing fails, it falls back
// to standard string comparison.
func naturalLess(a, b string) bool {
	aNum, aErr := strconv.ParseFloat(a, 64)
	bNum, bErr := strconv.ParseFloat(b, 64)

	// If both are valid numbers, compare numerically.
	if aErr == nil && bErr == nil {
		if aNum == bNum {
			// If numeric values are the same, compare strings to break ties
			return a < b
		}
		return aNum < bNum
	}

	// If only one is numeric, that one goes first.
	if aErr == nil && bErr != nil {
		return true
	}
	if aErr != nil && bErr == nil {
		return false
	}

	// Otherwise, do a normal string compare.
	return a < b
}

/*
// escapeNakedQuotes scans through s and inserts a backslash before
// any quote that isn’t already preceded by a backslash.
func escapeNakedQuotes(s string) string {
	var sb strings.Builder
	runes := []rune(s)

	for i := range runes {
		cur := runes[i]

		// If we hit a quote and the previous rune is NOT a backslash, escape it.
		if cur == '"' {
			if i == 0 || runes[i-1] != '\\' {
				sb.WriteString("\\\"")
				continue
			}
		}

		sb.WriteRune(cur)
	}

	return sb.String()
}

func (u *unified) toYML(w io.Writer, prefix []string, indent int, arrayItem bool) error {
	ind := strings.Repeat("  ", indent)

	val := escapeNakedQuotes(u.Value())

	if !arrayItem {
		// Write comment if present
		if u.doc != nil {
			if u.doc.Doc != "" {
				fmt.Fprintf(w, "\n")
				for _, line := range strings.Split(u.doc.Doc, "\n") {
					fmt.Fprintf(w, "%s# %s\n", ind, line)
				}
			}
			fmt.Fprintf(w, "%s# type: %s\n", ind, u.doc.TypeString())
		}
	}
	if len(u.children) > 0 {
		if arrayItem {
			fmt.Fprintf(w, "%s-\n", ind)
		} else {
			fmt.Fprintf(w, "%s%s:\n", ind, prefix[len(prefix)-1])
		}
		if u.array {
			childDoc := u.firstChildDoc()
			if childDoc != nil && childDoc.Doc != "" {
				ind := strings.Repeat("  ", indent+1)
				for _, line := range strings.Split(childDoc.Doc, "\n") {
					fmt.Fprintf(w, "%s# %s\n", ind, line)
				}
			}
		}
	} else {
		if arrayItem {
			fmt.Fprintf(w, "%s- %q\n", ind, val)
		} else {
			fmt.Fprintf(w, "%s%s: %q\n", ind, prefix[len(prefix)-1], val)
		}
	}

	indent++

	for _, key := range u.childrenKeys() {
		child := u.children[key]
		err := child.toYML(w, append(prefix, key), indent, u.array)
		if err != nil {
			return err
		}
	}
	return nil
}
*/

func calcUnified(d *doc.Object, compiled *meta.Object) (unified, error) {
	return calcUnifiedInt(nil, -1, d, compiled)
}

func calcUnifiedInt(name *string, indent int, d *doc.Object, compiled *meta.Object) (unified, error) {
	pName := "<nil>"
	if name != nil {
		pName = *name
	}
	fmt.Printf("calcUnifiedInt: name=%s, indent=%d\n", pName, indent)

	u := unified{
		indent: indent,
	}

	if nil != name {
		u.key = name
	}

	if d != nil {
		// For now since this prevents the printing out of a large document
		//tmp := d.ShallowCopy()
		//u.doc = &tmp
		u.doc = d
	}

	var docArray bool
	arrayLen := 0
	mapLen := 0
	if d != nil {
		if d.Type == doc.TYPE_ARRAY {
			docArray = true
		} else {
			mapLen = len(d.Children)
		}
	}
	if compiled != nil {
		arrayLen = max(arrayLen, len(compiled.Array))
		mapLen = max(mapLen, len(compiled.Map))
	}

	totalLen := arrayLen + mapLen

	// This is a leaf node.
	if totalLen == 0 {
		u.value = compiled.Value
		return u, nil
	}

	// Check to see if we have a conflicting definition.
	if (docArray && mapLen > 0) || (mapLen > 0 && arrayLen > 0) {
		return u, errors.New("conflicting definitions: array and map cannot coexist in the same object")
	}

	// Handle arrays first.
	if docArray || arrayLen > 0 {
		u.array = true
		nextDoc := d
		if nextDoc != nil {
			if tmp, ok := nextDoc.Children[doc.NAME_ARRAY]; ok {
				nextDoc = &tmp
			}
		}

		u.children = make(map[string]unified, arrayLen)
		for i, next := range compiled.Array {
			got, err := calcUnifiedInt(nil, indent+1, nextDoc, &next)
			// only output the docs for the first entry in the array.
			nextDoc = nil
			if err != nil {
				return unified{}, err
			}
			pp.Println("inserting array entry:", i, "->", got)
			u.children[strconv.Itoa(i)] = got
		}

		return u, nil
	}

	// Handle maps next.
	names := make(map[string]bool, mapLen)
	if d != nil {
		for key := range d.Children {
			names[key] = false
		}
	}
	if compiled != nil {
		for key := range compiled.Map {
			names[key] = false
		}
	}
	u.children = make(map[string]unified, mapLen)

	for key := range names {
		fmt.Println("working on: ", key)
		switch key {
		case doc.NAME_ARRAY:
			return unified{}, errors.New("array key cannot be used in a map object")
		case doc.NAME_EMBEDDED:
			return unified{}, errors.New("embedded key cannot be used in a map object")
		case doc.NAME_MAP_KEY:
			// TBD
		default:
			nextDoc := d
			if nextDoc != nil {
				if tmp, ok := nextDoc.Children[key]; ok {
					nextDoc = &tmp
				} else {
					nextDoc = nil // no doc for this key
				}
			}

			nextCompiled := compiled
			if nextCompiled != nil {
				if val, ok := nextCompiled.Map[key]; ok {
					nextCompiled = &val
				} else {
					nextCompiled = nil // no compiled for this key
				}
			}

			fmt.Println("recursing")
			got, err := calcUnifiedInt(&key, indent+1, nextDoc, nextCompiled)
			if err != nil {
				return unified{}, err
			}
			fmt.Printf("inserting: %s\n", key)
			u.children[key] = got
		}
	}

	return u, nil
}
