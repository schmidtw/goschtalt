// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "strings"

// KeyCase mode values
const (
	Unchanged = iota + 1 // The case of the key is not changed.
	Lower                // The key is made all lowercase.
	Upper                // The key is made all lowercase.
)

type caseChanger func(string) string

// KeyCase specifies the alteration to the case of the keys.
func KeyCase(mode int) Option {
	var fn caseChanger

	switch mode {
	case Unchanged:
		fn = func(s string) string {
			return s
		}
	case Lower:
		fn = strings.ToLower
	case Upper:
		fn = strings.ToUpper
	default:
	}

	return func(g *Goschtalt) error {
		if fn == nil {
			return ErrInvalidOption
		}
		g.keySwizzler = fn
		return nil
	}
}

func keycaseMap(to caseChanger, m map[string]any) {
	for key, val := range m {
		lower := to(key)
		delete(m, key)
		m[lower] = keycaseVal(to, val)
	}
}

func keycaseVal(to caseChanger, val any) any {
	switch val := val.(type) {
	case map[string]any:
		keycaseMap(to, val)
	case []any:
		keycaseArray(to, val)
	}
	return val
}

func keycaseArray(to caseChanger, a []any) {
	for i, val := range a {
		a[i] = keycaseVal(to, val)
	}
}
