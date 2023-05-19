// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"

	"github.com/goschtalt/goschtalt"
)

// FloatUnmarshal converts a string to a float32/float64 or *float32/*float64 if
// possible, or returns an error indicating the failure.
func FloatUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(marshalFloat{}, "FloatUnmarshal")
}

// MarshalFloat converts a float32/float64 into its string configuration form.
func MarshalFloat() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(marshalFloat{}, "MarshalFloat")
}

type marshalFloat struct{}

func (marshalFloat) From(from, to reflect.Value) (any, error) {
	if from.Kind() != reflect.String {
		return nil, goschtalt.ErrNotApplicable
	}

	ptr := false
	bits := 32
	switch to.Type() {
	case reflect.TypeOf(float32(0.0)):
	case reflect.TypeOf(new(float32)):
		ptr = true
	case reflect.TypeOf(float64(0.0)):
		bits = 64
	case reflect.TypeOf(new(float64)):
		bits = 64
		ptr = true
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	rv64, err := strconv.ParseFloat(from.Interface().(string), bits)
	if err != nil {
		return nil, goschtalt.ErrNotApplicable
	}

	if bits == 64 {
		if ptr {
			return &rv64, nil
		}
		return rv64, nil
	}

	rv32 := float32(rv64)
	if ptr {
		return &rv32, nil
	}
	return rv32, nil
}

func (marshalFloat) To(from reflect.Value) (any, error) {
	var f float64
	bits := 32

	switch from.Type() {
	case reflect.TypeOf(float32(0.0)):
		f = float64(from.Interface().(float32))
	case reflect.TypeOf(float64(0.0)):
		f = from.Interface().(float64)
		bits = 64
	case reflect.TypeOf(new(float32)):
		f = float64(*(from.Interface().(*float32)))
	case reflect.TypeOf(new(float64)):
		f = *(from.Interface().(*float64))
		bits = 64
	default:
		return nil, goschtalt.ErrNotApplicable
	}

	return strconv.FormatFloat(f, 'f', -1, bits), nil
}
