// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"

	"github.com/goschtalt/goschtalt"
)

// IntUnmarshal converts a string to a float32/float64 or *float32/*float64 if
// possible, or returns an error indicating the failure.
func IntUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalInt{}, "IntUnmarshal")
}

// MarshalInt converts a float32/float64 into its string configuration form.
func MarshalInt() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalInt{}, "MarshalInt")
}

type marshalInt struct{}

const intSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

func (marshalInt) From(from, to reflect.Value) (any, error) { //nolint:funlen
	if from.Kind() != reflect.String {
		return nil, goschtalt.ErrNotApplicable
	}

	var (
		ptr     bool
		intType bool
		bits    int
	)

	switch to.Type() {
	case reflect.TypeOf(int(0)):
		bits = intSize
		intType = true
	case reflect.TypeOf(int8(0)):
		bits = 8
	case reflect.TypeOf(int16(0)):
		bits = 16
	case reflect.TypeOf(int32(0)):
		bits = 32
	case reflect.TypeOf(int64(0)):
		bits = 64
	case reflect.TypeOf(new(int)):
		ptr = true
		bits = intSize
		intType = true
	case reflect.TypeOf(new(int8)):
		ptr = true
		bits = 8
	case reflect.TypeOf(new(int16)):
		ptr = true
		bits = 16
	case reflect.TypeOf(new(int32)):
		ptr = true
		bits = 32
	case reflect.TypeOf(new(int64)):
		ptr = true
		bits = 64
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	num, err := strconv.ParseInt(from.Interface().(string), 0, bits)
	if err != nil && err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
		return nil, goschtalt.ErrNotApplicable
	}
	if err != nil {
		return nil, err
	}

	if bits == 8 {
		rv := int8(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if bits == 16 {
		rv := int16(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if intType {
		rv := int(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if bits == 32 {
		rv := int32(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if ptr {
		return &num, nil
	}
	return num, nil
}

func (m marshalInt) To(from reflect.Value) (any, error) {
	if from.CanInt() {
		return strconv.FormatInt(from.Int(), 10), nil
	}

	if from.Kind() != reflect.Pointer {
		return nil, goschtalt.ErrNotApplicable
	}

	from = from.Elem()

	return m.To(from)
}
