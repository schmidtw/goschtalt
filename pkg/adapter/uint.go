// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"

	"github.com/goschtalt/goschtalt"
)

// UintUnmarshal converts a string to a float32/float64 or *float32/*float64 if
// possible, or returns an error indicating the failure.
func UintUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalUint{}, "UintUnmarshal")
}

// MarshalUint converts a float32/float64 into its string configuration form.
func MarshalUint() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalUint{}, "MarshalUint")
}

type marshalUint struct{}

const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

func (marshalUint) From(from, to reflect.Value) (any, error) { //nolint:funlen
	if from.Kind() != reflect.String {
		return nil, goschtalt.ErrNotApplicable
	}

	var (
		ptr      bool
		uintType bool
		bits     int
	)

	switch to.Type() {
	case reflect.TypeOf(uint(0)):
		bits = uintSize
		uintType = true
	case reflect.TypeOf(uint8(0)):
		bits = 8
	case reflect.TypeOf(uint16(0)):
		bits = 16
	case reflect.TypeOf(uint32(0)):
		bits = 32
	case reflect.TypeOf(uint64(0)):
		bits = 64
	case reflect.TypeOf(new(uint)):
		ptr = true
		bits = uintSize
		uintType = true
	case reflect.TypeOf(new(uint8)):
		ptr = true
		bits = 8
	case reflect.TypeOf(new(uint16)):
		ptr = true
		bits = 16
	case reflect.TypeOf(new(uint32)):
		ptr = true
		bits = 32
	case reflect.TypeOf(new(uint64)):
		ptr = true
		bits = 64
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	num, err := strconv.ParseUint(from.Interface().(string), 0, bits)
	if err != nil && err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
		return nil, goschtalt.ErrNotApplicable
	}
	if err != nil {
		return nil, err
	}

	if bits == 8 {
		rv := uint8(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if bits == 16 {
		rv := uint16(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if uintType {
		rv := uint(num)
		if ptr {
			return &rv, nil
		}
		return rv, nil
	}

	if bits == 32 {
		rv := uint32(num)
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

func (m marshalUint) To(from reflect.Value) (any, error) {
	if from.CanUint() {
		return strconv.FormatUint(from.Uint(), 10), nil
	}

	if from.Kind() != reflect.Pointer {
		return nil, goschtalt.ErrNotApplicable
	}

	from = from.Elem()

	return m.To(from)
}
