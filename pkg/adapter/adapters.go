// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"encoding"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/goschtalt/goschtalt"
)

// StringToDuration converts a string to a time.Duration or *time.Duration if
// possible, or returns an error indicating the failure.
func StringToDuration() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(stringToDuration, "StringToDuration")
}

func stringToDuration(from, to reflect.Value) (any, error) {
	if from.Kind() == reflect.String && to.Type() == reflect.TypeOf(time.Duration(1)) {
		return time.ParseDuration(from.Interface().(string))
	}

	return nil, goschtalt.ErrNotApplicable
}

// DurationToCfg converts a time.Duration into its configuration form.  The
// configuration form is a string.
func DurationToCfg() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(durationToCfg, "DurationToCfg")
}

func durationToCfg(from reflect.Value) (any, error) {
	if from.Type() == reflect.TypeOf(time.Duration(1)) {
		return from.Interface().(time.Duration).String(), nil
	}

	return nil, goschtalt.ErrNotApplicable
}

// StringToIP converts a string to a net.IP if possible, or returns an
// error indicating the failure.
func StringToIP() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(stringToIP, "StringToIP")
}

func stringToIP(from, to reflect.Value) (any, error) {
	if from.Kind() == reflect.String && to.Type() == reflect.TypeOf(net.IP{}) {
		ip := net.ParseIP(from.Interface().(string))
		if ip == nil {
			return nil, fmt.Errorf("failed parsing ip %v", from)
		}
		return ip, nil
	}

	return nil, goschtalt.ErrNotApplicable
}

// IPToCfg converts a net.IP into its configuration form.  The
// configuration form is a string.
func IPToCfg() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(ipToCfg, "IPToCfg")
}

func ipToCfg(from reflect.Value) (any, error) {
	if from.Type() == reflect.TypeOf(net.IP{}) {
		return from.Interface().(net.IP).String(), nil
	}

	return nil, goschtalt.ErrNotApplicable
}

// StringToTime converts a string to a time.Time if possible, or returns an
// error indicating the failure.  The specified layout is used as the string
// form.
func StringToTime(layout string) goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(stringToTime(layout), "StringToTime")
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

// TimeToCfg converts a time.Time into its configuration form. The
// configuration form is a string matching the specified layout.
func TimeToCfg(layout string) goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(timeToCfg(layout), "TimeToCfg")
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

type all struct{}

// AnyTextUnmarshal uses the TextUnmarshaler() method if present for an
// object for all types of objects.  The only configuration value type allowed
// is a string.
func AnyTextUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(withTextUnmarshal(all{}), "AnyTextUnmarshal")
}

// LimitedTextUnmarshal uses the TextUnmarshaler() method if present for
// an object that matches the type of object provided as argument t.  The only
// configuration value type allowed is a string.
func LimitedTextUnmarshal(t any) goschtalt.UnmarshalOption {
	s := "LimitedTextUnmarshal[" + reflect.TypeOf(t).String() + "]"
	return goschtalt.AdaptFromCfg(withTextUnmarshal(t), s)
}

func withTextUnmarshal(t any) func(reflect.Value, reflect.Value) (any, error) {
	return func(from, to reflect.Value) (any, error) {
		if from.Kind() != reflect.String {
			return nil, goschtalt.ErrNotApplicable
		}

		target := reflect.TypeOf(t)
		if target != to.Type() && target != reflect.TypeOf(all{}) {
			return nil, goschtalt.ErrNotApplicable
		}

		// Always make a writable pointer to the object version
		result := reflect.New(reflect.Indirect(to).Type()).Interface()

		u, ok := result.(encoding.TextUnmarshaler)
		if !ok {
			return nil, goschtalt.ErrNotApplicable
		}

		if err := u.UnmarshalText([]byte(from.Interface().(string))); err != nil {
			return nil, err
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

// AnyMarshalText inspects the source object to see if it implements
// an encoding.TextMarshaler interface.  If it does, the object is marshaled
// using the MarshalText() method.  An error in marshaling will generate an
// error.
func AnyMarshalText() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(withMarshalText(all{}), "AnyMarshalText")
}

// LimitedMarshalText uses the TextUnmarshaler() method if present for
// an object that matches the type of object provided as argument t.  The only
// configuration value type allowed is a string.
func LimitedMarshalText(t any) goschtalt.ValueOption {
	s := "LimitedMarshalText[" + reflect.TypeOf(t).String() + "]"
	return goschtalt.AdaptToCfg(withMarshalText(t), s)
}

func withMarshalText(t any) func(reflect.Value) (any, error) {
	return func(from reflect.Value) (any, error) {
		start := reflect.TypeOf(t)
		if start != from.Type() && start != reflect.TypeOf(all{}) {
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
