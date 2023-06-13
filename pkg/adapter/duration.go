// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"time"

	"github.com/goschtalt/approx"
	"github.com/goschtalt/goschtalt"
)

// DurationUnmarshal converts a string to a time.Duration or *time.Duration if
// possible, or returns an error indicating the failure.
func DurationUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalDuration{}, "DurationUnmarshal")
}

// MarshalDuration converts a time.Duration into its configuration form.  The
// configuration form is a string.
func MarshalDuration() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalDuration{}, "MarshalDuration")
}

type marshalDuration struct{}

func (marshalDuration) From(from, to reflect.Value) (any, error) {
	if from.Kind() != reflect.String {
		return nil, goschtalt.ErrNotApplicable
	}

	ptr := false
	switch to.Type() {
	case reflect.TypeOf(time.Duration(1)):
	case reflect.TypeOf(new(time.Duration)):
		ptr = true
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	d, err := approx.ParseDuration(from.Interface().(string))
	if err != nil {
		return nil, err
	}

	if ptr {
		return &d, nil
	}
	return d, nil
}

func (marshalDuration) To(from reflect.Value) (any, error) {
	if from.Type() == reflect.TypeOf(time.Duration(1)) {
		return approx.String(from.Interface().(time.Duration)), nil
	}

	return nil, goschtalt.ErrNotApplicable
}
