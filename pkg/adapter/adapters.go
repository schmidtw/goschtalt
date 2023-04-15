// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"encoding"
	"reflect"

	"github.com/goschtalt/goschtalt"
)

// Matcher provides a simple way to include or exclude types that are converted
// via TextUnmarshal() or MarshalText().
type Matcher func(t any) bool

// All matches all the possible types.
func All(t any) bool {
	return true
}

// TextUnmarshal uses the TextUnmarshaler() method if present for an object that
// the matcher function allows.  The only configuration value type allowed is a
// string.
func TextUnmarshal(m Matcher) goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(withTextUnmarshal(m), "TextUnmarshal")
}

func withTextUnmarshal(m Matcher) func(reflect.Value, reflect.Value) (any, error) {
	return func(from, to reflect.Value) (any, error) {
		if from.Kind() != reflect.String || !m(reflect.TypeOf(to)) {
			return nil, goschtalt.ErrNotApplicable
		}

		// Always make a writable pointer to the object version
		result := reflect.New(reflect.Indirect(to).Type()).Interface()

		u, ok := result.(encoding.TextUnmarshaler)
		if !ok {
			return nil, goschtalt.ErrNotApplicable
		}

		if err := u.UnmarshalText([]byte(from.Interface().(string))); err != nil {
			return nil, goschtalt.ErrNotApplicable
		}

		// Return the same type passed in.  Since we may have been passed a
		// non-pointer based value, dereference the point before return it in
		// those cases.
		if to.Type() != reflect.TypeOf(result) {
			result = reflect.Indirect(reflect.ValueOf(result)).Interface()
		}
		return result, nil
	}
}

// MarshalText uses the TextUnmarshaler() method if present for an object if the
// matcher function allows.  The only configuration value type allowed is a string.
func MarshalText(m Matcher) goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(withMarshalText(m), "MarshalText")
}

func withMarshalText(m Matcher) func(reflect.Value) (any, error) {
	return func(from reflect.Value) (any, error) {
		if !m(from.Type()) {
			return nil, goschtalt.ErrNotApplicable
		}

		if from.Kind() != reflect.Pointer {
			tmp := reflect.New(from.Type())
			tmp.Elem().Set(from)
			from = tmp
		}

		m, ok := from.Interface().(encoding.TextMarshaler)
		if !ok {
			return nil, goschtalt.ErrNotApplicable
		}

		b, err := m.MarshalText()
		if err != nil {
			return nil, err
		}

		return string(b), nil
	}
}
