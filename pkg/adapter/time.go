// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/goschtalt/goschtalt"
)

// AllButTime matches all types except the time.Time and *time.Time types.  By
// excluding these types, a custom TimeUnmarshal() and MarshalTime() can be
// set and used reliably.
func AllButTime(t any) bool {
	if t == reflect.TypeOf(time.Time{}) {
		return false
	}
	return All(t)
}

// TimeUnmarshal converts a string to a time.Time if possible, or returns an
// error indicating the failure.  The specified layout is used as the string
// form.
func TimeUnmarshal(layout string) goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalTime{layout: layout},
		fmt.Sprintf("TimeUnmarshal['%s']", layout))
}

// MarshalTime converts a time.Time into its configuration form. The
// configuration form is a string matching the specified layout.
func MarshalTime(layout string) goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalTime{layout},
		fmt.Sprintf("MarshalTime['%s']", layout))
}

type marshalTime struct {
	layout string
}

func (t marshalTime) From(from, to reflect.Value) (any, error) {
	if to.Type() != reflect.TypeOf(time.Time{}) {
		return nil, goschtalt.ErrNotApplicable
	}

	var sec, nsec int64
	switch from.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sec = from.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		got := from.Uint()
		if math.MaxInt64 < got {
			return nil, goschtalt.ErrNotApplicable
		}
		sec = int64(got) // nolint:gosec
	case reflect.Float32, reflect.Float64:
		got := from.Float()
		sec = int64(got)
		nsec = int64((got - float64(sec)) * 1e9)
	case reflect.String:
		pt, err := time.Parse(t.layout, from.Interface().(string))
		if err != nil {
			return nil, err
		}

		return pt.UTC(), nil
	}

	return time.Unix(sec, nsec).UTC(), nil
}

func (t marshalTime) To(from reflect.Value) (any, error) {
	if from.Type() == reflect.TypeOf(time.Time{}) {
		a := from.Interface().(time.Time).Format(t.layout)
		return a, nil
	}

	return nil, goschtalt.ErrNotApplicable
}
