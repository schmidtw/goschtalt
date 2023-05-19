// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"

	"github.com/goschtalt/goschtalt"
)

// BoolUnmarshal converts a string to a bool or *bool if possible, or returns an
// error indicating the failure.  The case is ignored, but only the following
// values are accepted: 'f', 'false', 't', 'true'
func BoolUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalBool{}, "BoolUnmarshal")
}

// MarshalBool converts a bool into its configuration form.  The
// configuration form is a string of value 'true' or 'false'.
func MarshalBool() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalBool{}, "MarshalBool")
}

type marshalBool struct{}

func (marshalBool) From(from, to reflect.Value) (any, error) {
	if from.Kind() != reflect.String {
		return nil, goschtalt.ErrNotApplicable
	}

	ptr := false
	switch to.Type() {
	case reflect.TypeOf(bool(true)):
	case reflect.TypeOf(new(bool)):
		ptr = true
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	rv, err := strconv.ParseBool(from.Interface().(string))
	if err != nil {
		return nil, goschtalt.ErrNotApplicable
	}

	if ptr {
		return &rv, nil
	}
	return rv, nil
}

func (marshalBool) To(from reflect.Value) (any, error) {
	var b bool

	switch from.Type() {
	case reflect.TypeOf(bool(true)):
		b = from.Bool()
	case reflect.TypeOf(new(bool)):
		b = *from.Interface().(*bool)
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	return strconv.FormatBool(b), nil
}
