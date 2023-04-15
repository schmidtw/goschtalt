// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
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
	return goschtalt.AdaptFromCfg(stringToTime(layout), "TimeUnmarshal")
}

func stringToTime(layout string) func(reflect.Value, reflect.Value) (any, error) {
	return func(from, to reflect.Value) (any, error) {
		if from.Kind() == reflect.String && to.Type() == reflect.TypeOf(time.Time{}) {
			a, e := time.Parse(layout, from.Interface().(string))
			return a, e
		}

		return nil, goschtalt.ErrNotApplicable
	}
}

// MarshalTime converts a time.Time into its configuration form. The
// configuration form is a string matching the specified layout.
func MarshalTime(layout string) goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(timeToCfg(layout), "MarshalTime")
}

func timeToCfg(layout string) func(reflect.Value) (any, error) {
	return func(from reflect.Value) (any, error) {
		if from.Type() == reflect.TypeOf(time.Time{}) {
			a := from.Interface().(time.Time).Format(layout)
			return a, nil
		}

		return nil, goschtalt.ErrNotApplicable
	}
}
