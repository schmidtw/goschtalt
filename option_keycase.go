// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "strings"

const (
	Unchanged = iota + 1
	Lower
	Upper
)

type caseChanger func(string) string

// TODO
// WithOriginalCaseKeys does not alter the map keys found in the configuration
// files.
func WithKeyCase(mode int) Option {
	var fn func(*map[string]any)

	switch mode {
	case Unchanged:
		fn = func(m *map[string]any) {}
	case Lower:
		fn = func(m *map[string]any) {
			keycaseMap(strings.ToLower, *m)
		}
	case Upper:
		fn = func(m *map[string]any) {
			keycaseMap(strings.ToUpper, *m)
		}
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
